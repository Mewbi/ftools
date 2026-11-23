package service

import (
	"context"
	"fmt"

	"ftools/entity"

	"github.com/skip2/go-qrcode"
)

var qrCodeLevels = map[entity.QRCodeLevel]qrcode.RecoveryLevel{
	entity.QRCodeLevelLow:      qrcode.Low,
	entity.QRCodeLevelMedium:   qrcode.Medium,
	entity.QRCodeLevelQuartile: qrcode.High,
	entity.QRCodeLevelHigh:     qrcode.Highest,
}

func (s *service) GenerateQRCode(ctx context.Context, req entity.QRCodeRequest) (entity.QRCode, error) {
	level, ok := qrCodeLevels[req.Level]
	if !ok {
		return entity.QRCode{}, fmt.Errorf("generate qr code: %w: %q", entity.ErrQRCodeInvalidLevel, req.Level)
	}

	code, err := qrcode.New(req.Content, level)
	if err != nil {
		return entity.QRCode{}, fmt.Errorf("generate qr code: %w", err)
	}

	png, err := code.PNG(req.Size)
	if err != nil {
		return entity.QRCode{}, fmt.Errorf("encode qr code png: %w", err)
	}

	return entity.QRCode{PNG: png, Size: req.Size, Content: req.Content}, nil
}
