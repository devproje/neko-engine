package controller

import (
	"strconv"

	"github.com/devproje/neko-engine/common/repository"
	"github.com/devproje/neko-engine/common/service"
	"github.com/gin-gonic/gin"
)

type MemoryController struct {
	Memory *service.MemoryService
}

type MemorySearchRequest struct {
	Query  string `json:"query" form:"query"`
	Limit  int    `json:"limit" form:"limit"`
	Offset int    `json:"offset" form:"offset"`
}

type MemoryUpdateRequest struct {
	Summary    string  `json:"summary"`
	Keywords   string  `json:"keywords"`
	Importance float64 `json:"importance"`
}

func NewMemoryController(memory *service.MemoryService) *MemoryController {
	return &MemoryController{Memory: memory}
}

func (mc *MemoryController) ListMemories(ctx *gin.Context) {
	userID := ctx.Query("user_id")
	limitStr := ctx.DefaultQuery("limit", "20")
	offsetStr := ctx.DefaultQuery("offset", "0")

	limit, _ := strconv.Atoi(limitStr)
	offset, _ := strconv.Atoi(offsetStr)

	if limit <= 0 || limit > 100 {
		limit = 20
	}

	var memories []*repository.Memory
	var err error

	if userID != "" {
		memories, err = mc.Memory.LoadMemories(userID, limit)
	} else {
		memories, err = mc.Memory.LoadAllMemories(limit, offset)
	}

	if err != nil {
		ctx.JSON(500, gin.H{
			"errno": "Failed to load memories",
		})
		return
	}

	ctx.JSON(200, gin.H{
		"memories": memories,
		"count":    len(memories),
	})
}

func (mc *MemoryController) SearchMemories(ctx *gin.Context) {
	var req MemorySearchRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(400, gin.H{
			"errno": "Invalid request parameters",
		})
		return
	}

	if req.Query == "" {
		ctx.JSON(400, gin.H{
			"errno": "Search query is required",
		})
		return
	}

	if req.Limit <= 0 || req.Limit > 100 {
		req.Limit = 20
	}

	userID := ctx.Query("user_id")
	memories, err := mc.Memory.SearchMemoriesByKeywords(userID, req.Query, req.Limit)
	if err != nil {
		ctx.JSON(500, gin.H{
			"errno": "Failed to search memories",
		})
		return
	}

	ctx.JSON(200, gin.H{
		"memories": memories,
		"count":    len(memories),
		"query":    req.Query,
	})
}

func (mc *MemoryController) GetMemory(ctx *gin.Context) {
	memoryIDStr := ctx.Param("memory_id")
	memoryID, err := strconv.ParseUint(memoryIDStr, 10, 32)
	if err != nil {
		ctx.JSON(400, gin.H{
			"errno": "Invalid memory ID",
		})
		return
	}

	memory, err := mc.Memory.GetMemoryByID(uint(memoryID))
	if err != nil {
		ctx.JSON(404, gin.H{
			"errno": "Memory not found",
		})
		return
	}

	ctx.JSON(200, gin.H{
		"memory": memory,
	})
}

func (mc *MemoryController) UpdateMemory(ctx *gin.Context) {
	memoryIDStr := ctx.Param("memory_id")
	memoryID, err := strconv.ParseUint(memoryIDStr, 10, 32)
	if err != nil {
		ctx.JSON(400, gin.H{
			"errno": "Invalid memory ID",
		})
		return
	}

	var req MemoryUpdateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(400, gin.H{
			"errno": "Invalid request format",
		})
		return
	}

	memory, err := mc.Memory.GetMemoryByID(uint(memoryID))
	if err != nil {
		ctx.JSON(404, gin.H{
			"errno": "Memory not found",
		})
		return
	}

	if req.Summary != "" {
		memory.Summary = req.Summary
	}
	if req.Keywords != "" {
		memory.Keywords = req.Keywords
	}
	if req.Importance >= 0.0 && req.Importance <= 1.0 {
		memory.Importance = req.Importance
	}

	if err := mc.Memory.UpdateMemory(memory); err != nil {
		ctx.JSON(500, gin.H{
			"errno": "Failed to update memory",
		})
		return
	}

	ctx.JSON(200, gin.H{
		"message": "Memory updated successfully",
		"memory":  memory,
	})
}

func (mc *MemoryController) DeleteMemory(ctx *gin.Context) {
	memoryIDStr := ctx.Param("memory_id")
	memoryID, err := strconv.ParseUint(memoryIDStr, 10, 32)
	if err != nil {
		ctx.JSON(400, gin.H{
			"errno": "Invalid memory ID",
		})
		return
	}

	if err := mc.Memory.DeleteMemory(uint(memoryID)); err != nil {
		ctx.JSON(404, gin.H{
			"errno": "Memory not found or failed to delete",
		})
		return
	}

	ctx.JSON(200, gin.H{
		"message": "Memory deleted successfully",
	})
}

func (mc *MemoryController) ReanalyzeMemory(ctx *gin.Context) {
	memoryIDStr := ctx.Param("memory_id")
	memoryID, err := strconv.ParseUint(memoryIDStr, 10, 32)
	if err != nil {
		ctx.JSON(400, gin.H{
			"errno": "Invalid memory ID",
		})
		return
	}

	if err := mc.Memory.ReanalyzeAndUpdateMemory(uint(memoryID)); err != nil {
		ctx.JSON(500, gin.H{
			"errno": "Failed to reanalyze memory: " + err.Error(),
		})
		return
	}

	ctx.JSON(200, gin.H{
		"message": "Memory reanalyzed successfully",
	})
}

func (mc *MemoryController) FlushUserMemories(ctx *gin.Context) {
	userID := ctx.Param("user_id")
	if userID == "" {
		ctx.JSON(400, gin.H{
			"errno": "User ID is required",
		})
		return
	}

	if err := mc.Memory.FlushUserMemories(userID); err != nil {
		ctx.JSON(500, gin.H{
			"errno": "Failed to flush user memories",
		})
		return
	}

	ctx.JSON(200, gin.H{
		"message": "User memories flushed successfully",
	})
}

func (mc *MemoryController) FlushUserHistory(ctx *gin.Context) {
	userID := ctx.Param("user_id")
	if userID == "" {
		ctx.JSON(400, gin.H{
			"errno": "User ID is required",
		})
		return
	}

	if err := mc.Memory.FlushHistory(userID); err != nil {
		ctx.JSON(500, gin.H{
			"errno": "Failed to flush user history",
		})
		return
	}

	ctx.JSON(200, gin.H{
		"message": "User history flushed successfully",
	})
}
