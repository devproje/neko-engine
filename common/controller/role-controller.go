package controller

import (
	"strconv"

	"github.com/devproje/neko-engine/common/repository"
	"github.com/devproje/neko-engine/common/service"
	"github.com/gin-gonic/gin"
)

type RoleController struct {
	Account *service.AccountService
}

type RoleRequest struct {
	Name  string `json:"name" binding:"required"`
	Limit int    `json:"limit" binding:"required"`
	Root  bool   `json:"root"`
}

func NewRoleController(account *service.AccountService) *RoleController {
	return &RoleController{Account: account}
}

func (rc *RoleController) ListRoles(ctx *gin.Context) {
	roles, err := rc.Account.ListRoles()
	if err != nil {
		ctx.JSON(500, gin.H{
			"errno": "Failed to load roles",
		})
		return
	}

	ctx.JSON(200, gin.H{
		"roles": roles,
		"count": len(roles),
	})
}

func (rc *RoleController) GetRole(ctx *gin.Context) {
	roleIDStr := ctx.Param("role_id")
	roleID, err := strconv.Atoi(roleIDStr)
	if err != nil {
		ctx.JSON(400, gin.H{
			"errno": "Invalid role ID",
		})
		return
	}

	role, err := rc.Account.GetRoleById(roleID)
	if err != nil {
		ctx.JSON(404, gin.H{
			"errno": "Role not found",
		})
		return
	}

	ctx.JSON(200, gin.H{
		"role": role,
	})
}

func (rc *RoleController) CreateRole(ctx *gin.Context) {
	var req RoleRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(400, gin.H{
			"errno": "Invalid request format",
		})
		return
	}

	role := &repository.Role{
		Name:  req.Name,
		Limit: req.Limit,
		Root:  req.Root,
	}

	if err := rc.Account.CreateRole(role); err != nil {
		ctx.JSON(400, gin.H{
			"errno": "Failed to create role: " + err.Error(),
		})
		return
	}

	ctx.JSON(201, gin.H{
		"message": "Role created successfully",
		"role":    role,
	})
}

func (rc *RoleController) UpdateRole(ctx *gin.Context) {
	roleIDStr := ctx.Param("role_id")
	roleID, err := strconv.Atoi(roleIDStr)
	if err != nil {
		ctx.JSON(400, gin.H{
			"errno": "Invalid role ID",
		})
		return
	}

	var req RoleRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(400, gin.H{
			"errno": "Invalid request format",
		})
		return
	}

	role, err := rc.Account.GetRoleById(roleID)
	if err != nil {
		ctx.JSON(404, gin.H{
			"errno": "Role not found",
		})
		return
	}

	role.Name = req.Name
	role.Limit = req.Limit
	role.Root = req.Root

	if err := rc.Account.UpdateRole(role); err != nil {
		ctx.JSON(400, gin.H{
			"errno": "Failed to update role: " + err.Error(),
		})
		return
	}

	ctx.JSON(200, gin.H{
		"message": "Role updated successfully",
		"role":    role,
	})
}

func (rc *RoleController) DeleteRole(ctx *gin.Context) {
	roleIDStr := ctx.Param("role_id")
	roleID, err := strconv.Atoi(roleIDStr)
	if err != nil {
		ctx.JSON(400, gin.H{
			"errno": "Invalid role ID",
		})
		return
	}

	if roleID <= 3 {
		ctx.JSON(400, gin.H{
			"errno": "Cannot delete system roles (root, user, server)",
		})
		return
	}

	if err := rc.Account.DeleteRole(roleID); err != nil {
		ctx.JSON(400, gin.H{
			"errno": "Failed to delete role: " + err.Error(),
		})
		return
	}

	ctx.JSON(200, gin.H{
		"message": "Role deleted successfully",
	})
}
