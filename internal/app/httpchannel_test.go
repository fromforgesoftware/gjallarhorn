package app_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/fromforgesoftware/gjallarhorn/internal/app"
	"github.com/fromforgesoftware/gjallarhorn/internal/domain"
)

func TestSMSSend_PostsToGatewayWithAuth(t *testing.T) {
	doer := &fakeDoer{}
	sender := app.NewSMSSenderWithClient(app.SMSConfig{EndpointURL: "https://sms.example/send", AuthHeader: "Bearer t"}, doer)

	n := domain.NewNotification("+15551234567", domain.ChannelSMS, domain.WithNotificationBody("code 1234"))
	require.NoError(t, sender.Send(t.Context(), n))

	assert.Equal(t, "https://sms.example/send", doer.lastReq.URL.String())
	assert.Equal(t, "Bearer t", doer.lastReq.Header.Get("Authorization"))
	var payload map[string]any
	require.NoError(t, json.Unmarshal([]byte(doer.body), &payload))
	assert.Equal(t, "+15551234567", payload["to"])
	assert.Equal(t, "code 1234", payload["body"])
}

func TestPushSend_PostsTokenTitleBody(t *testing.T) {
	doer := &fakeDoer{}
	sender := app.NewPushSenderWithClient(app.PushConfig{EndpointURL: "https://push.example/send"}, doer)

	n := domain.NewNotification("device-tok", domain.ChannelPush,
		domain.WithNotificationSubject("Deploy done"), domain.WithNotificationBody("v2 is live"))
	require.NoError(t, sender.Send(t.Context(), n))

	var payload map[string]any
	require.NoError(t, json.Unmarshal([]byte(doer.body), &payload))
	assert.Equal(t, "device-tok", payload["token"])
	assert.Equal(t, "Deploy done", payload["title"])
	assert.Empty(t, doer.lastReq.Header.Get("Authorization"))
}

func TestHTTPChannel_UnconfiguredEndpointErrors(t *testing.T) {
	sender := app.NewSMSSender(app.SMSConfig{})
	err := sender.Send(t.Context(), domain.NewNotification("+1", domain.ChannelSMS))
	require.Error(t, err)
}

func TestHTTPChannel_Non2xxIsError(t *testing.T) {
	doer := &fakeDoer{status: http.StatusBadGateway}
	sender := app.NewPushSenderWithClient(app.PushConfig{EndpointURL: "https://push.example"}, doer)
	require.Error(t, sender.Send(t.Context(), domain.NewNotification("tok", domain.ChannelPush)))
}

func TestHTTPChannel_Names(t *testing.T) {
	assert.Equal(t, domain.ChannelSMS, app.NewSMSSender(app.SMSConfig{}).Channel())
	assert.Equal(t, domain.ChannelPush, app.NewPushSender(app.PushConfig{}).Channel())
}
