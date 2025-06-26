package route

import (
	"github.com/devproje/neko-engine/common"
	"github.com/gin-gonic/gin"
)

func InternalRouter(app *gin.Engine, sl *common.ServiceLoader) {
	app.GET("/", func(ctx *gin.Context) {
	})

	app.POST("/chat", func(ctx *gin.Context) {
	})
}
