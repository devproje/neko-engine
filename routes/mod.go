package routes

import (
	"github.com/devproje/neko-engine/controller"
	"github.com/devproje/neko-engine/service"
	"github.com/gin-gonic/gin"
)

func New(app *gin.Engine) {
	api := app.Group("/api")

	gemini := service.NewGeminiService()
	account := service.NewAccountService()
	history := service.NewHistoryService()
	memory := service.NewMemoryService(account, gemini)
	discord := controller.NewAccountController(account)
	chat := controller.NewChatController(gemini, account, history, memory)

	api.GET("/discord/add", discord.AddDiscordApp)
	api.GET("/discord/invite", discord.AddDiscordServer)
	api.GET("/discord/callback", discord.Callback)
	api.GET("/discord/@me/:id", discord.ReadAccount)

	api.POST("/chat/discord", chat.Hit)
}
