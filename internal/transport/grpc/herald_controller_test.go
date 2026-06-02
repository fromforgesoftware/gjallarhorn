package grpc_test

import (
	"context"
	"testing"

	"github.com/fromforgesoftware/go-kit/resource"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/fromforgesoftware/gjallarhorn/internal/app/apptest"
	"github.com/fromforgesoftware/gjallarhorn/internal/domain"
	gjallarhorngrpc "github.com/fromforgesoftware/gjallarhorn/internal/transport/grpc"
	gjallarhornv1 "github.com/fromforgesoftware/gjallarhorn/pkg/api/gjallarhorn/v1"
)

func TestController_ListDeliveryAttempts(t *testing.T) {
	notifications := apptest.NewNotificationUsecase(t)
	attempts := apptest.NewDeliveryAttemptUsecase(t)
	items := []domain.DeliveryAttempt{
		domain.NewDeliveryAttempt("n-1", 1, domain.StatusFailed, domain.WithDeliveryAttemptError("smtp down")),
		domain.NewDeliveryAttempt("n-1", 2, domain.StatusDelivered),
	}
	attempts.EXPECT().List(mock.Anything, mock.Anything).Return(resource.NewListResponse(items, len(items)), nil)

	c := gjallarhorngrpc.NewGjallarhornController(notifications, attempts).(interface {
		ListDeliveryAttempts(context.Context, *gjallarhornv1.ListDeliveryAttemptsRequest) (*gjallarhornv1.ListDeliveryAttemptsResponse, error)
	})

	resp, err := c.ListDeliveryAttempts(context.Background(), &gjallarhornv1.ListDeliveryAttemptsRequest{NotificationId: "n-1"})
	require.NoError(t, err)
	require.Len(t, resp.GetAttempts(), 2)
	assert.Equal(t, "n-1", resp.GetAttempts()[0].GetNotificationId())
	assert.Equal(t, domain.StatusFailed.String(), resp.GetAttempts()[0].GetStatus())
	assert.Equal(t, int32(2), resp.GetAttempts()[1].GetAttempt())
	assert.Equal(t, domain.StatusDelivered.String(), resp.GetAttempts()[1].GetStatus())
}
