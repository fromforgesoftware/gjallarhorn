package http_test

import (
	"net/http"
	"testing"

	"github.com/fromforgesoftware/go-kit/resource"
	"github.com/fromforgesoftware/go-kit/transport/rest/restest"
	"github.com/stretchr/testify/mock"

	"github.com/fromforgesoftware/gjallarhorn/internal/app/apptest"
	"github.com/fromforgesoftware/gjallarhorn/internal/domain"
	"github.com/fromforgesoftware/gjallarhorn/internal/internaltest"
	gjallarhornhttp "github.com/fromforgesoftware/gjallarhorn/internal/transport/http"
)

func TestDeliveryAttemptController_List(t *testing.T) {
	uc := apptest.NewDeliveryAttemptUsecase(t)
	// The handler parses filter[notificationId] off the query string into the
	// search options; the usecase reads it and returns the attempt log.
	items := []domain.DeliveryAttempt{
		internaltest.NewDeliveryAttempt("n-1", 1, domain.StatusFailed, domain.WithDeliveryAttemptError("smtp down")),
		internaltest.NewDeliveryAttempt("n-1", 2, domain.StatusDelivered),
	}
	uc.EXPECT().List(mock.Anything, mock.Anything).Return(resource.NewListResponse(items, len(items)), nil)

	restest.NewHandlerSuite(
		restest.NewHandlerTest(
			"GET /api/delivery-attempts returns 200",
			internaltest.NewRESTHandler(gjallarhornhttp.NewDeliveryAttemptController(uc)),
			jsonapiReq(t, http.MethodGet, "/api/delivery-attempts?filter[notificationId][eq]=n-1", ""),
			restest.AssertResponseStatus(http.StatusOK),
		),
	).Exec(t)
}
