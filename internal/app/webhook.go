package app

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"

	apierrors "github.com/fromforgesoftware/go-kit/errors"

	"github.com/fromforgesoftware/gjallarhorn/internal/domain"
)

// HTTPDoer is the seam over net/http so the webhook sender is testable without
// a live server.
type HTTPDoer interface {
	Do(*http.Request) (*http.Response, error)
}

// WebhookConfig configures the WEBHOOK channel. SigningSecret, when set, signs
// each payload with an HMAC-SHA256 header so receivers can verify authenticity.
type WebhookConfig struct {
	SigningSecret string `envconfig:"WEBHOOK_SIGNING_SECRET"`
}

// WebhookSender delivers a notification by POSTing a JSON payload to the
// recipient URL. The recipient of a WEBHOOK notification is the target URL.
type WebhookSender struct {
	cfg    WebhookConfig
	client HTTPDoer
}

func NewWebhookSender(cfg WebhookConfig) *WebhookSender {
	return &WebhookSender{cfg: cfg, client: http.DefaultClient}
}

func NewWebhookSenderWithClient(cfg WebhookConfig, client HTTPDoer) *WebhookSender {
	return &WebhookSender{cfg: cfg, client: client}
}

func (s *WebhookSender) Channel() string { return domain.ChannelWebhook }

func (s *WebhookSender) Send(ctx context.Context, n domain.Notification) error {
	payload, err := webhookPayload(n)
	if err != nil {
		return apierrors.InternalError("failed to encode webhook payload")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, n.Recipient(), bytes.NewReader(payload))
	if err != nil {
		return apierrors.InvalidArgument("invalid webhook url: " + err.Error())
	}
	req.Header.Set("Content-Type", "application/json")
	if s.cfg.SigningSecret != "" {
		req.Header.Set("X-Gjallarhorn-Signature", "sha256="+signPayload(s.cfg.SigningSecret, payload))
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("webhook returned status %d", resp.StatusCode)
	}
	return nil
}

// webhookPayload is the JSON body delivered to the endpoint. Pure — unit-tested
// so the wire shape is pinned.
func webhookPayload(n domain.Notification) ([]byte, error) {
	return json.Marshal(map[string]any{
		"id":       n.ID(),
		"realmId":  n.RealmID(),
		"template": n.Template(),
		"subject":  n.Subject(),
		"body":     n.Body(),
		"data":     n.Data(),
	})
}

// signPayload returns the hex HMAC-SHA256 of the payload under the secret.
func signPayload(secret string, payload []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	return hex.EncodeToString(mac.Sum(nil))
}
