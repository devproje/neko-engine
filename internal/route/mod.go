package route

import (
	"github.com/devproje/neko-engine/common"
	"github.com/devproje/neko-engine/config"
	"github.com/devproje/neko-engine/middleware"
	"github.com/gin-gonic/gin"
)

func InternalRouter(app *gin.Engine, sl *common.ServiceLoader) {
	app.Use(middleware.CheckBot)

	app.GET("/", func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{
			"service": "neko-engine",
			"status":  "running",
			"info":    config.GetVersionInfo(),
		})
	})

	app.GET("/@me/:id", sl.Acc.FetchAccount)
	app.POST("/chat", sl.Chat.SendChat)
	app.POST("/register", sl.Acc.RegisterUser)

	adminGroup := app.Group("/admin")
	adminGroup.Use(middleware.RequireRoot)
	{
		adminGroup.POST("/ban", sl.Admin.BanUser)
		adminGroup.POST("/unban", sl.Admin.UnbanUser)
		adminGroup.POST("/promote", sl.Admin.PromoteToRoot)
		adminGroup.POST("/demote", sl.Admin.DemoteFromRoot)
		adminGroup.GET("/user/:user_id", sl.Admin.GetUserInfo)
		adminGroup.GET("/users", sl.Admin.ListUsers)
		adminGroup.GET("/users/search", sl.Admin.SearchUsers)
		adminGroup.POST("/reset-count", sl.Admin.ResetUserCount)
	}

	app.GET("/memories", sl.Memory.ListMemories)
	app.GET("/memories/search", sl.Memory.SearchMemories)
	app.GET("/memories/:memory_id", sl.Memory.GetMemory)
	app.PUT("/memories/:memory_id", sl.Memory.UpdateMemory)
	app.DELETE("/memories/:memory_id", sl.Memory.DeleteMemory)
	app.POST("/memories/:memory_id/reanalyze", sl.Memory.ReanalyzeMemory)
	app.DELETE("/memories/user/:user_id", sl.Memory.FlushUserMemories)

	app.DELETE("/history/user/:user_id", sl.Memory.FlushUserHistory)

	roleGroup := app.Group("/roles")
	roleGroup.Use(middleware.RequireRoot)
	{
		roleGroup.GET("", sl.Role.ListRoles)
		roleGroup.GET("/:role_id", sl.Role.GetRole)
		roleGroup.POST("", sl.Role.CreateRole)
		roleGroup.PUT("/:role_id", sl.Role.UpdateRole)
		roleGroup.DELETE("/:role_id", sl.Role.DeleteRole)
	}

	app.GET("/stats", sl.Stats.GetSystemStats)
	app.GET("/stats/users", sl.Stats.GetUserStats)
	app.GET("/stats/memories", sl.Stats.GetMemoryStats)
	app.GET("/stats/top-users", sl.Stats.GetTopUsers)
}
