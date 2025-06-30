package middleware

import (
	"github.com/devproje/neko-engine/common/service"
	"github.com/gin-gonic/gin"
)

func RequireRoot(ctx *gin.Context) {
	userID := ctx.GetHeader("X-User-ID")
	if userID == "" {
		ctx.AbortWithStatusJSON(401, gin.H{
			"errno": "User ID is required",
		})
		return
	}

	accountService := service.NewAccountService()
	isRoot, err := accountService.IsRoot(userID)
	if err != nil {
		ctx.AbortWithStatusJSON(500, gin.H{
			"errno": "Failed to verify permissions",
		})
		return
	}

	if !isRoot {
		ctx.AbortWithStatusJSON(403, gin.H{
			"errno": "Root permissions required",
		})
		return
	}

	ctx.Set("userID", userID)
	ctx.Set("isRoot", true)
	ctx.Next()
}

func CheckBanStatus(ctx *gin.Context) {
	userID := ctx.GetHeader("X-User-ID")
	if userID == "" {
		ctx.AbortWithStatusJSON(401, gin.H{
			"errno": "User ID is required",
		})
		return
	}

	accountService := service.NewAccountService()
	user, err := accountService.ReadUser(userID)
	if err != nil {
		ctx.AbortWithStatusJSON(404, gin.H{
			"errno": "User not found",
		})
		return
	}

	if user.Banned {
		ctx.AbortWithStatusJSON(403, gin.H{
			"errno": "User is banned",
		})
		return
	}

	ctx.Set("userID", userID)
	ctx.Set("user", user)
	ctx.Next()
}

func RequireAuth(ctx *gin.Context) {
	userID := ctx.GetHeader("X-User-ID")
	if userID == "" {
		ctx.AbortWithStatusJSON(401, gin.H{
			"errno": "User ID is required",
		})
		return
	}

	accountService := service.NewAccountService()
	user, err := accountService.ReadUser(userID)
	if err != nil {
		ctx.AbortWithStatusJSON(404, gin.H{
			"errno": "User not found",
		})
		return
	}

	ctx.Set("userID", userID)
	ctx.Set("user", user)
	ctx.Next()
}
