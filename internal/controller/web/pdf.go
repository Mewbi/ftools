package web

import "github.com/gin-gonic/gin"

func (c *controller) PDFPage(ctx *gin.Context) {
	c.page(ctx, "pages/tools/pdf", "Compress PDF")
}
