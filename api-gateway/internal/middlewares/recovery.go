package middlewares

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
)

func Recovery(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				requestID, _ := c.Get("request_id")

				logger.Error("panic recovered",
					"error", err,
					"request_id", requestID,
					"path", c.Request.URL.Path,
					"method", c.Request.Method,
				)

				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"title":  "Internal Server Error",
					"type":   "https://my.app.com/errors/internal",
					"status": http.StatusInternalServerError,
					"detail": "An unexpected error has occurred",
				})
			}
		}()

		c.Next()
	}
}
