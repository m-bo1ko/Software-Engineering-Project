package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"security-service/internal/models"
	"security-service/pkg/utils"
)

// TestLoginRequest tests the login request model
func TestLoginRequest(t *testing.T) {
	tests := []struct {
		name     string
		username string
		password string
	}{
		{"valid login", "testuser", "password123"},
		{"admin login", "admin", "adminpass"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := models.LoginRequest{
				Username: tt.username,
				Password: tt.password,
			}
			assert.Equal(t, tt.username, req.Username)
			assert.Equal(t, tt.password, req.Password)
		})
	}
}

// TestJWTManager tests JWT token generation and validation
func TestJWTManager(t *testing.T) {
	jwtManager := utils.NewJWTManager("test-secret-key", 15*time.Minute, 7*24*time.Hour)

	t.Run("Generate and validate access token", func(t *testing.T) {
		user := &models.User{
			Username: "testuser",
			Email:    "test@example.com",
			Roles:    []string{"user"},
		}
		user.ID = [12]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12}

		token, err := jwtManager.GenerateAccessToken(user)
		require.NoError(t, err)
		assert.NotEmpty(t, token)

		claims, err := jwtManager.ValidateAccessToken(token)
		require.NoError(t, err)
		assert.Equal(t, user.Username, claims.Username)
		assert.Equal(t, user.Email, claims.Email)
		assert.Equal(t, user.Roles, claims.Roles)
	})

	t.Run("Generate and validate refresh token", func(t *testing.T) {
		userID := "507f1f77bcf86cd799439011"

		token, expiresAt, err := jwtManager.GenerateRefreshToken(userID)
		require.NoError(t, err)
		assert.NotEmpty(t, token)
		assert.True(t, expiresAt.After(time.Now()))

		extractedUserID, err := jwtManager.ValidateRefreshToken(token)
		require.NoError(t, err)
		assert.Equal(t, userID, extractedUserID)
	})

	t.Run("Invalid token validation", func(t *testing.T) {
		_, err := jwtManager.ValidateAccessToken("invalid-token")
		assert.Error(t, err)
	})

	t.Run("Extract token from header", func(t *testing.T) {
		validHeader := "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.test"
		token, err := utils.ExtractTokenFromHeader(validHeader)
		require.NoError(t, err)
		assert.Equal(t, "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.test", token)

		invalidHeader := "InvalidHeader"
		_, err = utils.ExtractTokenFromHeader(invalidHeader)
		assert.Error(t, err)
	})
}

// TestPasswordHashing tests password hashing and verification
func TestPasswordHashing(t *testing.T) {
	t.Run("Hash and verify password", func(t *testing.T) {
		password := "testPassword123!"

		hash, err := utils.HashPassword(password)
		require.NoError(t, err)
		assert.NotEmpty(t, hash)
		assert.NotEqual(t, password, hash)

		// Verify correct password
		assert.True(t, utils.CheckPassword(password, hash))

		// Verify incorrect password
		assert.False(t, utils.CheckPassword("wrongpassword", hash))
	})

	t.Run("Different passwords produce different hashes", func(t *testing.T) {
		hash1, _ := utils.HashPassword("password1")
		hash2, _ := utils.HashPassword("password2")
		assert.NotEqual(t, hash1, hash2)
	})
}

// TestEncryption tests encryption and decryption
func TestEncryption(t *testing.T) {
	encryptor, err := utils.NewEncryptor("32-byte-encryption-key-here!!!!")
	require.NoError(t, err)

	t.Run("Encrypt and decrypt", func(t *testing.T) {
		plaintext := "sensitive data to encrypt"

		ciphertext, err := encryptor.Encrypt(plaintext)
		require.NoError(t, err)
		assert.NotEmpty(t, ciphertext)
		assert.NotEqual(t, plaintext, ciphertext)

		decrypted, err := encryptor.Decrypt(ciphertext)
		require.NoError(t, err)
		assert.Equal(t, plaintext, decrypted)
	})

	t.Run("Different plaintexts produce different ciphertexts", func(t *testing.T) {
		ct1, _ := encryptor.Encrypt("text1")
		ct2, _ := encryptor.Encrypt("text2")
		assert.NotEqual(t, ct1, ct2)
	})
}

// TestLoginHandler tests the login endpoint handler
func TestLoginHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Missing request body", func(t *testing.T) {
		router := gin.New()
		router.POST("/auth/login", func(c *gin.Context) {
			var req models.LoginRequest
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, models.NewErrorResponse(
					models.ErrCodeValidationFailed,
					"Invalid request body",
					err.Error(),
				))
				return
			}
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/auth/login", nil)
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Valid request body format", func(t *testing.T) {
		router := gin.New()
		router.POST("/auth/login", func(c *gin.Context) {
			var req models.LoginRequest
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, models.NewErrorResponse(
					models.ErrCodeValidationFailed,
					"Invalid request body",
					err.Error(),
				))
				return
			}
			c.JSON(http.StatusOK, gin.H{"received": true})
		})

		body := `{"username": "testuser", "password": "testpass"}`
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/auth/login", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

// TestTokenValidationResponse tests token validation response model
func TestTokenValidationResponse(t *testing.T) {
	t.Run("Valid token response", func(t *testing.T) {
		resp := models.TokenValidationResponse{
			Valid:  true,
			UserID: "user123",
			Roles:  []string{"admin", "user"},
		}

		assert.True(t, resp.Valid)
		assert.Equal(t, "user123", resp.UserID)
		assert.Contains(t, resp.Roles, "admin")
	})

	t.Run("Invalid token response", func(t *testing.T) {
		resp := models.TokenValidationResponse{
			Valid:   false,
			Message: "token expired",
		}

		assert.False(t, resp.Valid)
		assert.Equal(t, "token expired", resp.Message)
	})
}

// TestLoginResponse tests login response structure
func TestLoginResponse(t *testing.T) {
	resp := models.LoginResponse{
		AccessToken:  "access.token.here",
		RefreshToken: "refresh.token.here",
		TokenType:    "Bearer",
		ExpiresIn:    900,
		Roles:        []string{"user"},
		UserID:       "user123",
	}

	jsonData, err := json.Marshal(resp)
	require.NoError(t, err)

	var decoded models.LoginResponse
	err = json.Unmarshal(jsonData, &decoded)
	require.NoError(t, err)

	assert.Equal(t, resp.AccessToken, decoded.AccessToken)
	assert.Equal(t, resp.RefreshToken, decoded.RefreshToken)
	assert.Equal(t, resp.TokenType, decoded.TokenType)
	assert.Equal(t, resp.ExpiresIn, decoded.ExpiresIn)
	assert.Equal(t, resp.Roles, decoded.Roles)
	assert.Equal(t, resp.UserID, decoded.UserID)
}

// TestCheckPermissionRequest tests permission check request
func TestCheckPermissionRequest(t *testing.T) {
	req := models.CheckPermissionRequest{
		UserID:   "user123",
		Resource: "buildings",
		Action:   "read",
	}

	assert.Equal(t, "user123", req.UserID)
	assert.Equal(t, "buildings", req.Resource)
	assert.Equal(t, "read", req.Action)
}
