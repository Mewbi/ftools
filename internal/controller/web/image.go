package web

import "github.com/gin-gonic/gin"

func (c *controller) ImagePage(ctx *gin.Context) {
	c.page(ctx, "pages/tools/image", "Compress and resize images")
}
