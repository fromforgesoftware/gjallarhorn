//go:build integration

package db_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/fromforgesoftware/gjallarhorn/internal/db"
	"github.com/fromforgesoftware/gjallarhorn/internal/domain"
	"github.com/fromforgesoftware/gjallarhorn/internal/internaltest"
)

// TestClaimDue verifies the dispatcher's claim against real Postgres: only due
// scheduled rows are claimed (flipped SCHEDULED→QUEUED), future and non-
// scheduled rows are left untouched, and a claimed row isn't claimed twice.
func TestClaimDue(t *testing.T) {
	client := internaltest.GetDB(t)
	t.Cleanup(func() { internaltest.TruncateTables(t, client) })

	ctx := context.Background()
	repo, err := db.NewNotificationRepository(client)
	require.NoError(t, err)

	now := time.Now().UTC()

	dueID := mustCreate(t, repo, internaltest.NewNotification(
		internaltest.WithNStatus(domain.StatusScheduled), internaltest.WithNScheduledAt(now.Add(-time.Minute))))
	mustCreate(t, repo, internaltest.NewNotification(
		internaltest.WithNStatus(domain.StatusScheduled), internaltest.WithNScheduledAt(now.Add(time.Hour))))
	mustCreate(t, repo, internaltest.NewNotification(internaltest.WithNStatus(domain.StatusQueued)))

	claimed, err := repo.ClaimDue(ctx, now, 10)
	require.NoError(t, err)
	require.Len(t, claimed, 1)
	assert.Equal(t, dueID, claimed[0].ID())
	assert.Equal(t, domain.StatusQueued, claimed[0].Status(), "claimed row is flipped to QUEUED")

	// The same row is not claimed twice.
	again, err := repo.ClaimDue(ctx, now, 10)
	require.NoError(t, err)
	assert.Empty(t, again)
}

type notificationCreator interface {
	Create(context.Context, domain.Notification) (domain.Notification, error)
}

func mustCreate(t *testing.T, repo notificationCreator, n domain.Notification) string {
	t.Helper()
	created, err := repo.Create(context.Background(), n)
	require.NoError(t, err)
	return created.ID()
}
