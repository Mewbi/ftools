package web

import (
	"net/http"

	"ftools/internal/service"

	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type WebController interface {
	Home(ctx *gin.Context)
	QRCodePage(ctx *gin.Context)
	ImagePage(ctx *gin.Context)
	PDFPage(ctx *gin.Context)

	GenerateQRCode(ctx *gin.Context)
}

var _ WebController = (*controller)(nil)

var Module = fx.Module("web_controller", fx.Provide(New))

type Params struct {
	fx.In
	Service service.Service
	Logger  *zap.Logger
}

type controller struct {
	service service.Service
	logger  *zap.Logger
}

func New(p Params) WebController {
	return &controller{
		service: p.Service,
		logger:  p.Logger,
	}
}

func (c *controller) Home(ctx *gin.Context) {
	c.page(ctx, "pages/home", "")
}

// page renders a full page with the data every layout needs.
func (c *controller) page(ctx *gin.Context, name, title string) {
	ctx.HTML(http.StatusOK, name, gin.H{
		"Title": title,
		"Path":  ctx.Request.URL.Path,
	})
}
