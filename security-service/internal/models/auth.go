package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// LoginRequest represents the login request body
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse represents the successful login response
type LoginResponse struct {
	AccessToken  string   `json:"accessToken"`
	RefreshToken string   `json:"refreshToken"`
	TokenType    string   `json:"tokenType"`
	ExpiresIn    int64    `json:"expiresIn"` // seconds until access token expires
	Roles        []string `json:"roles"`
	UserID       string   `json:"userId"`
}

// RefreshTokenRequest represents the token refresh request body
type RefreshTokenRequest struct {
	RefreshToken string `json:"refreshToken" binding:"required"`
}

// RefreshTokenResponse represents the token refresh response
type RefreshTokenResponse struct {
	AccessToken string `json:"accessToken"`
	TokenType   string `json:"tokenType"`
	ExpiresIn   int64  `json:"expiresIn"`
}

// TokenValidationResponse represents the token validation response
type TokenValidationResponse struct {
	Valid   bool     `json:"valid"`
	UserID  string   `json:"userId,omitempty"`
	Roles   []string `json:"roles,omitempty"`
	Message string   `json:"message,omitempty"`
}

// CheckPermissionRequest represents the permission check request body
type CheckPermissionRequest struct {
	UserID   string `json:"userId" binding:"required"`
	Resource string `json:"resource" binding:"required"`
	Action   string `json:"action" binding:"required"`
}

// CheckPermissionResponse represents the permission check response
type CheckPermissionResponse struct {
	Allowed bool   `json:"allowed"`
	Reason  string `json:"reason,omitempty"`
}

// AuthCredential stores encrypted credentials and tokens for external services
type AuthCredential struct {
	ID                 primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	ServiceName        string             `bson:"service_name" json:"serviceName"` // e.g., "energy_provider"
	EncryptedAPIKey    string             `bson:"encrypted_api_key" json:"-"`
	EncryptedToken     string             `bson:"encrypted_token" json:"-"`
	TokenExpiresAt     *time.Time         `bson:"token_expires_at,omitempty" json:"tokenExpiresAt,omitempty"`
	AdditionalMetadata map[string]string  `bson:"additional_metadata,omitempty" json:"additionalMetadata,omitempty"`
	CreatedAt          time.Time          `bson:"created_at" json:"createdAt"`
	UpdatedAt          time.Time          `bson:"updated_at" json:"updatedAt"`
}

// RefreshToken stores refresh tokens for session management
type RefreshToken struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	UserID    primitive.ObjectID `bson:"user_id"`
	Token     string             `bson:"token"`
	ExpiresAt time.Time          `bson:"expires_at"`
	Revoked   bool               `bson:"revoked"`
	CreatedAt time.Time          `bson:"created_at"`
}

// TokenClaims represents the JWT token claims
type TokenClaims struct {
	UserID   string   `json:"userId"`
	Username string   `json:"username"`
	Email    string   `json:"email"`
	Roles    []string `json:"roles"`
}

// UserInfoResponse represents the user info response for /auth/user-info
type UserInfoResponse struct {
	ID        string   `json:"id"`
	Username  string   `json:"username"`
	Email     string   `json:"email"`
	FirstName string   `json:"firstName"`
	LastName  string   `json:"lastName"`
	Roles     []string `json:"roles"`
}
