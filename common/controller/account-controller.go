package controller

import (
	"github.com/devproje/neko-engine/common/service"
	"github.com/gin-gonic/gin"
)

type AccountController struct {
	Acc *service.AccountService
}

type UserForm struct {
	Id     string `json:"id"`
	Author string `json:"author"`
}

func NewAccountController(acc *service.AccountService) *AccountController {
	return &AccountController{Acc: acc}
}

func (ac *AccountController) RegisterUser(ctx *gin.Context) {
	var req UserForm
	if err := ctx.ShouldBindBodyWithJSON(&req); err != nil {
		ctx.JSON(400, gin.H{
			"errno": "JSON input parameter is missing. Please check the sent values.",
		})
		return
	}

	if err := ac.Acc.CreateUser(req.Id, req.Author); err != nil {
		ctx.JSON(403, gin.H{
			"errno": "Account already exists.",
		})
		return
	}

	ctx.JSON(200, gin.H{
		"message": "Congratulations on registering! You can now use the bot normally!",
	})
}

func (ac *AccountController) FetchAccount(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		ctx.JSON(400, gin.H{
			"errno": "The \"id\" parameter is missing.",
		})
		return
	}

	acc, err := ac.Acc.ReadUser(id)
	if err != nil {
		ctx.JSON(401, gin.H{
			"errno": "Could not find account information.",
		})
		return
	}

	role, _ := ac.Acc.GetRoleById(acc.RoleID)
	ctx.JSON(200, gin.H{
		"id":       acc.ID,
		"role":     role.Name,
		"limit":    role.Limit,
		"nickname": acc.Username,
		"usage": gin.H{
			"current": acc.Count,
			"total":   acc.Total,
		},
	})
}
