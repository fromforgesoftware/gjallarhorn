package app_test

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/fromforgesoftware/gjallarhorn/internal/app"
	"github.com/fromforgesoftware/gjallarhorn/internal/domain"
)

type fakeDoer struct {
	lastReq *http.Request
	body    string
	status  int
	err     error
}

func (d *fakeDoer) Do(req *http.Request) (*http.Response, error) {
	if d.err != nil {
		return nil, d.err
	}
	if req.Body != nil {
		b, _ := io.ReadAll(req.Body)
		d.body = string(b)
	}
	d.lastReq = req
	status := d.status
	if status == 0 {
		status = http.StatusOK
	}
	return &http.Response{StatusCode: status, Body: io.NopCloser(strings.NewReader(""))}, nil
}

func webhookNotification() domain.Notification {
	return domain.NewNotification("https://hooks.example.com/x", domain.ChannelWebhook,
		domain.WithNotificationSubject("evt deployed"))
}

func TestWebhookSend_PostsSignedJSON(t *testing.T) {
	doer := &fakeDoer{}
	sender := app.NewWebhookSenderWithClient(app.WebhookConfig{SigningSecret: "shh"}, doer)

	require.NoError(t, sender.Send(t.Context(), webhookNotification()))

	assert.Equal(t, http.MethodPost, doer.lastReq.Method)
	assert.Equal(t, "https://hooks.example.com/x", doer.lastReq.URL.String())
	assert.Equal(t, "application/json", doer.lastReq.Header.Get("Content-Type"))
	assert.True(t, strings.HasPrefix(doer.lastReq.Header.Get("X-Gjallarhorn-Signature"), "sha256="),
		"a signing secret produces an HMAC header")

	var payload map[string]any
	require.NoError(t, json.Unmarshal([]byte(doer.body), &payload))
	assert.Equal(t, "evt deployed", payload["subject"])
}

func TestWebhookSend_NoSignatureWithoutSecret(t *testing.T) {
	doer := &fakeDoer{}
	sender := app.NewWebhookSenderWithClient(app.WebhookConfig{}, doer)
	require.NoError(t, sender.Send(t.Context(), webhookNotification()))
	assert.Empty(t, doer.lastReq.Header.Get("X-Gjallarhorn-Signature"))
}

func TestWebhookSend_Non2xxIsError(t *testing.T) {
	doer := &fakeDoer{status: http.StatusInternalServerError}
	sender := app.NewWebhookSenderWithClient(app.WebhookConfig{}, doer)
	require.Error(t, sender.Send(t.Context(), webhookNotification()))
}

func TestWebhookSend_TransportErrorPropagates(t *testing.T) {
	doer := &fakeDoer{err: errors.New("dial failed")}
	sender := app.NewWebhookSenderWithClient(app.WebhookConfig{}, doer)
	require.Error(t, sender.Send(t.Context(), webhookNotification()))
}

func TestWebhookSender_ChannelName(t *testing.T) {
	sender := app.NewWebhookSender(app.WebhookConfig{})
	assert.Equal(t, domain.ChannelWebhook, sender.Channel())
}
