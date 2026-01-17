// Package integrations handles external service integrations
package integrations

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"security-service/internal/config"
)

// NotificationClient handles communication with external notification services
type NotificationClient struct {
	httpClient *http.Client
	emailURL   string
	smsURL     string
	pushURL    string
}

// NewNotificationClient creates a new notification client
func NewNotificationClient(cfg *config.Config) *NotificationClient {
	return &NotificationClient{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		emailURL: cfg.Notification.EmailURL,
		smsURL:   cfg.Notification.SMSURL,
		pushURL:  cfg.Notification.PushURL,
	}
}

// EmailRequest represents the request body for sending email
type EmailRequest struct {
	To      string `json:"to"`
	Subject string `json:"subject"`
	Body    string `json:"body"`
	IsHTML  bool   `json:"isHtml"`
}

// SMSRequest represents the request body for sending SMS
type SMSRequest struct {
	PhoneNumber string `json:"phoneNumber"`
	Message     string `json:"message"`
}

// PushRequest represents the request body for sending push notification
type PushRequest struct {
	DeviceToken string            `json:"deviceToken"`
	Title       string            `json:"title"`
	Body        string            `json:"body"`
	Data        map[string]string `json:"data,omitempty"`
}

// NotificationResponse represents the response from notification services
type NotificationResponse struct {
	Success   bool   `json:"success"`
	MessageID string `json:"messageId,omitempty"`
	Error     string `json:"error,omitempty"`
}

// SendEmail sends an email notification
func (c *NotificationClient) SendEmail(ctx context.Context, to, subject, body string) error {
	req := EmailRequest{
		To:      to,
		Subject: subject,
		Body:    body,
		IsHTML:  false,
	}

	return c.sendRequest(ctx, c.emailURL, req)
}

// SendEmailHTML sends an HTML email notification
func (c *NotificationClient) SendEmailHTML(ctx context.Context, to, subject, body string) error {
	req := EmailRequest{
		To:      to,
		Subject: subject,
		Body:    body,
		IsHTML:  true,
	}

	return c.sendRequest(ctx, c.emailURL, req)
}

// SendSMS sends an SMS notification
func (c *NotificationClient) SendSMS(ctx context.Context, phoneNumber, message string) error {
	req := SMSRequest{
		PhoneNumber: phoneNumber,
		Message:     message,
	}

	return c.sendRequest(ctx, c.smsURL, req)
}

// SendPush sends a push notification
func (c *NotificationClient) SendPush(ctx context.Context, deviceToken, title, body string) error {
	req := PushRequest{
		DeviceToken: deviceToken,
		Title:       title,
		Body:        body,
	}

	return c.sendRequest(ctx, c.pushURL, req)
}

// SendPushWithData sends a push notification with additional data
func (c *NotificationClient) SendPushWithData(ctx context.Context, deviceToken, title, body string, data map[string]string) error {
	req := PushRequest{
		DeviceToken: deviceToken,
		Title:       title,
		Body:        body,
		Data:        data,
	}

	return c.sendRequest(ctx, c.pushURL, req)
}

// sendRequest sends a request to the notification service
func (c *NotificationClient) sendRequest(ctx context.Context, url string, payload interface{}) error {
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusAccepted {
		var notifResp NotificationResponse
		if err := json.NewDecoder(resp.Body).Decode(&notifResp); err == nil && notifResp.Error != "" {
			return fmt.Errorf("notification service error: %s", notifResp.Error)
		}
		return fmt.Errorf("notification service returned status: %d", resp.StatusCode)
	}

	return nil
}
