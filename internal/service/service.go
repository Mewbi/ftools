package service

import (
	"context"

	"ftools/entity"

	"go.uber.org/fx"
	"go.uber.org/zap"
)

type Service interface {
	GenerateQRCode(ctx context.Context, req entity.QRCodeRequest) (entity.QRCode, error)
}

var _ Service = (*service)(nil)

var Module = fx.Module("service", fx.Provide(New))

type Params struct {
	fx.In
	Logger *zap.Logger
}

type service struct {
	logger *zap.Logger
}

func New(p Params) Service {
	return &service{
		logger: p.Logger,
	}
}
