package api

import (
	"github.com/gin-gonic/gin"
	"github.com/tareqpi/transfer-system/internal/config"
	"github.com/tareqpi/transfer-system/internal/logger"
	"github.com/tareqpi/transfer-system/internal/service"
	"go.uber.org/zap"
)

func Setup(applicationService service.Service) {
	router := gin.New()
	router.Use(RequestID(), Logging(), Recovery())

	router.StaticFile("/openapi.yaml", "/docs/openapi.yaml")

	router.GET("/docs", func(c *gin.Context) {
		const html = `<!doctype html>
<html>
  <head>
    <meta charset="utf-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1" />
    <title>Transfer System API Docs</title>
    <style>body { margin: 0; padding: 0; }</style>
    <link rel="icon" href="data:," />
  </head>
  <body>
    <redoc spec-url="/openapi.yaml"></redoc>
    <script src="https://cdn.redoc.ly/redoc/latest/bundles/redoc.standalone.js"></script>
  </body>
</html>`
		c.Data(200, "text/html; charset=utf-8", []byte(html))
	})
	v1 := router.Group("/api/v1")
	handler := NewHandler(applicationService)

	account := v1.Group("/accounts")
	{
		account.POST("", handler.CreateAccount)
		account.GET("/:account_id", handler.GetAccount)
	}

	transaction := v1.Group("/transactions")
	{
		transaction.POST("", handler.TransferMoney)
	}
	if err := router.Run(":" + config.Get().Port); err != nil {
		logger.L().Fatal("failed to start HTTP server", zap.Error(err))
	}
}
