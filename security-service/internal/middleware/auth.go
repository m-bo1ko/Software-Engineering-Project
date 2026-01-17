// Package middleware provides HTTP middleware functions
package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"security-service/internal/models"
	"security-service/pkg/utils"
)

// AuthMiddleware creates a new authentication middleware
type AuthMiddleware struct {
	jwtManager *utils.JWTManager
}

// NewAuthMiddleware creates a new auth middleware instance
func NewAuthMiddleware(jwtManager *utils.JWTManager) *AuthMiddleware {
	return &AuthMiddleware{jwtManager: jwtManager}
}

// RequireAuth validates the access token and sets user info in context
func (m *AuthMiddleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, models.NewErrorResponse(
				models.ErrCodeUnauthorized,
				"Authorization header is required",
				"",
			))
			return
		}

		token, err := utils.ExtractTokenFromHeader(authHeader)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, models.NewErrorResponse(
				models.ErrCodeUnauthorized,
				"Invalid authorization header format",
				"Expected format: Bearer <token>",
			))
			return
		}

		claims, err := m.jwtManager.ValidateAccessToken(token)
		if err != nil {
			code := models.ErrCodeTokenInvalid
			if strings.Contains(err.Error(), "expired") {
				code = models.ErrCodeTokenExpired
			}
			c.AbortWithStatusJSON(http.StatusUnauthorized, models.NewErrorResponse(
				code,
				err.Error(),
				"",
			))
			return
		}

		// Set user info in context
		c.Set("userID", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("email", claims.Email)
		c.Set("roles", claims.Roles)
		c.Set("token", token)

		c.Next()
	}
}

// RequireRoles checks if the user has any of the specified roles
func (m *AuthMiddleware) RequireRoles(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRoles, exists := c.Get("roles")
		if !exists {
			c.AbortWithStatusJSON(http.StatusForbidden, models.NewErrorResponse(
				models.ErrCodeForbidden,
				"User roles not found",
				"",
			))
			return
		}

		userRolesList, ok := userRoles.([]string)
		if !ok {
			c.AbortWithStatusJSON(http.StatusInternalServerError, models.NewErrorResponse(
				models.ErrCodeInternalError,
				"Invalid roles format",
				"",
			))
			return
		}

		// Check if user has any of the required roles
		hasRole := false
		for _, requiredRole := range roles {
			for _, userRole := range userRolesList {
				if userRole == requiredRole || userRole == "admin" {
					hasRole = true
					break
				}
			}
			if hasRole {
				break
			}
		}

		if !hasRole {
			c.AbortWithStatusJSON(http.StatusForbidden, models.NewErrorResponse(
				models.ErrCodeForbidden,
				"Insufficient permissions",
				"Required roles: "+strings.Join(roles, ", "),
			))
			return
		}

		c.Next()
	}
}

// RequireAdmin is a convenience method that requires the admin role
func (m *AuthMiddleware) RequireAdmin() gin.HandlerFunc {
	return m.RequireRoles("admin")
}

// OptionalAuth validates the token if present but doesn't require it
func (m *AuthMiddleware) OptionalAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Next()
			return
		}

		token, err := utils.ExtractTokenFromHeader(authHeader)
		if err != nil {
			c.Next()
			return
		}

		claims, err := m.jwtManager.ValidateAccessToken(token)
		if err != nil {
			c.Next()
			return
		}

		// Set user info in context
		c.Set("userID", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("email", claims.Email)
		c.Set("roles", claims.Roles)
		c.Set("token", token)

		c.Next()
	}
}

// GetUserID retrieves the user ID from context
func GetUserID(c *gin.Context) string {
	userID, exists := c.Get("userID")
	if !exists {
		return ""
	}
	return userID.(string)
}

// GetUsername retrieves the username from context
func GetUsername(c *gin.Context) string {
	username, exists := c.Get("username")
	if !exists {
		return ""
	}
	return username.(string)
}

// GetUserRoles retrieves the user roles from context
func GetUserRoles(c *gin.Context) []string {
	roles, exists := c.Get("roles")
	if !exists {
		return []string{}
	}
	return roles.([]string)
}

// GetToken retrieves the access token from context
func GetToken(c *gin.Context) string {
	token, exists := c.Get("token")
	if !exists {
		return ""
	}
	return token.(string)
}

// HasRole checks if the user has a specific role
func HasRole(c *gin.Context, role string) bool {
	roles := GetUserRoles(c)
	for _, r := range roles {
		if r == role || r == "admin" {
			return true
		}
	}
	return false
}
