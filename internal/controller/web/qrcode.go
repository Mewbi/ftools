package web

import (
	"encoding/base64"
	"html/template"
	"net/http"
	"strconv"

	"ftools/entity"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func (c *controller) QRCodePage(ctx *gin.Context) {
	c.page(ctx, "pages/tools/qrcode", "QR code generator")
}

func (c *controller) GenerateQRCode(ctx *gin.Context) {
	// An invalid number leaves Size at 0, which Validate rejects.
	size, _ := strconv.Atoi(ctx.PostForm("size"))

	req := entity.QRCodeRequest{
		Content: ctx.PostForm("content"),
		Size:    size,
		Level:   entity.QRCodeLevel(ctx.PostForm("level")),
	}
	if err := req.Validate(); err != nil {
		c.convertError(ctx, err)
		return
	}

	qr, err := c.service.GenerateQRCode(ctx.Request.Context(), req)
	if err != nil {
		c.logger.Error("generate qr code", zap.Error(err))
		c.convertError(ctx, err)
		return
	}

	ctx.HTML(http.StatusOK, "partials/qrcode_result", gin.H{
		"DataURL": template.URL("data:image/png;base64," + base64.StdEncoding.EncodeToString(qr.PNG)),
		"Content": qr.Content,
		"Size":    qr.Size,
	})
}
