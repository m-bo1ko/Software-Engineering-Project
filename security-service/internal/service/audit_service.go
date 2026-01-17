package service

import (
	"context"

	"security-service/internal/models"
	"security-service/internal/repository"
)

// AuditService handles audit logging business logic
type AuditService struct {
	auditRepo *repository.AuditRepository
}

// NewAuditService creates a new audit service
func NewAuditService(auditRepo *repository.AuditRepository) *AuditService {
	return &AuditService{auditRepo: auditRepo}
}

// CreateLog creates a new audit log entry
func (s *AuditService) CreateLog(ctx context.Context, req *models.AuditLogCreateRequest) (*models.AuditLogResponse, error) {
	log := &models.AuditLog{
		UserID:      req.UserID,
		Username:    req.Username,
		Service:     req.Service,
		Action:      req.Action,
		Resource:    req.Resource,
		ResourceID:  req.ResourceID,
		Details:     req.Details,
		IPAddress:   req.IPAddress,
		UserAgent:   req.UserAgent,
		Status:      req.Status,
		ErrorMsg:    req.ErrorMsg,
		RequestPath: req.RequestPath,
		Method:      req.Method,
	}

	createdLog, err := s.auditRepo.Create(ctx, log)
	if err != nil {
		return nil, err
	}

	return createdLog.ToResponse(), nil
}

// GetLogs retrieves audit logs with filters
func (s *AuditService) GetLogs(ctx context.Context, params models.AuditLogQueryParams) (*models.PaginatedAuditLogsResponse, error) {
	return s.auditRepo.GetPaginatedResponse(ctx, params)
}

// GetLogByID retrieves a specific audit log by ID
func (s *AuditService) GetLogByID(ctx context.Context, id string) (*models.AuditLogResponse, error) {
	log, err := s.auditRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return log.ToResponse(), nil
}

// Log creates an audit log entry (convenience method)
func (s *AuditService) Log(ctx context.Context, userID, username, service, action, resource, resourceID, status, errorMsg, ipAddress, userAgent, requestPath, method string, details map[string]interface{}) error {
	log := &models.AuditLog{
		UserID:      userID,
		Username:    username,
		Service:     service,
		Action:      action,
		Resource:    resource,
		ResourceID:  resourceID,
		Details:     details,
		IPAddress:   ipAddress,
		UserAgent:   userAgent,
		Status:      status,
		ErrorMsg:    errorMsg,
		RequestPath: requestPath,
		Method:      method,
	}

	_, err := s.auditRepo.Create(ctx, log)
	return err
}
