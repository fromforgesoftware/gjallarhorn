//go:build integration

package db_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/fromforgesoftware/gjallarhorn/internal/db"
	"github.com/fromforgesoftware/gjallarhorn/internal/domain"
	"github.com/fromforgesoftware/gjallarhorn/internal/internaltest"
)

// TestNotificationPreferenceUpsert verifies the opt-out repo against real
// Postgres: a missing row reads as not-suppressed, Set upserts idempotently,
// and toggling back to opted-in is reflected.
func TestNotificationPreferenceUpsert(t *testing.T) {
	client := internaltest.GetDB(t)
	t.Cleanup(func() { internaltest.TruncateTables(t, client) })

	ctx := context.Background()
	repo := db.NewNotificationPreferenceRepository(client)

	// No row yet → not suppressed.
	got, err := repo.IsSuppressed(ctx, "r", "user@x.com", domain.ChannelEmail)
	require.NoError(t, err)
	assert.False(t, got)

	// Opt out.
	require.NoError(t, repo.Set(ctx, "r", "user@x.com", domain.ChannelEmail, true))
	got, err = repo.IsSuppressed(ctx, "r", "user@x.com", domain.ChannelEmail)
	require.NoError(t, err)
	assert.True(t, got)

	// Upsert is idempotent and can flip back to opted-in.
	require.NoError(t, repo.Set(ctx, "r", "user@x.com", domain.ChannelEmail, false))
	got, err = repo.IsSuppressed(ctx, "r", "user@x.com", domain.ChannelEmail)
	require.NoError(t, err)
	assert.False(t, got)

	// A different channel is independent.
	require.NoError(t, repo.Set(ctx, "r", "user@x.com", domain.ChannelWebhook, true))
	got, err = repo.IsSuppressed(ctx, "r", "user@x.com", domain.ChannelWebhook)
	require.NoError(t, err)
	assert.True(t, got)
}
