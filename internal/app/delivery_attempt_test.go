package app_test

import (
	"context"
	"errors"
	"testing"

	apierrors "github.com/fromforgesoftware/go-kit/errors"
	"github.com/fromforgesoftware/go-kit/filter"
	"github.com/fromforgesoftware/go-kit/search"
	"github.com/fromforgesoftware/go-kit/search/query"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/fromforgesoftware/gjallarhorn/internal/app"
	"github.com/fromforgesoftware/gjallarhorn/internal/app/apptest"
	"github.com/fromforgesoftware/gjallarhorn/internal/domain"
	"github.com/fromforgesoftware/gjallarhorn/internal/fields"
	"github.com/fromforgesoftware/gjallarhorn/internal/internaltest"
)

func notificationIDFilter(id string) search.Option {
	return search.WithQueryOpts(query.FilterBy(filter.OpEq, fields.NotificationID, id))
}

func TestDeliveryAttemptUsecase_List_ReturnsAttempts(t *testing.T) {
	repo := apptest.NewDeliveryAttemptRepository(t)
	want := []domain.DeliveryAttempt{
		internaltest.NewDeliveryAttempt("n-1", 1, domain.StatusFailed, domain.WithDeliveryAttemptError("smtp down")),
		internaltest.NewDeliveryAttempt("n-1", 2, domain.StatusDelivered),
	}
	repo.EXPECT().ListByNotification(mock.Anything, "n-1").Return(want, nil)

	uc := app.NewDeliveryAttemptUsecase(repo)
	res, err := uc.List(context.Background(), notificationIDFilter("n-1"))
	require.NoError(t, err)
	require.Len(t, res.Results(), 2)
	assert.Equal(t, 2, res.TotalCount())
	assert.Equal(t, domain.StatusFailed, res.Results()[0].Status())
	assert.Equal(t, domain.StatusDelivered, res.Results()[1].Status())
}

func TestDeliveryAttemptUsecase_List_RequiresNotificationID(t *testing.T) {
	repo := apptest.NewDeliveryAttemptRepository(t)
	uc := app.NewDeliveryAttemptUsecase(repo)

	_, err := uc.List(context.Background())
	require.Error(t, err)
	assert.True(t, apierrors.Is(err, apierrors.CodeInvalidArgument))
}

func TestDeliveryAttemptUsecase_List_PropagatesRepoError(t *testing.T) {
	repo := apptest.NewDeliveryAttemptRepository(t)
	repo.EXPECT().ListByNotification(mock.Anything, "n-1").Return(nil, errors.New("db down"))

	uc := app.NewDeliveryAttemptUsecase(repo)
	_, err := uc.List(context.Background(), notificationIDFilter("n-1"))
	require.Error(t, err)
}
