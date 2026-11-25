package service

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"ftools/entity"

	"go.uber.org/zap"
)

// ghostscriptBin is the Ghostscript executable
const ghostscriptBin = "gs"

// ghostscriptWaitDelay is how long to wait for gs to exit after its context is
// done before its pipes are closed, so a hung process can't block the handler.
const ghostscriptWaitDelay = 5 * time.Second

// ghostscriptOutputLimit caps how much of gs's output is kept for the logs.
const ghostscriptOutputLimit = 4 << 10

var pdfPresets = map[entity.PDFLevel]string{
	entity.PDFLevelLow:    "/printer",
	entity.PDFLevelMedium: "/ebook",
	entity.PDFLevelHigh:   "/screen",
}

// ghostscriptArgs builds the gs command line.
func ghostscriptArgs(req entity.PDFCompressRequest, in, out string) ([]string, error) {
	preset, ok := pdfPresets[req.Level]
	if !ok {
		return nil, fmt.Errorf("%w: invalid compression level %q", entity.ErrInvalidRequest, req.Level)
	}
	dpi := strconv.Itoa(req.DPI)

	args := []string{
		"-sDEVICE=pdfwrite",
		"-dCompatibilityLevel=1.5",
		"-dPDFSETTINGS=" + preset,
		"-dSAFER",
		"-dBATCH",
		"-dNOPAUSE",
		"-dQUIET",
		// Without this, gs exits 0 on password-protected or broken files and
		// writes an empty PDF.
		"-dPDFSTOPONERROR",
		"-dDetectDuplicateImages=true",
		// Mono (1-bit) images stay on the preset so scanned text stays readable.
		"-dDownsampleColorImages=true",
		"-dColorImageResolution=" + dpi,
		"-dDownsampleGrayImages=true",
		"-dGrayImageResolution=" + dpi,
	}

	switch req.Color {
	case entity.PDFColorOriginal:
	case entity.PDFColorGrayscale:
		args = append(args, "-sColorConversionStrategy=Gray", "-dProcessColorModel=/DeviceGray")
	default:
		return nil, fmt.Errorf("%w: invalid color option %q", entity.ErrInvalidRequest, req.Color)
	}

	return append(args, "-sOutputFile="+out, in), nil
}

func (s *service) CompressPDF(ctx context.Context, req entity.PDFCompressRequest) (entity.PDFCompressResult, error) {
	release, err := s.acquirePDFSlot(ctx)
	if err != nil {
		return entity.PDFCompressResult{}, err
	}
	defer release()

	in, err := writeTemp("ftools-in-*.pdf", req.File)
	if err != nil {
		s.logger.Error("failed to write pdf input", zap.Error(err))
		return entity.PDFCompressResult{}, entity.ErrInternalError
	}
	defer os.Remove(in)

	out, err := writeTemp("ftools-out-*.pdf", nil)
	if err != nil {
		s.logger.Error("failed to create pdf output", zap.Error(err))
		return entity.PDFCompressResult{}, entity.ErrInternalError
	}
	defer os.Remove(out)

	args, err := ghostscriptArgs(req, in, out)
	if err != nil {
		return entity.PDFCompressResult{}, err
	}

	if err := s.runGhostscript(ctx, args); err != nil {
		return entity.PDFCompressResult{}, err
	}

	compressed, err := os.ReadFile(out)
	if err != nil {
		s.logger.Error("failed to read pdf output", zap.Error(err))
		return entity.PDFCompressResult{}, entity.ErrInternalError
	}
	if len(compressed) == 0 {
		s.logger.Error("ghostscript produced an empty pdf")
		return entity.PDFCompressResult{}, entity.ErrInternalError
	}

	result := entity.PDFCompressResult{
		PDF:          compressed,
		OriginalSize: len(req.File),
		ResultSize:   len(compressed),
	}
	if len(compressed) >= len(req.File) {
		result.PDF = req.File
		result.ResultSize = len(req.File)
		result.AlreadyOptimized = true
	}
	return result, nil
}

// acquirePDFSlot waits for a free Ghostscript slot. It gives up with
// ErrBusy after the queue timeout or when the request is canceled.
func (s *service) acquirePDFSlot(ctx context.Context) (release func(), err error) {
	wait, cancel := context.WithTimeout(ctx, s.pdf.QueueTimeout)
	defer cancel()

	select {
	case s.pdfSlots <- struct{}{}:
		return func() { <-s.pdfSlots }, nil
	case <-wait.Done():
		return nil, entity.ErrResourceExhausted
	}
}

func (s *service) runGhostscript(ctx context.Context, args []string) error {
	ctx, cancel := context.WithTimeout(ctx, s.pdf.Timeout)
	defer cancel()

	output := &cappedBuffer{limit: ghostscriptOutputLimit}
	cmd := exec.CommandContext(ctx, ghostscriptBin, args...)
	cmd.Stdout = output
	cmd.Stderr = output
	cmd.WaitDelay = ghostscriptWaitDelay

	start := time.Now()
	err := cmd.Run()
	if err == nil {
		s.logger.Info("ghostscript finished", zap.Duration("duration", time.Since(start)))
		return nil
	}

	fields := []zap.Field{
		zap.Error(err),
		zap.Duration("duration", time.Since(start)),
		zap.String("output", output.String()),
	}

	switch {
	case errors.Is(ctx.Err(), context.DeadlineExceeded):
		s.logger.Warn("ghostscript timed out", fields...)
		return fmt.Errorf("%w: this PDF took too long to process", entity.ErrDeadlineExceeded)
	case ctx.Err() != nil:
		// The client went away; nobody reads the response.
		return fmt.Errorf("%w: client canceled the request: %w", entity.ErrDeadlineExceeded, ctx.Err())
	}

	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		s.logger.Error("ghostscript failed", fields...)
		return fmt.Errorf("%w: couldn't read this PDF.", entity.ErrInternalError)
	}

	s.logger.Error("failed to run ghostscript", fields...)
	return entity.ErrInternalError
}

// checkGhostscript fails if gs isn't installed and returns its version.
func checkGhostscript() (string, error) {
	path, err := exec.LookPath(ghostscriptBin)
	if err != nil {
		return "", fmt.Errorf("ghostscript (%s) is required for PDF compression: %w", ghostscriptBin, err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	version, err := exec.CommandContext(ctx, path, "--version").Output()
	if err != nil {
		return "", fmt.Errorf("failed to run %s --version: %w", path, err)
	}
	return strings.TrimSpace(string(version)), nil
}

// writeTemp creates a temp file with data and returns its path.
func writeTemp(pattern string, data []byte) (string, error) {
	f, err := os.CreateTemp("", pattern)
	if err != nil {
		return "", err
	}
	_, werr := f.Write(data)
	cerr := f.Close()
	if err := errors.Join(werr, cerr); err != nil {
		os.Remove(f.Name())
		return "", err
	}
	return f.Name(), nil
}

// cappedBuffer keeps the first limit bytes written to it and drops the rest.
type cappedBuffer struct {
	buf   bytes.Buffer
	limit int
}

func (b *cappedBuffer) Write(p []byte) (int, error) {
	if room := b.limit - b.buf.Len(); room > 0 {
		b.buf.Write(p[:min(len(p), room)])
	}
	return len(p), nil
}

func (b *cappedBuffer) String() string { return b.buf.String() }
