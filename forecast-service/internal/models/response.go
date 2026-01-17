package models

// APIResponse represents a standard API response wrapper
type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   *APIError   `json:"error,omitempty"`
}

// APIError represents an error in the API response
type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// NewSuccessResponse creates a successful API response
func NewSuccessResponse(data interface{}, message string) *APIResponse {
	return &APIResponse{
		Success: true,
		Message: message,
		Data:    data,
	}
}

// NewErrorResponse creates an error API response
func NewErrorResponse(code, message, details string) *APIResponse {
	return &APIResponse{
		Success: false,
		Error: &APIError{
			Code:    code,
			Message: message,
			Details: details,
		},
	}
}

// Common error codes
const (
	ErrCodeInvalidRequest   = "INVALID_REQUEST"
	ErrCodeUnauthorized     = "UNAUTHORIZED"
	ErrCodeForbidden        = "FORBIDDEN"
	ErrCodeNotFound         = "NOT_FOUND"
	ErrCodeConflict         = "CONFLICT"
	ErrCodeInternalError    = "INTERNAL_ERROR"
	ErrCodeValidationFailed = "VALIDATION_FAILED"
	ErrCodeTokenExpired     = "TOKEN_EXPIRED"
	ErrCodeTokenInvalid     = "TOKEN_INVALID"
	ErrCodeExternalAPIError = "EXTERNAL_API_ERROR"
	ErrCodeForecastFailed   = "FORECAST_FAILED"
	ErrCodeOptimizationFailed = "OPTIMIZATION_FAILED"
)

// TokenValidationResponse represents the response from security service
type TokenValidationResponse struct {
	Valid   bool     `json:"valid"`
	UserID  string   `json:"userId,omitempty"`
	Roles   []string `json:"roles,omitempty"`
	Message string   `json:"message,omitempty"`
}

// AuditLogRequest represents a request to log an audit event
type AuditLogRequest struct {
	UserID      string                 `json:"userId"`
	Username    string                 `json:"username"`
	Service     string                 `json:"service"`
	Action      string                 `json:"action"`
	Resource    string                 `json:"resource"`
	ResourceID  string                 `json:"resourceId"`
	Status      string                 `json:"status"`
	ErrorMsg    string                 `json:"errorMsg,omitempty"`
	Details     map[string]interface{} `json:"details,omitempty"`
	IPAddress   string                 `json:"ipAddress"`
	UserAgent   string                 `json:"userAgent"`
	RequestPath string                 `json:"requestPath"`
	Method      string                 `json:"method"`
}
