package service

import (
	"context"
	"runtime"
	"time"

	"ftools/config"
	"ftools/entity"

	"go.uber.org/fx"
	"go.uber.org/zap"
)

type Service interface {
	GenerateQRCode(ctx context.Context, req entity.QRCodeRequest) (entity.QRCode, error)
	CompressPDF(ctx context.Context, req entity.PDFCompressRequest) (entity.PDFCompressResult, error)
}

var _ Service = (*service)(nil)

var Module = fx.Module("service", fx.Provide(New))

type Params struct {
	fx.In
	Logger *zap.Logger
	PDF    config.PDF
}

type service struct {
	logger   *zap.Logger
	pdf      config.PDF
	pdfSlots chan struct{}
}

func New(p Params) (Service, error) {
	version, err := checkGhostscript()
	if err != nil {
		return nil, err
	}
	p.Logger.Info("ghostscript found", zap.String("version", version))

	pdf := p.PDF
	if pdf.Timeout <= 0 {
		pdf.Timeout = 60 * time.Second
	}
	if pdf.QueueTimeout <= 0 {
		pdf.QueueTimeout = 30 * time.Second
	}
	if pdf.MaxConcurrent <= 0 {
		pdf.MaxConcurrent = runtime.GOMAXPROCS(0)
	}

	return &service{
		logger:   p.Logger,
		pdf:      pdf,
		pdfSlots: make(chan struct{}, pdf.MaxConcurrent),
	}, nil
}
