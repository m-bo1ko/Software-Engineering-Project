package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"security-service/internal/middleware"
	"security-service/internal/models"
	"security-service/internal/service"
)

// RoleHandler handles role management requests
type RoleHandler struct {
	roleService *service.RoleService
}

// NewRoleHandler creates a new role handler
func NewRoleHandler(roleService *service.RoleService) *RoleHandler {
	return &RoleHandler{roleService: roleService}
}

// ListRoles retrieves all roles
// GET /roles
func (h *RoleHandler) ListRoles(c *gin.Context) {
	roles, err := h.roleService.ListRoles(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrCodeInternalError,
			"Failed to retrieve roles",
			err.Error(),
		))
		return
	}

	c.JSON(http.StatusOK, models.NewSuccessResponse(gin.H{
		"roles": roles,
	}, ""))
}

// GetRole retrieves a role by name
// GET /roles/:name
func (h *RoleHandler) GetRole(c *gin.Context) {
	name := c.Param("roleName")

	role, err := h.roleService.GetRole(c.Request.Context(), name)
	if err != nil {
		if err.Error() == "role not found" {
			c.JSON(http.StatusNotFound, models.NewErrorResponse(
				models.ErrCodeNotFound,
				"Role not found",
				"",
			))
			return
		}
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrCodeInternalError,
			"Failed to retrieve role",
			err.Error(),
		))
		return
	}

	c.JSON(http.StatusOK, models.NewSuccessResponse(role, ""))
}

// CreateRole creates a new role
// POST /roles
func (h *RoleHandler) CreateRole(c *gin.Context) {
	var req models.RoleCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrCodeValidationFailed,
			"Invalid request body",
			err.Error(),
		))
		return
	}

	creatorID := middleware.GetUserID(c)

	role, err := h.roleService.CreateRole(c.Request.Context(), &req, creatorID)
	if err != nil {
		if err.Error() == "role with this name already exists" {
			c.JSON(http.StatusConflict, models.NewErrorResponse(
				models.ErrCodeConflict,
				err.Error(),
				"",
			))
			return
		}
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrCodeInternalError,
			"Failed to create role",
			err.Error(),
		))
		return
	}

	c.JSON(http.StatusCreated, models.NewSuccessResponse(role, "Role created successfully"))
}

// UpdateRole updates an existing role
// PUT /roles/:roleName
func (h *RoleHandler) UpdateRole(c *gin.Context) {
	name := c.Param("roleName")

	var req models.RoleUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrCodeValidationFailed,
			"Invalid request body",
			err.Error(),
		))
		return
	}

	updaterID := middleware.GetUserID(c)

	role, err := h.roleService.UpdateRole(c.Request.Context(), name, &req, updaterID)
	if err != nil {
		if err.Error() == "role not found" {
			c.JSON(http.StatusNotFound, models.NewErrorResponse(
				models.ErrCodeNotFound,
				"Role not found",
				"",
			))
			return
		}
		if err.Error() == "cannot modify description of system role" {
			c.JSON(http.StatusForbidden, models.NewErrorResponse(
				models.ErrCodeForbidden,
				err.Error(),
				"",
			))
			return
		}
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrCodeInternalError,
			"Failed to update role",
			err.Error(),
		))
		return
	}

	c.JSON(http.StatusOK, models.NewSuccessResponse(role, "Role updated successfully"))
}

// DeleteRole deletes a role
// DELETE /roles/:roleName
func (h *RoleHandler) DeleteRole(c *gin.Context) {
	name := c.Param("roleName")
	deleterID := middleware.GetUserID(c)

	err := h.roleService.DeleteRole(c.Request.Context(), name, deleterID)
	if err != nil {
		if err.Error() == "role not found" {
			c.JSON(http.StatusNotFound, models.NewErrorResponse(
				models.ErrCodeNotFound,
				"Role not found",
				"",
			))
			return
		}
		if err.Error() == "cannot delete system role" {
			c.JSON(http.StatusForbidden, models.NewErrorResponse(
				models.ErrCodeForbidden,
				err.Error(),
				"",
			))
			return
		}
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrCodeInternalError,
			"Failed to delete role",
			err.Error(),
		))
		return
	}

	c.JSON(http.StatusOK, models.NewSuccessResponse(nil, "Role deleted successfully"))
}
