// Package service contains business logic for the application
package service

import (
	"context"
	"errors"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"security-service/internal/models"
	"security-service/internal/repository"
	"security-service/pkg/utils"
)

// AuthService handles authentication business logic
type AuthService struct {
	userRepo   *repository.UserRepository
	roleRepo   *repository.RoleRepository
	authRepo   *repository.AuthRepository
	auditRepo  *repository.AuditRepository
	jwtManager *utils.JWTManager
}

// NewAuthService creates a new authentication service
func NewAuthService(
	userRepo *repository.UserRepository,
	roleRepo *repository.RoleRepository,
	authRepo *repository.AuthRepository,
	auditRepo *repository.AuditRepository,
	jwtManager *utils.JWTManager,
) *AuthService {
	return &AuthService{
		userRepo:   userRepo,
		roleRepo:   roleRepo,
		authRepo:   authRepo,
		auditRepo:  auditRepo,
		jwtManager: jwtManager,
	}
}

// Login authenticates a user and returns tokens
func (s *AuthService) Login(ctx context.Context, req *models.LoginRequest, ipAddress, userAgent string) (*models.LoginResponse, error) {
	// Find user by username
	user, err := s.userRepo.FindByUsername(ctx, req.Username)
	if err != nil {
		s.logAuditEvent(ctx, "", req.Username, "LOGIN", "auth", "FAILURE", "Invalid credentials", ipAddress, userAgent)
		return nil, errors.New("invalid username or password")
	}

	// Check if user is active
	if !user.IsActive {
		s.logAuditEvent(ctx, user.ID.Hex(), user.Username, "LOGIN", "auth", "FAILURE", "Account is disabled", ipAddress, userAgent)
		return nil, errors.New("account is disabled")
	}

	// Verify password
	if !utils.CheckPassword(req.Password, user.PasswordHash) {
		s.logAuditEvent(ctx, user.ID.Hex(), user.Username, "LOGIN", "auth", "FAILURE", "Invalid credentials", ipAddress, userAgent)
		return nil, errors.New("invalid username or password")
	}

	// Generate access token
	accessToken, err := s.jwtManager.GenerateAccessToken(user)
	if err != nil {
		return nil, errors.New("failed to generate access token")
	}

	// Generate refresh token
	refreshTokenString, expiresAt, err := s.jwtManager.GenerateRefreshToken(user.ID.Hex())
	if err != nil {
		return nil, errors.New("failed to generate refresh token")
	}

	// Store refresh token
	refreshToken := &models.RefreshToken{
		UserID:    user.ID,
		Token:     refreshTokenString,
		ExpiresAt: expiresAt,
		Revoked:   false,
	}

	if err := s.authRepo.SaveRefreshToken(ctx, refreshToken); err != nil {
		return nil, errors.New("failed to save refresh token")
	}

	// Update last login time
	if err := s.userRepo.UpdateLastLogin(ctx, user.ID.Hex()); err != nil {
		log.Printf("Failed to update last login time: %v", err)
	}

	// Log successful login
	s.logAuditEvent(ctx, user.ID.Hex(), user.Username, "LOGIN", "auth", "SUCCESS", "", ipAddress, userAgent)

	return &models.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshTokenString,
		TokenType:    "Bearer",
		ExpiresIn:    int64(s.jwtManager.GetAccessTokenExpiry().Seconds()),
		Roles:        user.Roles,
		UserID:       user.ID.Hex(),
	}, nil
}

// RefreshToken refreshes the access token using a refresh token
func (s *AuthService) RefreshToken(ctx context.Context, req *models.RefreshTokenRequest) (*models.RefreshTokenResponse, error) {
	// Validate the refresh token format
	userID, err := s.jwtManager.ValidateRefreshToken(req.RefreshToken)
	if err != nil {
		return nil, err
	}

	// Check if refresh token exists and is not revoked
	storedToken, err := s.authRepo.FindRefreshToken(ctx, req.RefreshToken)
	if err != nil {
		return nil, err
	}

	// Verify user ID matches
	if storedToken.UserID.Hex() != userID {
		return nil, errors.New("token mismatch")
	}

	// Get user for new access token
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	if !user.IsActive {
		return nil, errors.New("account is disabled")
	}

	// Generate new access token
	accessToken, err := s.jwtManager.GenerateAccessToken(user)
	if err != nil {
		return nil, errors.New("failed to generate access token")
	}

	return &models.RefreshTokenResponse{
		AccessToken: accessToken,
		TokenType:   "Bearer",
		ExpiresIn:   int64(s.jwtManager.GetAccessTokenExpiry().Seconds()),
	}, nil
}

// Logout revokes the user's refresh tokens
func (s *AuthService) Logout(ctx context.Context, refreshToken, userID, ipAddress, userAgent string) error {
	// Revoke the specific refresh token
	if err := s.authRepo.RevokeRefreshToken(ctx, refreshToken); err != nil {
		log.Printf("Failed to revoke refresh token: %v", err)
	}

	// Log logout event
	s.logAuditEvent(ctx, userID, "", "LOGOUT", "auth", "SUCCESS", "", ipAddress, userAgent)

	return nil
}

// ValidateToken validates an access token and returns user info
func (s *AuthService) ValidateToken(ctx context.Context, token string) (*models.TokenValidationResponse, error) {
	claims, err := s.jwtManager.ValidateAccessToken(token)
	if err != nil {
		return &models.TokenValidationResponse{
			Valid:   false,
			Message: err.Error(),
		}, nil
	}

	// Verify user still exists and is active
	user, err := s.userRepo.FindByID(ctx, claims.UserID)
	if err != nil {
		return &models.TokenValidationResponse{
			Valid:   false,
			Message: "user not found",
		}, nil
	}

	if !user.IsActive {
		return &models.TokenValidationResponse{
			Valid:   false,
			Message: "account is disabled",
		}, nil
	}

	return &models.TokenValidationResponse{
		Valid:  true,
		UserID: claims.UserID,
		Roles:  claims.Roles,
	}, nil
}

// CheckPermission checks if a user has permission for a specific action on a resource
func (s *AuthService) CheckPermission(ctx context.Context, req *models.CheckPermissionRequest) (*models.CheckPermissionResponse, error) {
	// Get user
	user, err := s.userRepo.FindByID(ctx, req.UserID)
	if err != nil {
		return &models.CheckPermissionResponse{
			Allowed: false,
			Reason:  "user not found",
		}, nil
	}

	if !user.IsActive {
		return &models.CheckPermissionResponse{
			Allowed: false,
			Reason:  "account is disabled",
		}, nil
	}

	// Get user's roles
	roles, err := s.roleRepo.FindByNames(ctx, user.Roles)
	if err != nil {
		return &models.CheckPermissionResponse{
			Allowed: false,
			Reason:  "failed to retrieve roles",
		}, nil
	}

	// Check if any role has the required permission
	for _, role := range roles {
		if role.HasPermission(req.Resource, req.Action) {
			return &models.CheckPermissionResponse{
				Allowed: true,
			}, nil
		}
	}

	return &models.CheckPermissionResponse{
		Allowed: false,
		Reason:  "insufficient permissions",
	}, nil
}

// GetUserInfo returns user profile information based on the access token
func (s *AuthService) GetUserInfo(ctx context.Context, token string) (*models.UserInfoResponse, error) {
	claims, err := s.jwtManager.ValidateAccessToken(token)
	if err != nil {
		return nil, err
	}

	user, err := s.userRepo.FindByID(ctx, claims.UserID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	return &models.UserInfoResponse{
		ID:        user.ID.Hex(),
		Username:  user.Username,
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Roles:     user.Roles,
	}, nil
}

// logAuditEvent logs an authentication-related audit event
func (s *AuthService) logAuditEvent(ctx context.Context, userID, username, action, resource, status, errorMsg, ipAddress, userAgent string) {
	// Переименовали переменную с log на auditLog
	auditLog := &models.AuditLog{
		ID:        primitive.NewObjectID(),
		UserID:    userID,
		Username:  username,
		Service:   "security-service",
		Action:    action,
		Resource:  resource,
		Status:    status,
		ErrorMsg:  errorMsg,
		IPAddress: ipAddress,
		UserAgent: userAgent,
		Timestamp: time.Now(),
	}

	if _, err := s.auditRepo.Create(ctx, auditLog); err != nil {
		// Теперь log.Printf обращается к стандартному пакету log, а не к структуре
		log.Printf("Failed to create audit log: %v", err)
	}
}
