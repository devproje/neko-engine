package controller

import (
	"github.com/devproje/neko-engine/common/service"
	"github.com/gin-gonic/gin"
)

type StatsController struct {
	Account *service.AccountService
	Memory  *service.MemoryService
}

func NewStatsController(account *service.AccountService, memory *service.MemoryService) *StatsController {
	return &StatsController{Account: account, Memory: memory}
}

func (sc *StatsController) GetSystemStats(ctx *gin.Context) {
	userStats, err := sc.Account.GetUserStats()
	if err != nil {
		ctx.JSON(500, gin.H{
			"errno": "Failed to get user stats",
		})
		return
	}

	memoryStats, err := sc.Memory.GetMemoryStats()
	if err != nil {
		ctx.JSON(500, gin.H{
			"errno": "Failed to get memory stats",
		})
		return
	}

	roleStats, err := sc.Account.GetRoleStats()
	if err != nil {
		ctx.JSON(500, gin.H{
			"errno": "Failed to get role stats",
		})
		return
	}

	ctx.JSON(200, gin.H{
		"users":    userStats,
		"memories": memoryStats,
		"roles":    roleStats,
	})
}

func (sc *StatsController) GetUserStats(ctx *gin.Context) {
	stats, err := sc.Account.GetUserStats()
	if err != nil {
		ctx.JSON(500, gin.H{
			"errno": "Failed to get user statistics",
		})
		return
	}

	ctx.JSON(200, gin.H{
		"stats": stats,
	})
}

func (sc *StatsController) GetMemoryStats(ctx *gin.Context) {
	stats, err := sc.Memory.GetMemoryStats()
	if err != nil {
		ctx.JSON(500, gin.H{
			"errno": "Failed to get memory statistics",
		})
		return
	}

	ctx.JSON(200, gin.H{
		"stats": stats,
	})
}

func (sc *StatsController) GetTopUsers(ctx *gin.Context) {
	users, err := sc.Account.GetTopUsersByTotal(10)
	if err != nil {
		ctx.JSON(500, gin.H{
			"errno": "Failed to get top users",
		})
		return
	}

	ctx.JSON(200, gin.H{
		"top_users": users,
	})
}
