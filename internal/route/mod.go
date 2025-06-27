package route

import (
	"github.com/devproje/neko-engine/common"
	"github.com/devproje/neko-engine/config"
	"github.com/devproje/neko-engine/middleware"
	"github.com/gin-gonic/gin"
)

func InternalRouter(app *gin.Engine, sl *common.ServiceLoader) {
	app.Use(middleware.CheckBot)

	app.GET("/", func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{
			"service": "neko-engine",
			"status":  "running",
			"info":    config.GetVersionInfo(),
		})
	})
	app.GET("/@me/:id", sl.Acc.FetchAccount)

	app.POST("/chat", sl.Chat.SendChat)
	app.POST("/register", sl.Acc.RegisterUser)
}
