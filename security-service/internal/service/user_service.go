package service

import (
	"context"
	"errors"
	"math"
	"time"

	"go.mongodb.org/mongo-driver/bson"

	"security-service/internal/models"
	"security-service/internal/repository"
	"security-service/pkg/utils"
)

// UserService handles user management business logic
type UserService struct {
	userRepo  *repository.UserRepository
	roleRepo  *repository.RoleRepository
	auditRepo *repository.AuditRepository
}

// NewUserService creates a new user service
func NewUserService(
	userRepo *repository.UserRepository,
	roleRepo *repository.RoleRepository,
	auditRepo *repository.AuditRepository,
) *UserService {
	return &UserService{
		userRepo:  userRepo,
		roleRepo:  roleRepo,
		auditRepo: auditRepo,
	}
}

// CreateUser creates a new user
func (s *UserService) CreateUser(ctx context.Context, req *models.UserCreateRequest, creatorID string) (*models.UserResponse, error) {
	// Validate roles exist
	if len(req.Roles) > 0 {
		roles, err := s.roleRepo.FindByNames(ctx, req.Roles)
		if err != nil {
			return nil, err
		}
		if len(roles) != len(req.Roles) {
			return nil, errors.New("one or more roles do not exist")
		}
	}

	// Hash password
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, errors.New("failed to hash password")
	}

	// Create user
	user := &models.User{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: hashedPassword,
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		Roles:        req.Roles,
		IsActive:     true,
	}

	if user.Roles == nil {
		user.Roles = []string{"user"} // Default role
	}

	createdUser, err := s.userRepo.Create(ctx, user)
	if err != nil {
		return nil, err
	}

	// Log audit event
	s.logAuditEvent(ctx, creatorID, "CREATE_USER", "user", createdUser.ID.Hex(), "SUCCESS", "")

	return createdUser.ToResponse(), nil
}

// GetUser retrieves a user by ID
func (s *UserService) GetUser(ctx context.Context, id string) (*models.UserResponse, error) {
	user, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return user.ToResponse(), nil
}

// ListUsers retrieves all users with pagination
func (s *UserService) ListUsers(ctx context.Context, page, limit int) ([]*models.UserResponse, int64, int, error) {
	users, total, err := s.userRepo.FindAll(ctx, page, limit)
	if err != nil {
		return nil, 0, 0, err
	}

	// Set defaults for pagination
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	responses := make([]*models.UserResponse, len(users))
	for i, user := range users {
		responses[i] = user.ToResponse()
	}

	return responses, total, totalPages, nil
}

// UpdateUser updates an existing user
func (s *UserService) UpdateUser(ctx context.Context, id string, req *models.UserUpdateRequest, updaterID string) (*models.UserResponse, error) {
	// Check if user exists
	_, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Validate roles if provided
	if len(req.Roles) > 0 {
		roles, err := s.roleRepo.FindByNames(ctx, req.Roles)
		if err != nil {
			return nil, err
		}
		if len(roles) != len(req.Roles) {
			return nil, errors.New("one or more roles do not exist")
		}
	}

	// Build update document
	updates := bson.M{}

	if req.Email != "" {
		updates["email"] = req.Email
	}
	if req.FirstName != "" {
		updates["first_name"] = req.FirstName
	}
	if req.LastName != "" {
		updates["last_name"] = req.LastName
	}
	if req.Roles != nil {
		updates["roles"] = req.Roles
	}
	if req.IsActive != nil {
		updates["is_active"] = *req.IsActive
	}
	if req.Password != "" {
		hashedPassword, err := utils.HashPassword(req.Password)
		if err != nil {
			return nil, errors.New("failed to hash password")
		}
		updates["password_hash"] = hashedPassword
	}

	if len(updates) == 0 {
		return nil, errors.New("no updates provided")
	}

	updatedUser, err := s.userRepo.Update(ctx, id, updates)
	if err != nil {
		return nil, err
	}

	// Log audit event
	s.logAuditEvent(ctx, updaterID, "UPDATE_USER", "user", id, "SUCCESS", "")

	return updatedUser.ToResponse(), nil
}

// DeleteUser deletes a user
func (s *UserService) DeleteUser(ctx context.Context, id, deleterID string) error {
	// Check if user exists
	user, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	// Prevent self-deletion
	if id == deleterID {
		return errors.New("cannot delete your own account")
	}

	// Prevent deletion of admin users by non-admins (additional check could be added here)
	for _, role := range user.Roles {
		if role == "admin" {
			// Could add additional checks here
			break
		}
	}

	if err := s.userRepo.Delete(ctx, id); err != nil {
		return err
	}

	// Log audit event
	s.logAuditEvent(ctx, deleterID, "DELETE_USER", "user", id, "SUCCESS", "")

	return nil
}

// logAuditEvent logs a user management audit event
func (s *UserService) logAuditEvent(ctx context.Context, userID, action, resource, resourceID, status, errorMsg string) {
	log := &models.AuditLog{
		UserID:     userID,
		Service:    "security-service",
		Action:     action,
		Resource:   resource,
		ResourceID: resourceID,
		Status:     status,
		ErrorMsg:   errorMsg,
		Timestamp:  time.Now(),
	}

	s.auditRepo.Create(ctx, log)
}

// InitializeAdminUser creates the default admin user if it doesn't exist
func (s *UserService) InitializeAdminUser(ctx context.Context) error {
	exists, err := s.userRepo.ExistsByUsername(ctx, "admin")
	if err != nil {
		return err
	}

	if !exists {
		hashedPassword, err := utils.HashPassword("admin123") // Default password, should be changed
		if err != nil {
			return err
		}

		admin := &models.User{
			Username:     "admin",
			Email:        "admin@emsib.local",
			PasswordHash: hashedPassword,
			FirstName:    "System",
			LastName:     "Administrator",
			Roles:        []string{"admin"},
			IsActive:     true,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}

		if _, err := s.userRepo.Create(ctx, admin); err != nil {
			return err
		}
	}

	return nil
}
