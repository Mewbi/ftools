package entity

import (
	"bytes"
	"fmt"
)

// PDFLevel is how hard to compress. Each level maps to a Ghostscript preset.
type PDFLevel string

const (
	PDFLevelLow    PDFLevel = "low"    // /printer
	PDFLevelMedium PDFLevel = "medium" // /ebook
	PDFLevelHigh   PDFLevel = "high"   // /screen
)

// PDFColor says whether to keep colors or convert the document to grayscale.
type PDFColor string

const (
	PDFColorOriginal  PDFColor = "original"
	PDFColorGrayscale PDFColor = "grayscale"
)

// Allowed image resolution range, in dpi.
const (
	PDFMinDPI = 72
	PDFMaxDPI = 300
)

var (
	pdfLevels     = map[PDFLevel]bool{PDFLevelLow: true, PDFLevelMedium: true, PDFLevelHigh: true}
	pdfColors     = map[PDFColor]bool{PDFColorOriginal: true, PDFColorGrayscale: true}
	pdfFilePrefix = []byte("%PDF-")
)

type PDFCompressRequest struct {
	File  []byte // the uploaded PDF
	Level PDFLevel
	Color PDFColor
	DPI   int // target resolution for color and gray images
}

// Validate checks the request before compressing.
func (r PDFCompressRequest) Validate() error {
	if len(r.File) == 0 {
		return fmt.Errorf("%w: choose a PDF to compress", ErrInvalidRequest)
	}
	if !bytes.HasPrefix(r.File, pdfFilePrefix) {
		return fmt.Errorf("%w: this file isn't a PDF", ErrInvalidRequest)
	}
	if !pdfLevels[r.Level] {
		return fmt.Errorf("%w: invalid compression level %q", ErrInvalidRequest, r.Level)
	}
	if !pdfColors[r.Color] {
		return fmt.Errorf("%w: invalid color option %q", ErrInvalidRequest, r.Color)
	}
	if r.DPI < PDFMinDPI || r.DPI > PDFMaxDPI {
		return fmt.Errorf("%w: image resolution must be between %d and %d dpi", ErrInvalidRequest, PDFMinDPI, PDFMaxDPI)
	}
	return nil
}

type PDFCompressResult struct {
	PDF          []byte
	OriginalSize int
	ResultSize   int
	// AlreadyOptimized is true when compressing didn't make the file smaller,
	// so PDF holds the original file.
	AlreadyOptimized bool
}
