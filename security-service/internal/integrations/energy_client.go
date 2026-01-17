package integrations

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"

	"security-service/internal/config"
	"security-service/internal/models"
	"security-service/internal/repository"
	"security-service/pkg/utils"
)

// EnergyProviderClient handles communication with external energy providers
type EnergyProviderClient struct {
	httpClient   *http.Client
	baseURL      string
	apiKey       string
	clientID     string
	clientSecret string
	authRepo     *repository.AuthRepository
	encryptor    *utils.Encryptor

	// Token cache
	mu          sync.RWMutex
	accessToken string
	tokenExpiry time.Time
}

// NewEnergyProviderClient creates a new energy provider client
func NewEnergyProviderClient(cfg *config.Config, authRepo *repository.AuthRepository) (*EnergyProviderClient, error) {
	encryptor, err := utils.NewEncryptor(cfg.Encryption.Key)
	if err != nil {
		return nil, err
	}

	return &EnergyProviderClient{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL:      cfg.Energy.BaseURL,
		apiKey:       cfg.Energy.APIKey,
		clientID:     cfg.Energy.ClientID,
		clientSecret: cfg.Energy.ClientSecret,
		authRepo:     authRepo,
		encryptor:    encryptor,
	}, nil
}

// EnergyTokenResponse represents the OAuth token response from energy provider
type EnergyTokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

// GetConsumption retrieves energy consumption data
func (c *EnergyProviderClient) GetConsumption(ctx context.Context, buildingID string, from, to time.Time) (*models.EnergyConsumption, error) {
	token, err := c.getAccessToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get access token: %w", err)
	}

	// Build URL with query parameters
	endpoint := fmt.Sprintf("%s/consumption", c.baseURL)
	params := url.Values{}
	params.Set("buildingId", buildingID)
	params.Set("from", from.Format(time.RFC3339))
	params.Set("to", to.Format(time.RFC3339))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint+"?"+params.Encode(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("X-API-Key", c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, c.handleErrorResponse(resp)
	}

	var consumption models.EnergyConsumption
	if err := json.NewDecoder(resp.Body).Decode(&consumption); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	consumption.RetrievedAt = time.Now()
	return &consumption, nil
}

// GetTariffs retrieves tariff information for a region
func (c *EnergyProviderClient) GetTariffs(ctx context.Context, region string) (*models.Tariff, error) {
	token, err := c.getAccessToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get access token: %w", err)
	}

	// Build URL with query parameters
	endpoint := fmt.Sprintf("%s/tariffs", c.baseURL)
	params := url.Values{}
	params.Set("region", region)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint+"?"+params.Encode(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("X-API-Key", c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, c.handleErrorResponse(resp)
	}

	var tariff models.Tariff
	if err := json.NewDecoder(resp.Body).Decode(&tariff); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	tariff.RetrievedAt = time.Now()
	return &tariff, nil
}

// RefreshToken refreshes the OAuth token for the energy provider
func (c *EnergyProviderClient) RefreshToken(ctx context.Context) (*models.ExternalTokenRefreshResponse, error) {
	// Request new token from energy provider
	endpoint := fmt.Sprintf("%s/oauth/token", c.baseURL)

	params := url.Values{}
	params.Set("grant_type", "client_credentials")
	params.Set("client_id", c.clientID)
	params.Set("client_secret", c.clientSecret)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.URL.RawQuery = params.Encode()
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return &models.ExternalTokenRefreshResponse{
			Provider: "energy_provider",
			Success:  false,
			Message:  fmt.Sprintf("token refresh failed with status: %d", resp.StatusCode),
		}, nil
	}

	var tokenResp EnergyTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Calculate expiry time
	expiresAt := time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)

	// Encrypt and store the token
	encryptedToken, err := c.encryptor.Encrypt(tokenResp.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt token: %w", err)
	}

	if err := c.authRepo.UpdateAuthCredentialToken(ctx, "energy_provider", encryptedToken, expiresAt); err != nil {
		// If credential doesn't exist, create it
		cred := &models.AuthCredential{
			ServiceName:    "energy_provider",
			EncryptedToken: encryptedToken,
			TokenExpiresAt: &expiresAt,
		}
		if err := c.authRepo.SaveAuthCredential(ctx, cred); err != nil {
			return nil, fmt.Errorf("failed to save token: %w", err)
		}
	}

	// Update cached token
	c.mu.Lock()
	c.accessToken = tokenResp.AccessToken
	c.tokenExpiry = expiresAt
	c.mu.Unlock()

	return &models.ExternalTokenRefreshResponse{
		Provider:  "energy_provider",
		Success:   true,
		ExpiresAt: expiresAt,
		Message:   "Token refreshed successfully",
	}, nil
}

// getAccessToken retrieves or refreshes the access token
func (c *EnergyProviderClient) getAccessToken(ctx context.Context) (string, error) {
	c.mu.RLock()
	if c.accessToken != "" && time.Now().Before(c.tokenExpiry.Add(-5*time.Minute)) {
		token := c.accessToken
		c.mu.RUnlock()
		return token, nil
	}
	c.mu.RUnlock()

	// Try to get from database
	cred, err := c.authRepo.FindAuthCredential(ctx, "energy_provider")
	if err == nil && cred.TokenExpiresAt != nil && time.Now().Before(cred.TokenExpiresAt.Add(-5*time.Minute)) {
		decryptedToken, err := c.encryptor.Decrypt(cred.EncryptedToken)
		if err == nil {
			c.mu.Lock()
			c.accessToken = decryptedToken
			c.tokenExpiry = *cred.TokenExpiresAt
			c.mu.Unlock()
			return decryptedToken, nil
		}
	}

	// Refresh token
	resp, err := c.RefreshToken(ctx)
	if err != nil {
		return "", err
	}
	if !resp.Success {
		return "", fmt.Errorf("token refresh failed: %s", resp.Message)
	}

	c.mu.RLock()
	token := c.accessToken
	c.mu.RUnlock()
	return token, nil
}

// handleErrorResponse handles error responses from the energy provider
func (c *EnergyProviderClient) handleErrorResponse(resp *http.Response) error {
	var apiErr models.ExternalAPIError
	if err := json.NewDecoder(resp.Body).Decode(&apiErr); err == nil {
		apiErr.Provider = "energy_provider"
		apiErr.StatusCode = resp.StatusCode
		return fmt.Errorf("energy provider error: %s (status: %d)", apiErr.Message, apiErr.StatusCode)
	}
	return fmt.Errorf("energy provider returned status: %d", resp.StatusCode)
}
