package controller

import (
	"context"
	"ftools/config"
	"ftools/frontend"
	"ftools/internal/controller/web"
	"html/template"
	"net"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type server struct {
	logger        *zap.Logger
	config        config.Server
	webController web.WebController
	server        *http.Server
}

var Module = fx.Module("server",
	fx.Provide(New),
	fx.Invoke(func(*server) {}), // Force start server
)

type Params struct {
	fx.In
	Lifecycle     fx.Lifecycle
	Config        config.Server
	WebController web.WebController
	Logger        *zap.Logger
}

func (s *server) setupWebRoutes(r *gin.Engine) error {
	tpl, err := template.ParseFS(frontend.Templates, frontend.TemplatePatterns...)
	if err != nil {
		return err
	}

	r.SetHTMLTemplate(tpl)
	r.StaticFS("/static", http.FS(frontend.Static()))

	r.GET("/", s.webController.Home)
	r.GET("qrcode", s.webController.QRCodePage)
	r.GET("image", s.webController.ImagePage)
	r.GET("pdf", s.webController.PDFPage)

	r.POST("qrcode/generate", s.webController.GenerateQRCode)
	r.POST("pdf/compress", s.webController.CompressPDF)

	return nil
}

func New(p Params) *server {
	r := gin.New()
	r.Use(gin.Recovery())

	s := &server{
		webController: p.WebController,
		logger:        p.Logger,
		config:        p.Config,
	}

	r.Use(s.maxUploadSizeMiddleware())

	if err := s.setupWebRoutes(r); err != nil {
		p.Logger.Fatal("failed to setup web routes", zap.Error(err))
	}

	srv := &http.Server{Addr: p.Config.Addr, Handler: r}

	p.Lifecycle.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			ln, err := net.Listen("tcp", srv.Addr)
			if err != nil {
				return err
			}
			go srv.Serve(ln)
			return nil
		},

		OnStop: func(ctx context.Context) error {
			return srv.Shutdown(ctx)
		},
	})

	return s
}
