package route

import (
	"github.com/devproje/neko-engine/common"
	"github.com/devproje/neko-engine/middleware"
	"github.com/gin-gonic/gin"
)

func InternalRouter(app *gin.Engine, sl *common.ServiceLoader) {
	app.Use(middleware.CheckBot)
	app.GET("/", func(ctx *gin.Context) {
	})

	app.POST("/chat", sl.Chat.SendChat)
}
