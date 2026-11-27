package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (s *server) maxUploadSizeMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {

		// If the request body is larger than the configured max upload size, return a 413 error.
		if ctx.Request.ContentLength > s.config.MaxUploadSize {
			ctx.AbortWithStatus(http.StatusRequestEntityTooLarge)
			return
		}

		ctx.Request.Body = http.MaxBytesReader(ctx.Writer, ctx.Request.Body, s.config.MaxUploadSize)
		ctx.Next()
	}
}
