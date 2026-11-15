package routing

import (
	"context"
	http2 "net/http"

	api "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-catalog-service/docs"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-catalog-service/internal/app/container"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-catalog-service/internal/app/middleware"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-catalog-service/internal/pkg/auth"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-catalog-service/internal/pkg/logger"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-catalog-service/internal/pkg/web"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/zap"
)

// SetupRouter wires the public /v1 API (works and supplies) and the internal
// /internal API used by other services. internalToken guards /internal/*.
func SetupRouter(c *container.HandlersWrapper, m *container.Middlewares, internalToken string) *gin.Engine {
	router := gin.New()

	for name, mid := range *m {
		logger.Global().Info("Starting new middleware", zap.String("name", name))
		router.Use(mid)
	}

	mountSwaggerUI(router)

	router.GET("/ping", func(c *gin.Context) {
		c.JSON(http2.StatusOK, gin.H{"message": "pong"})
	})

	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	v1 := router.Group("/v1")

	v1.GET("/works", middleware.Auth(auth.AttendantAndMechanicRoles...), GinHandler(c.WorkHandler.List))
	v1.POST("/works", middleware.Auth(auth.AttendantRoles...), GinHandler(c.WorkHandler.Create, http2.StatusCreated))
	v1.PUT("/works/:id", middleware.Auth(auth.AttendantRoles...), GinHandler(c.WorkHandler.Update))
	v1.DELETE("/works/:id", middleware.Auth(auth.AttendantRoles...), GinOnlyErrorHandler(c.WorkHandler.Delete))

	v1.GET("/supplies", middleware.Auth(auth.AttendantAndMechanicRoles...), GinHandler(c.SupplyHandler.List))
	v1.POST("/supplies", middleware.Auth(auth.AttendantAndMechanicRoles...), GinHandler(c.SupplyHandler.Create, http2.StatusCreated))
	v1.PUT("/supplies/:id", middleware.Auth(auth.AttendantAndMechanicRoles...), GinHandler(c.SupplyHandler.Update))
	v1.DELETE("/supplies/:id", middleware.Auth(auth.AttendantAndMechanicRoles...), GinOnlyErrorHandler(c.SupplyHandler.Delete))

	internal := router.Group("/internal", middleware.InternalAuth(internalToken))
	internal.GET("/works/:id", GinHandler(c.WorkInternalHandler.Get))
	internal.POST("/reservations", GinHandler(c.ReservationHandler.Reserve, http2.StatusCreated))
	internal.DELETE("/reservations/:id", GinOnlyErrorHandler(c.ReservationHandler.Release))

	return router
}

func mountSwaggerUI(router *gin.Engine) {
	router.GET("/swagger.yaml", func(c *gin.Context) {
		c.Data(http2.StatusOK, "application/yaml; charset=utf-8", api.SwaggerBytes)
	})

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, ginSwagger.URL("/swagger.yaml")))
}

func GinOnlyErrorHandler(cb func(context.Context, *http2.Request) error, customStatus ...int) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		// context keys are gin param names (strings) intentionally;
		// collision risk is acceptable here since params are route-scoped.
		for _, v := range c.Params {
			ctx = context.WithValue(ctx, v.Key, v.Value) //nolint:staticcheck
		}

		if err := cb(ctx, c.Request); err != nil {
			respondError(c, err)
			return
		}

		if len(customStatus) == 0 {
			c.Status(http2.StatusNoContent)
			c.Writer.WriteHeaderNow()
			return
		}

		c.Status(customStatus[0])
		c.Writer.WriteHeaderNow()
	}
}

func GinHandler[T any](cb func(context.Context, *http2.Request) (T, error), customStatus ...int) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		// context keys are gin param names (strings) intentionally;
		// collision risk is acceptable here since params are route-scoped.
		for _, v := range c.Params {
			ctx = context.WithValue(ctx, v.Key, v.Value) //nolint:staticcheck
		}

		resp, err := cb(ctx, c.Request)
		if err != nil {
			respondError(c, err)
			return
		}

		if len(customStatus) == 0 {
			c.JSON(http2.StatusOK, resp)
			return
		}

		c.JSON(customStatus[0], resp)
	}
}

// respondError writes the public error body and logs the full cause server-side
// (5xx and infrastructure 409/422) with the request_id carried by the request context.
func respondError(c *gin.Context, err error) {
	status, errResp := web.Error(err)
	if status >= http2.StatusInternalServerError || web.IsInfrastructureError(err) {
		logger.Of(c.Request.Context()).Error(err, zap.Int("status", status))
	}
	c.JSON(status, errResp)
}
