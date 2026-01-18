package handlers

import (
	"net/http"
	"regexp"
	"strconv"

	"github.com/gin-gonic/gin"

	"security-service/internal/middleware"
	"security-service/internal/models"
	"security-service/internal/service"
)

// objectIDRegex validates MongoDB ObjectID format (24 hex characters)
var objectIDRegex = regexp.MustCompile(`^[a-fA-F0-9]{24}$`)

// UserHandler handles user management requests
type UserHandler struct {
	userService *service.UserService
}

// NewUserHandler creates a new user handler
func NewUserHandler(userService *service.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

// ListUsers retrieves all users
// GET /users
func (h *UserHandler) ListUsers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	users, total, totalPages, err := h.userService.ListUsers(c.Request.Context(), page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrCodeInternalError,
			"Failed to retrieve users",
			err.Error(),
		))
		return
	}

	c.JSON(http.StatusOK, models.NewSuccessResponse(gin.H{
		"users":      users,
		"total":      total,
		"page":       page,
		"limit":      limit,
		"totalPages": totalPages,
	}, ""))
}

// GetUser retrieves a user by ID
// GET /users/:id
func (h *UserHandler) GetUser(c *gin.Context) {
	id := c.Param("id")

	// Validate ID format
	if !objectIDRegex.MatchString(id) {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrCodeValidationFailed,
			"Invalid user ID format",
			"ID must be a valid 24-character hex string",
		))
		return
	}

	user, err := h.userService.GetUser(c.Request.Context(), id)
	if err != nil {
		if err.Error() == "user not found" {
			c.JSON(http.StatusNotFound, models.NewErrorResponse(
				models.ErrCodeNotFound,
				"User not found",
				"",
			))
			return
		}
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrCodeInternalError,
			"Failed to retrieve user",
			err.Error(),
		))
		return
	}

	c.JSON(http.StatusOK, models.NewSuccessResponse(user, ""))
}

// CreateUser creates a new user
// POST /users
func (h *UserHandler) CreateUser(c *gin.Context) {
	var req models.UserCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrCodeValidationFailed,
			"Invalid request body",
			err.Error(),
		))
		return
	}

	creatorID := middleware.GetUserID(c)

	user, err := h.userService.CreateUser(c.Request.Context(), &req, creatorID)
	if err != nil {
		if err.Error() == "user with this username or email already exists" {
			c.JSON(http.StatusConflict, models.NewErrorResponse(
				models.ErrCodeConflict,
				err.Error(),
				"",
			))
			return
		}
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrCodeInternalError,
			"Failed to create user",
			err.Error(),
		))
		return
	}

	c.JSON(http.StatusCreated, models.NewSuccessResponse(user, "User created successfully"))
}

// UpdateUser updates an existing user
// PUT /users/:id
func (h *UserHandler) UpdateUser(c *gin.Context) {
	id := c.Param("id")

	var req models.UserUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrCodeValidationFailed,
			"Invalid request body",
			err.Error(),
		))
		return
	}

	updaterID := middleware.GetUserID(c)

	user, err := h.userService.UpdateUser(c.Request.Context(), id, &req, updaterID)
	if err != nil {
		if err.Error() == "user not found" {
			c.JSON(http.StatusNotFound, models.NewErrorResponse(
				models.ErrCodeNotFound,
				"User not found",
				"",
			))
			return
		}
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrCodeInternalError,
			"Failed to update user",
			err.Error(),
		))
		return
	}

	c.JSON(http.StatusOK, models.NewSuccessResponse(user, "User updated successfully"))
}

// DeleteUser deletes a user
// DELETE /users/:id
func (h *UserHandler) DeleteUser(c *gin.Context) {
	id := c.Param("id")
	deleterID := middleware.GetUserID(c)

	err := h.userService.DeleteUser(c.Request.Context(), id, deleterID)
	if err != nil {
		if err.Error() == "user not found" {
			c.JSON(http.StatusNotFound, models.NewErrorResponse(
				models.ErrCodeNotFound,
				"User not found",
				"",
			))
			return
		}
		if err.Error() == "cannot delete your own account" {
			c.JSON(http.StatusBadRequest, models.NewErrorResponse(
				models.ErrCodeInvalidRequest,
				err.Error(),
				"",
			))
			return
		}
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrCodeInternalError,
			"Failed to delete user",
			err.Error(),
		))
		return
	}

	c.JSON(http.StatusOK, models.NewSuccessResponse(nil, "User deleted successfully"))
}
