package entity

import (
	"fmt"
	"strings"
)

// QRCodeLevel is the error correction level: how much of the code can be
// damaged or covered and still scan. Higher levels hold less content.
type QRCodeLevel string

const (
	QRCodeLevelLow      QRCodeLevel = "low"      // ~7% recoverable
	QRCodeLevelMedium   QRCodeLevel = "medium"   // ~15%
	QRCodeLevelQuartile QRCodeLevel = "quartile" // ~25%
	QRCodeLevelHigh     QRCodeLevel = "high"     // ~30%
)

// qrCodeCapacity is the most bytes a QR code can hold per error correction level
// (version 40, byte mode). Content is UTF-8, so this counts bytes, not characters.
var qrCodeCapacity = map[QRCodeLevel]int{
	QRCodeLevelLow:      2953,
	QRCodeLevelMedium:   2331,
	QRCodeLevelQuartile: 1663,
	QRCodeLevelHigh:     1273,
}

// qrCodeSizes are the allowed image sizes in pixels.
var qrCodeSizes = map[int]bool{256: true, 512: true, 1024: true}

type QRCodeRequest struct {
	Content string
	Size    int // image width and height in pixels
	Level   QRCodeLevel
}

// Validate checks the request before generating a QR code.
func (r QRCodeRequest) Validate() error {
	if strings.TrimSpace(r.Content) == "" {
		return fmt.Errorf("%w: qr content cannot be empty", ErrInvalidRequest)
	}
	if !qrCodeSizes[r.Size] {
		return fmt.Errorf("%w: invalid size %d", ErrInvalidRequest, r.Size)
	}
	capacity, ok := qrCodeCapacity[r.Level]
	if !ok {
		return fmt.Errorf("%w: invalid level %q", ErrInvalidRequest, r.Level)
	}
	if len(r.Content) > capacity {
		return fmt.Errorf("%w: content length %d exceeds capacity", ErrInvalidRequest, len(r.Content))
	}
	return nil
}

type QRCode struct {
	PNG     []byte
	Size    int
	Content string
}
