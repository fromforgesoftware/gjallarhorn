package app

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	apierrors "github.com/fromforgesoftware/go-kit/errors"

	"github.com/fromforgesoftware/gjallarhorn/internal/domain"
)

// SMSConfig / PushConfig point the SMS and PUSH channels at an HTTP gateway
// (Twilio, a carrier relay, FCM HTTP v1, …). The integration is provider-
// agnostic: a JSON POST with an optional Authorization header. AuthHeader is
// the full header value (e.g. "Bearer …" or "Basic …").
type SMSConfig struct {
	EndpointURL string `envconfig:"SMS_ENDPOINT_URL"`
	AuthHeader  string `envconfig:"SMS_AUTH_HEADER"`
}

type PushConfig struct {
	EndpointURL string `envconfig:"PUSH_ENDPOINT_URL"`
	AuthHeader  string `envconfig:"PUSH_AUTH_HEADER"`
}

// SMSSender delivers a notification as an SMS via the configured HTTP gateway.
// The recipient is the destination phone number.
type SMSSender struct {
	cfg    SMSConfig
	client HTTPDoer
}

func NewSMSSender(cfg SMSConfig) *SMSSender { return &SMSSender{cfg: cfg, client: http.DefaultClient} }
func NewSMSSenderWithClient(cfg SMSConfig, client HTTPDoer) *SMSSender {
	return &SMSSender{cfg: cfg, client: client}
}
func (s *SMSSender) Channel() string { return domain.ChannelSMS }

func (s *SMSSender) Send(ctx context.Context, n domain.Notification) error {
	return postJSON(ctx, s.client, s.cfg.EndpointURL, s.cfg.AuthHeader, "SMS", map[string]any{
		"to":   n.Recipient(),
		"body": n.Body(),
	})
}

// PushSender delivers a notification as a push message via the configured HTTP
// gateway. The recipient is the device token.
type PushSender struct {
	cfg    PushConfig
	client HTTPDoer
}

func NewPushSender(cfg PushConfig) *PushSender {
	return &PushSender{cfg: cfg, client: http.DefaultClient}
}
func NewPushSenderWithClient(cfg PushConfig, client HTTPDoer) *PushSender {
	return &PushSender{cfg: cfg, client: client}
}
func (s *PushSender) Channel() string { return domain.ChannelPush }

func (s *PushSender) Send(ctx context.Context, n domain.Notification) error {
	return postJSON(ctx, s.client, s.cfg.EndpointURL, s.cfg.AuthHeader, "push", map[string]any{
		"token": n.Recipient(),
		"title": n.Subject(),
		"body":  n.Body(),
		"data":  n.Data(),
	})
}

// postJSON POSTs a JSON body to an HTTP gateway, treating non-2xx as an error
// so it flows through the usecase's retry/backoff. An unconfigured endpoint is
// a clear configuration error rather than a silent drop.
func postJSON(ctx context.Context, client HTTPDoer, endpoint, authHeader, channel string, payload map[string]any) error {
	if endpoint == "" {
		return apierrors.InvalidArgument(channel + " channel is not configured")
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return apierrors.InternalError("failed to encode " + channel + " payload")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return apierrors.InvalidArgument("invalid " + channel + " endpoint: " + err.Error())
	}
	req.Header.Set("Content-Type", "application/json")
	if authHeader != "" {
		req.Header.Set("Authorization", authHeader)
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("%s gateway returned status %d", channel, resp.StatusCode)
	}
	return nil
}
