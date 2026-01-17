package service

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson"

	"security-service/internal/models"
	"security-service/internal/repository"
)

// RoleService handles role management business logic
type RoleService struct {
	roleRepo  *repository.RoleRepository
	auditRepo *repository.AuditRepository
}

// NewRoleService creates a new role service
func NewRoleService(roleRepo *repository.RoleRepository, auditRepo *repository.AuditRepository) *RoleService {
	return &RoleService{
		roleRepo:  roleRepo,
		auditRepo: auditRepo,
	}
}

// CreateRole creates a new role
func (s *RoleService) CreateRole(ctx context.Context, req *models.RoleCreateRequest, creatorID string) (*models.RoleResponse, error) {
	// Check if role already exists
	exists, err := s.roleRepo.ExistsByName(ctx, req.Name)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("role with this name already exists")
	}

	role := &models.Role{
		Name:        req.Name,
		Description: req.Description,
		Permissions: req.Permissions,
		IsSystem:    false,
	}

	if role.Permissions == nil {
		role.Permissions = []models.Permission{}
	}

	createdRole, err := s.roleRepo.Create(ctx, role)
	if err != nil {
		return nil, err
	}

	// Log audit event
	s.logAuditEvent(ctx, creatorID, "CREATE_ROLE", "role", createdRole.Name, "SUCCESS", "")

	return createdRole.ToResponse(), nil
}

// GetRole retrieves a role by name
func (s *RoleService) GetRole(ctx context.Context, name string) (*models.RoleResponse, error) {
	role, err := s.roleRepo.FindByName(ctx, name)
	if err != nil {
		return nil, err
	}
	return role.ToResponse(), nil
}

// ListRoles retrieves all roles
func (s *RoleService) ListRoles(ctx context.Context) ([]*models.RoleResponse, error) {
	roles, err := s.roleRepo.FindAll(ctx)
	if err != nil {
		return nil, err
	}

	responses := make([]*models.RoleResponse, len(roles))
	for i, role := range roles {
		responses[i] = role.ToResponse()
	}

	return responses, nil
}

// UpdateRole updates an existing role's permissions
func (s *RoleService) UpdateRole(ctx context.Context, name string, req *models.RoleUpdateRequest, updaterID string) (*models.RoleResponse, error) {
	// Check if role exists
	existingRole, err := s.roleRepo.FindByName(ctx, name)
	if err != nil {
		return nil, err
	}

	// Prevent modification of system roles' core properties
	if existingRole.IsSystem {
		// Allow updating permissions but not other properties
		if req.Description != "" && req.Description != existingRole.Description {
			return nil, errors.New("cannot modify description of system role")
		}
	}

	// Build update document
	updates := bson.M{}

	if req.Description != "" {
		updates["description"] = req.Description
	}
	if req.Permissions != nil {
		updates["permissions"] = req.Permissions
	}

	if len(updates) == 0 {
		return nil, errors.New("no updates provided")
	}

	updatedRole, err := s.roleRepo.Update(ctx, name, updates)
	if err != nil {
		return nil, err
	}

	// Log audit event
	s.logAuditEvent(ctx, updaterID, "UPDATE_ROLE", "role", name, "SUCCESS", "")

	return updatedRole.ToResponse(), nil
}

// DeleteRole deletes a role
func (s *RoleService) DeleteRole(ctx context.Context, name, deleterID string) error {
	if err := s.roleRepo.Delete(ctx, name); err != nil {
		return err
	}

	// Log audit event
	s.logAuditEvent(ctx, deleterID, "DELETE_ROLE", "role", name, "SUCCESS", "")

	return nil
}

// InitializeDefaultRoles creates default system roles
func (s *RoleService) InitializeDefaultRoles(ctx context.Context) error {
	return s.roleRepo.InitializeDefaultRoles(ctx)
}

// logAuditEvent logs a role management audit event
func (s *RoleService) logAuditEvent(ctx context.Context, userID, action, resource, resourceID, status, errorMsg string) {
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
