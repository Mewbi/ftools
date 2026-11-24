package service

import (
	"context"
	"fmt"

	"ftools/entity"

	"github.com/skip2/go-qrcode"
	"go.uber.org/zap"
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
		return entity.QRCode{}, fmt.Errorf("%w: invalid request level: %s", entity.ErrInvalidRequest, req.Level)
	}

	code, err := qrcode.New(req.Content, level)
	if err != nil {
		s.logger.Error("failed to generate QR code", zap.Error(err))
		return entity.QRCode{}, entity.ErrInternalError
	}

	png, err := code.PNG(req.Size)
	if err != nil {
		s.logger.Error("failed to encode QR code to PNG", zap.Error(err))
		return entity.QRCode{}, entity.ErrInternalError
	}

	return entity.QRCode{PNG: png, Size: req.Size, Content: req.Content}, nil
}
