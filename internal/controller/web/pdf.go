package web

import (
	"encoding/base64"
	"errors"
	"fmt"
	"html/template"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"ftools/entity"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func (c *controller) PDFPage(ctx *gin.Context) {
	c.page(ctx, "pages/tools/pdf", "Compress PDF")
}

func (c *controller) CompressPDF(ctx *gin.Context) {
	header, err := ctx.FormFile("file")
	if err != nil {
		var tooBig *http.MaxBytesError
		if errors.As(err, &tooBig) {
			c.convertError(ctx, fmt.Errorf("%w: file exceeds %d bytes", entity.ErrInvalidRequest, tooBig.Limit))
			return
		}
		c.convertError(ctx, fmt.Errorf("%w: choose a PDF to compress", entity.ErrInvalidRequest))
		return
	}
	data, err := readFormFile(header)
	if err != nil {
		c.logger.Error("read uploaded pdf", zap.Error(err))
		c.convertError(ctx, entity.ErrInternalError)
		return
	}

	// An invalid number leaves DPI at 0, which Validate rejects.
	dpi, _ := strconv.Atoi(ctx.PostForm("dpi"))

	req := entity.PDFCompressRequest{
		File:  data,
		Level: entity.PDFLevel(ctx.PostForm("level")),
		Color: entity.PDFColor(ctx.PostForm("color")),
		DPI:   dpi,
	}
	if err := req.Validate(); err != nil {
		c.convertError(ctx, err)
		return
	}

	res, err := c.service.CompressPDF(ctx.Request.Context(), req)
	if err != nil {
		c.convertError(ctx, err)
		return
	}

	saved := 0
	if !res.AlreadyOptimized && res.OriginalSize > 0 {
		saved = 100 - res.ResultSize*100/res.OriginalSize
	}

	// TODO: data URLs are heavy for multi-MB PDFs. Switch to a temp file and
	// GET /download/:id once that's decided (README, "Downloads and cleanup").
	ctx.HTML(http.StatusOK, "partials/file_result", gin.H{
		"Filename":     header.Filename,
		"OriginalSize": formatSize(res.OriginalSize),
		"ResultSize":   formatSize(res.ResultSize),
		"Saved":        saved,
		"DownloadURL":  template.URL("data:application/pdf;base64," + base64.StdEncoding.EncodeToString(res.PDF)),
		"DownloadName": compressedName(header.Filename),
	})
}

func readFormFile(header *multipart.FileHeader) ([]byte, error) {
	f, err := header.Open()
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return io.ReadAll(f)
}

// compressedName suggests a download name: "report.pdf" becomes "report-compressed.pdf".
func compressedName(filename string) string {
	base := strings.TrimSuffix(filepath.Base(filename), filepath.Ext(filename))
	if base == "" || base == "." || base == "/" {
		base = "document"
	}
	return base + "-compressed.pdf"
}

// formatSize formats a byte count for people, e.g. "2.4 MB" or "640 KB".
func formatSize(n int) string {
	const unit = 1024
	switch {
	case n < unit:
		return fmt.Sprintf("%d B", n)
	case n < unit*unit:
		return fmt.Sprintf("%.0f KB", float64(n)/unit)
	default:
		return fmt.Sprintf("%.1f MB", float64(n)/(unit*unit))
	}
}
