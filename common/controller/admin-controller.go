package controller

import (
	"strconv"

	"github.com/devproje/neko-engine/common/service"
	"github.com/gin-gonic/gin"
)

type AdminController struct {
	Account *service.AccountService
}

type AdminRequest struct {
	TargetUserID string `json:"target_user_id" binding:"required"`
}

func NewAdminController(account *service.AccountService) *AdminController {
	return &AdminController{Account: account}
}

func (ac *AdminController) BanUser(ctx *gin.Context) {
	var req AdminRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(400, gin.H{
			"errno": "Invalid request format",
		})
		return
	}

	adminID := ctx.GetString("userID")
	if adminID == "" {
		ctx.JSON(401, gin.H{
			"errno": "Admin ID not found",
		})
		return
	}

	if err := ac.Account.BanUser(req.TargetUserID, adminID); err != nil {
		ctx.JSON(400, gin.H{
			"errno": err.Error(),
		})
		return
	}

	ctx.JSON(200, gin.H{
		"message":        "User banned successfully",
		"target_user_id": req.TargetUserID,
	})
}

func (ac *AdminController) UnbanUser(ctx *gin.Context) {
	var req AdminRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(400, gin.H{
			"errno": "Invalid request format",
		})
		return
	}

	adminID := ctx.GetString("userID")
	if adminID == "" {
		ctx.JSON(401, gin.H{
			"errno": "Admin ID not found",
		})
		return
	}

	if err := ac.Account.UnbanUser(req.TargetUserID, adminID); err != nil {
		ctx.JSON(400, gin.H{
			"errno": err.Error(),
		})
		return
	}

	ctx.JSON(200, gin.H{
		"message":        "User unbanned successfully",
		"target_user_id": req.TargetUserID,
	})
}

func (ac *AdminController) PromoteToRoot(ctx *gin.Context) {
	var req AdminRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(400, gin.H{
			"errno": "Invalid request format",
		})
		return
	}

	adminID := ctx.GetString("userID")
	if adminID == "" {
		ctx.JSON(401, gin.H{
			"errno": "Admin ID not found",
		})
		return
	}

	if err := ac.Account.PromoteToRoot(req.TargetUserID, adminID); err != nil {
		ctx.JSON(400, gin.H{
			"errno": err.Error(),
		})
		return
	}

	ctx.JSON(200, gin.H{
		"message":        "User promoted to root successfully",
		"target_user_id": req.TargetUserID,
	})
}

func (ac *AdminController) DemoteFromRoot(ctx *gin.Context) {
	var req AdminRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(400, gin.H{
			"errno": "Invalid request format",
		})
		return
	}

	adminID := ctx.GetString("userID")
	if adminID == "" {
		ctx.JSON(401, gin.H{
			"errno": "Admin ID not found",
		})
		return
	}

	if err := ac.Account.DemoteFromRoot(req.TargetUserID, adminID); err != nil {
		ctx.JSON(400, gin.H{
			"errno": err.Error(),
		})
		return
	}

	ctx.JSON(200, gin.H{
		"message":        "User demoted from root successfully",
		"target_user_id": req.TargetUserID,
	})
}

func (ac *AdminController) GetUserInfo(ctx *gin.Context) {
	userID := ctx.Param("user_id")
	if userID == "" {
		ctx.JSON(400, gin.H{
			"errno": "User ID is required",
		})
		return
	}

	user, err := ac.Account.ReadUser(userID)
	if err != nil {
		ctx.JSON(404, gin.H{
			"errno": "User not found",
		})
		return
	}

	role, _ := ac.Account.GetRoleById(user.RoleID)

	ctx.JSON(200, gin.H{
		"user_id":  user.ID,
		"username": user.Username,
		"role": gin.H{
			"id":      user.RoleID,
			"name":    role.Name,
			"is_root": role.Root,
		},
		"banned": user.Banned,
		"count":  user.Count,
		"total":  user.Total,
	})
}

func (ac *AdminController) ListUsers(ctx *gin.Context) {
	limitStr := ctx.DefaultQuery("limit", "20")
	offsetStr := ctx.DefaultQuery("offset", "0")

	limit, _ := strconv.Atoi(limitStr)
	offset, _ := strconv.Atoi(offsetStr)

	if limit <= 0 || limit > 100 {
		limit = 20
	}

	users, total, err := ac.Account.ListUsers(limit, offset)
	if err != nil {
		ctx.JSON(500, gin.H{
			"errno": "Failed to load users",
		})
		return
	}

	ctx.JSON(200, gin.H{
		"users":  users,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

func (ac *AdminController) SearchUsers(ctx *gin.Context) {
	query := ctx.Query("q")
	if query == "" {
		ctx.JSON(400, gin.H{
			"errno": "Search query is required",
		})
		return
	}

	limitStr := ctx.DefaultQuery("limit", "20")
	offsetStr := ctx.DefaultQuery("offset", "0")

	limit, _ := strconv.Atoi(limitStr)
	offset, _ := strconv.Atoi(offsetStr)

	if limit <= 0 || limit > 100 {
		limit = 20
	}

	users, total, err := ac.Account.SearchUsers(query, limit, offset)
	if err != nil {
		ctx.JSON(500, gin.H{
			"errno": "Failed to search users",
		})
		return
	}

	ctx.JSON(200, gin.H{
		"users":  users,
		"total":  total,
		"query":  query,
		"limit":  limit,
		"offset": offset,
	})
}

func (ac *AdminController) ResetUserCount(ctx *gin.Context) {
	if err := ac.Account.ResetCount(); err != nil {
		ctx.JSON(500, gin.H{
			"errno": "Failed to reset user counts",
		})
		return
	}

	ctx.JSON(200, gin.H{
		"message": "All user counts reset successfully",
	})
}
