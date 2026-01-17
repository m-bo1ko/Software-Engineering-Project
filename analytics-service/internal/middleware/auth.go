// Package middleware provides HTTP middleware functions
package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"analytics-service/internal/integrations"
	"analytics-service/internal/models"
)

// AuthMiddleware handles JWT authentication via Security service
type AuthMiddleware struct {
	securityClient *integrations.SecurityClient
}

// NewAuthMiddleware creates a new auth middleware instance
func NewAuthMiddleware(securityClient *integrations.SecurityClient) *AuthMiddleware {
	return &AuthMiddleware{
		securityClient: securityClient,
	}
}

// RequireAuth validates the access token via Security service and sets user info in context
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

		token, err := extractTokenFromHeader(authHeader)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, models.NewErrorResponse(
				models.ErrCodeUnauthorized,
				"Invalid authorization header format",
				"Expected format: Bearer <token>",
			))
			return
		}

		// Validate token via Security service
		validationResp, err := m.securityClient.ValidateToken(c.Request.Context(), token)
		if err != nil || !validationResp.Valid {
			code := models.ErrCodeTokenInvalid
			if validationResp != nil && strings.Contains(validationResp.Message, "expired") {
				code = models.ErrCodeTokenExpired
			}
			c.AbortWithStatusJSON(http.StatusUnauthorized, models.NewErrorResponse(
				code,
				"Invalid or expired token",
				validationResp.Message,
			))
			return
		}

		// Set user info in context
		c.Set("userID", validationResp.UserID)
		c.Set("roles", validationResp.Roles)
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
	return m.RequireRoles("admin", "AnalyticsEngine")
}

// extractTokenFromHeader extracts the token from the Authorization header
func extractTokenFromHeader(authHeader string) (string, error) {
	if len(authHeader) < 7 || authHeader[:7] != "Bearer " {
		return "", fmt.Errorf("invalid authorization header format")
	}
	return authHeader[7:], nil
}

// GetUserID retrieves the user ID from context
func GetUserID(c *gin.Context) string {
	userID, exists := c.Get("userID")
	if !exists {
		return ""
	}
	if id, ok := userID.(string); ok {
		return id
	}
	return ""
}

// GetUserRoles retrieves the user roles from context
func GetUserRoles(c *gin.Context) []string {
	roles, exists := c.Get("roles")
	if !exists {
		return []string{}
	}
	if rolesList, ok := roles.([]string); ok {
		return rolesList
	}
	return []string{}
}

// GetToken retrieves the access token from context
func GetToken(c *gin.Context) string {
	token, exists := c.Get("token")
	if !exists {
		return ""
	}
	if t, ok := token.(string); ok {
		return t
	}
	return ""
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
