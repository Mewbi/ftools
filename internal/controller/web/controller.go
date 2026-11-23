package web

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
)

type WebController interface {
	Home(ctx *gin.Context)
	QRCodePage(ctx *gin.Context)
	ImagePage(ctx *gin.Context)
	PDFPage(ctx *gin.Context)
}

var _ WebController = (*controller)(nil)

var Module = fx.Module("web_controller", fx.Provide(New))

type Params struct {
	fx.In
}

type controller struct{}

func New(p Params) WebController {
	return &controller{}
}

func (c *controller) Home(ctx *gin.Context) {
	c.page(ctx, "pages/home", "")
}

func (c *controller) QRCodePage(ctx *gin.Context) {
	c.page(ctx, "pages/tools/qrcode", "QR code generator")
}

func (c *controller) ImagePage(ctx *gin.Context) {
	c.page(ctx, "pages/tools/image", "Compress and resize images")
}

func (c *controller) PDFPage(ctx *gin.Context) {
	c.page(ctx, "pages/tools/pdf", "Compress PDF")
}

// page renders a full page with the data every layout needs.
func (c *controller) page(ctx *gin.Context, name, title string) {
	ctx.HTML(http.StatusOK, name, gin.H{
		"Title": title,
		"Path":  ctx.Request.URL.Path,
	})
}
