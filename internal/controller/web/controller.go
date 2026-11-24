package web

import (
	"errors"
	"net/http"

	"ftools/entity"
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

// convertError maps a domain error to an HTTP status and renders the error partial.
// Only client errors expose their message; anything else gets a generic one.
func (c *controller) convertError(ctx *gin.Context, err error) {
	switch {
	case errors.Is(err, entity.ErrInvalidRequest):
		ctx.HTML(http.StatusUnprocessableEntity, "partials/error", gin.H{"Message": err.Error()})
	case errors.Is(err, entity.ErrInternalError):
		ctx.HTML(http.StatusInternalServerError, "partials/error", gin.H{
			"Message": "The server couldn't process your request. Try again in a moment.",
		})
	default:
		c.logger.Error("unhandled error", zap.Error(err), zap.String("path", ctx.Request.URL.Path))
		ctx.HTML(http.StatusInternalServerError, "partials/error", gin.H{
			"Message": "The server couldn't process your request. Try again in a moment.",
		})
	}
}
