package controller

import (
	"github.com/devproje/neko-engine/common/service"
	"github.com/gin-gonic/gin"
)

type DiscordController struct {
	accountService *service.AccountService
}

func NewDiscordController(accountService *service.AccountService) *DiscordController {
	return &DiscordController{
		accountService: accountService,
	}
}

func (dc *DiscordController) DiscordLogin(ctx *gin.Context) {
	ctx.JSON(501, gin.H{
		"error": "Discord OAuth2 login not implemented yet",
	})
}

func (dc *DiscordController) DiscordCallback(ctx *gin.Context) {
	ctx.JSON(501, gin.H{
		"error": "Discord OAuth2 callback not implemented yet",
	})
}

func (dc *DiscordController) Logout(ctx *gin.Context) {
	ctx.JSON(501, gin.H{
		"error": "Logout not implemented yet",
	})
}

func (dc *DiscordController) GetMe(ctx *gin.Context) {
	ctx.JSON(501, gin.H{
		"error": "Get me endpoint not implemented yet",
	})
}

func (dc *DiscordController) UpdateProfile(ctx *gin.Context) {
	ctx.JSON(501, gin.H{
		"error": "Update profile not implemented yet",
	})
}
