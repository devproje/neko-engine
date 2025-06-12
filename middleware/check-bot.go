package middleware

import (
	"github.com/devproje/neko-engine/config"
	"github.com/gin-gonic/gin"
)

func CheckBot(ctx *gin.Context) bool {
	cnf := config.Load()
	key := ctx.GetHeader("X-BOT-API-KEY")
	if key == "" {
		ctx.AbortWithStatusJSON(403, gin.H{
			"errno": "this features can use only discord bot",
		})

		return false
	}

	if key != cnf.Server.Secret {
		ctx.AbortWithStatusJSON(401, gin.H{
			"errno": "secret key is not matches",
		})

		return false
	}

	ctx.Next()
	return true
}
