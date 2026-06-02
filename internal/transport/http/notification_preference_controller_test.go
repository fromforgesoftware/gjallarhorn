package http_test

import (
	"context"
	"net/http"
	"strings"
	"testing"

	"github.com/fromforgesoftware/go-kit/transport/rest/restest"
	"github.com/stretchr/testify/mock"

	"github.com/fromforgesoftware/gjallarhorn/internal/app/apptest"
	"github.com/fromforgesoftware/gjallarhorn/internal/internaltest"
	gjallarhornhttp "github.com/fromforgesoftware/gjallarhorn/internal/transport/http"
)

func jsonapiReq(t *testing.T, method, target, body string) *http.Request {
	req := restest.NewReq(t, context.Background(), method, target, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/vnd.api+json")
	return req
}

func TestNotificationPreferenceController_Set(t *testing.T) {
	uc := apptest.NewPreferenceUsecase(t)
	uc.EXPECT().Set(mock.Anything, "r", "user@x.com", "EMAIL", true).Return(nil)

	restest.NewHandlerSuite(
		restest.NewHandlerTest(
			"POST /api/notification-preferences returns 200",
			internaltest.NewRESTHandler(gjallarhornhttp.NewNotificationPreferenceController(uc)),
			jsonapiReq(t, http.MethodPost, "/api/notification-preferences",
				`{"data":{"type":"notification-preferences","attributes":{"realmId":"r","recipient":"user@x.com","channel":"EMAIL","suppressed":true}}}`),
			restest.AssertResponseStatus(http.StatusOK),
		),
	).Exec(t)
}
