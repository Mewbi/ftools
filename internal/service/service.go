package service

import (
	"context"

	"ftools/entity"

	"go.uber.org/fx"
)

type Service interface {
	GenerateQRCode(ctx context.Context, req entity.QRCodeRequest) (entity.QRCode, error)
}

var _ Service = (*service)(nil)

var Module = fx.Module("service", fx.Provide(New))

type Params struct {
	fx.In
}

type service struct{}

func New(p Params) Service {
	return &service{}
}
