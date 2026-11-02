package middleware

import (
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-catalog-service/internal/pkg/logger"
)

func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				stack := debug.Stack()
				logger.Of(c).Error(err.(error),
					zap.String("behavior", "[PANIC RECOVERED]"),
					zap.String("stack", string(stack)),
				)

				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"code":  http.StatusInternalServerError,
					"error": "internal server error",
				})
			}
		}()

		c.Next()
	}
}

func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) == 0 {
			return
		}

		errs := c.Errors.Errors()

		status := c.Writer.Status()
		if status == http.StatusOK {
			status = http.StatusInternalServerError
		}

		if status >= http.StatusInternalServerError {
			// Causes stay in the server log; the client gets only the generic text.
			c.JSON(status, gin.H{"code": status, "error": http.StatusText(status)})
			return
		}

		c.JSON(status, gin.H{
			"code":    status,
			"error":   http.StatusText(status),
			"details": errs,
		})
	}
}
