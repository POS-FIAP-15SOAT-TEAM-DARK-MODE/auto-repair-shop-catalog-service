package middleware

import (
	"errors"
	"net/http"
	"strings"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-catalog-service/internal/pkg/auth"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

const (
	authorizationHeader = "Authorization"
	bearerPrefix        = "Bearer "
)

func Auth(routeRoles ...auth.Role) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader(authorizationHeader)
		if token == "" || !strings.HasPrefix(token, bearerPrefix) {
			abortWithError(c, http.StatusUnauthorized, "Unauthorized")
			return
		}

		token = strings.TrimPrefix(token, bearerPrefix)

		userClaims, err := auth.GetClaims(token)
		if err != nil {
			if errors.Is(err, jwt.ErrTokenExpired) {
				abortWithError(c, http.StatusUnauthorized, "Token expired")
				return
			}
			abortWithError(c, http.StatusUnauthorized, "Unauthorized")
			return
		}

		if !auth.HasRequiredRoles(userClaims.Roles, routeRoles) {
			abortWithError(c, http.StatusForbidden, "Forbidden")
			return
		}
	}
}

// abortWithError answers the standard error body {"code":N,"error":"..."} and stops the chain.
func abortWithError(c *gin.Context, status int, msg string) {
	c.AbortWithStatusJSON(status, gin.H{"code": status, "error": msg})
}
