package controller

import (
	"fmt"
	"net/url"

	"github.com/devproje/neko-engine/config"
	"github.com/devproje/neko-engine/service"
	"github.com/gin-gonic/gin"
)

type AccountController struct {
	Account *service.AccountService
}

func NewAccountController(acc *service.AccountService) *AccountController {
	return &AccountController{
		Account: acc,
	}
}

func (*AccountController) genLink(sco string, it int) string {
	cnf := config.Load()
	redirect := fmt.Sprintf("redirect_uri=%s?type=%d", url.QueryEscape(cnf.Bot.RedirectURI), it)
	scope := fmt.Sprintf("scope=%s", url.QueryEscape(sco))

	return fmt.Sprintf(
		"https://discord.com/api/oauth2/authorize?response_type=code&client_id=%s&integration_type=%d&%s&%s",
		cnf.Bot.ClientId,
		it,
		scope,
		redirect,
	)
}

func (ac *AccountController) ReadAccount(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		ctx.JSON(400, gin.H{"errno": "'id' parameter must be contained"})
		return
	}

	acc, err := ac.Account.Read(id)
	if err != nil {
		ctx.JSON(404, gin.H{"errno": fmt.Sprintf("account id %s is not found", id)})
		return
	}

	ctx.JSON(200, gin.H{
		"id":   acc.Id,
		"name": acc.Author,
		"role": acc.Role,
		"usage": gin.H{
			"count": acc.Count,
			"total": acc.Total,
		},
		"created_at": acc.CreatedAt,
	})
}

func (ac *AccountController) AddDiscordApp(ctx *gin.Context) {
	link := ac.genLink("identify guilds", 1)
	fmt.Println(link)
	ctx.Redirect(302, link)
}

func (ac *AccountController) AddDiscordServer(ctx *gin.Context) {
	link := ac.genLink("bot", 0)
	fmt.Println(link)
	ctx.Redirect(302, link)
}

func (ac *AccountController) Callback(ctx *gin.Context) {
	tp := ctx.Query("type")
	code := ctx.Query("code")
	if code == "" {
		ctx.JSON(400, gin.H{"errno": "'code' parameter must be contained"})
		ctx.Abort()
		return
	}

	var message string
	switch tp {
	case "0":
		if err := ac.Account.HandleServerOAuth2Callback(code); err != nil {
			ctx.JSON(500, gin.H{"errno": err.Error()})
			ctx.Abort()
			return
		}

		message = "bot has been successfully invited to your server"
	case "1":
		var ok bool
		var err error
		if ok, err = ac.Account.HandleUserOAuth2Callback(code); err != nil {
			ctx.JSON(500, gin.H{"errno": err.Error()})
			ctx.Abort()
			return
		}

		if !ok {
			ctx.JSON(403, gin.H{"message": "you have already registered"})
			ctx.Abort()
			return
		}

		message = "account has been successfully registered"
	default:
		ctx.JSON(400, gin.H{"errno": "invalid 'type' parameter"})
		ctx.Abort()
		return
	}

	ctx.JSON(200, gin.H{
		"message": message,
	})
}
