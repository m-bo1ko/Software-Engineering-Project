package repository

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"security-service/internal/models"
)

// AuthRepository handles authentication-related database operations
type AuthRepository struct {
	refreshTokens   *mongo.Collection
	authCredentials *mongo.Collection
}

// NewAuthRepository creates a new auth repository
func NewAuthRepository(refreshTokens, authCredentials *mongo.Collection) *AuthRepository {
	return &AuthRepository{
		refreshTokens:   refreshTokens,
		authCredentials: authCredentials,
	}
}

// SaveRefreshToken stores a new refresh token
func (r *AuthRepository) SaveRefreshToken(ctx context.Context, token *models.RefreshToken) error {
	token.CreatedAt = time.Now()
	_, err := r.refreshTokens.InsertOne(ctx, token)
	return err
}

// FindRefreshToken retrieves a refresh token by its value
func (r *AuthRepository) FindRefreshToken(ctx context.Context, token string) (*models.RefreshToken, error) {
	var refreshToken models.RefreshToken
	err := r.refreshTokens.FindOne(ctx, bson.M{
		"token":   token,
		"revoked": false,
	}).Decode(&refreshToken)

	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errors.New("refresh token not found or revoked")
		}
		return nil, err
	}

	// Check if token has expired
	if refreshToken.ExpiresAt.Before(time.Now()) {
		return nil, errors.New("refresh token has expired")
	}

	return &refreshToken, nil
}

// RevokeRefreshToken marks a refresh token as revoked
func (r *AuthRepository) RevokeRefreshToken(ctx context.Context, token string) error {
	result, err := r.refreshTokens.UpdateOne(
		ctx,
		bson.M{"token": token},
		bson.M{"$set": bson.M{"revoked": true}},
	)

	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return errors.New("refresh token not found")
	}

	return nil
}

// RevokeUserTokens revokes all refresh tokens for a user
func (r *AuthRepository) RevokeUserTokens(ctx context.Context, userID string) error {
	objectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return errors.New("invalid user ID format")
	}

	_, err = r.refreshTokens.UpdateMany(
		ctx,
		bson.M{"user_id": objectID},
		bson.M{"$set": bson.M{"revoked": true}},
	)

	return err
}

// CleanupExpiredTokens removes expired refresh tokens
func (r *AuthRepository) CleanupExpiredTokens(ctx context.Context) (int64, error) {
	result, err := r.refreshTokens.DeleteMany(
		ctx,
		bson.M{"expires_at": bson.M{"$lt": time.Now()}},
	)

	if err != nil {
		return 0, err
	}

	return result.DeletedCount, nil
}

// SaveAuthCredential stores or updates credentials for an external service
func (r *AuthRepository) SaveAuthCredential(ctx context.Context, cred *models.AuthCredential) error {
	cred.UpdatedAt = time.Now()

	opts := options.Update().SetUpsert(true)
	_, err := r.authCredentials.UpdateOne(
		ctx,
		bson.M{"service_name": cred.ServiceName},
		bson.M{"$set": cred, "$setOnInsert": bson.M{"created_at": time.Now()}},
		opts,
	)

	return err
}

// FindAuthCredential retrieves credentials for an external service
func (r *AuthRepository) FindAuthCredential(ctx context.Context, serviceName string) (*models.AuthCredential, error) {
	var cred models.AuthCredential
	err := r.authCredentials.FindOne(ctx, bson.M{"service_name": serviceName}).Decode(&cred)

	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errors.New("credentials not found for service")
		}
		return nil, err
	}

	return &cred, nil
}

// UpdateAuthCredentialToken updates the token for an external service
func (r *AuthRepository) UpdateAuthCredentialToken(ctx context.Context, serviceName, encryptedToken string, expiresAt time.Time) error {
	result, err := r.authCredentials.UpdateOne(
		ctx,
		bson.M{"service_name": serviceName},
		bson.M{"$set": bson.M{
			"encrypted_token":   encryptedToken,
			"token_expires_at": expiresAt,
			"updated_at":       time.Now(),
		}},
	)

	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return errors.New("credentials not found for service")
	}

	return nil
}

// DeleteAuthCredential removes credentials for an external service
func (r *AuthRepository) DeleteAuthCredential(ctx context.Context, serviceName string) error {
	result, err := r.authCredentials.DeleteOne(ctx, bson.M{"service_name": serviceName})
	if err != nil {
		return err
	}

	if result.DeletedCount == 0 {
		return errors.New("credentials not found for service")
	}

	return nil
}

// CountActiveTokensForUser counts non-revoked tokens for a user
func (r *AuthRepository) CountActiveTokensForUser(ctx context.Context, userID string) (int64, error) {
	objectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return 0, errors.New("invalid user ID format")
	}

	return r.refreshTokens.CountDocuments(ctx, bson.M{
		"user_id":    objectID,
		"revoked":    false,
		"expires_at": bson.M{"$gt": time.Now()},
	})
}
