package middleware

import (
	"github.com/devproje/neko-engine/config"
	"github.com/gin-gonic/gin"
)

func CheckBot(ctx *gin.Context) {
	cnf := config.Load()
	if cnf.Server.Secret != ctx.GetHeader("X-BOT-API-KEY") {
		ctx.AbortWithStatusJSON(401, gin.H{
			"errno": "only bot can use this api",
		})

		return
	}

	ctx.Next()
}
