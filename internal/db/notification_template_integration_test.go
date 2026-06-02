//go:build integration

package db_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/fromforgesoftware/gjallarhorn/internal/db"
	"github.com/fromforgesoftware/gjallarhorn/internal/domain"
	"github.com/fromforgesoftware/gjallarhorn/internal/internaltest"
)

func TestNotificationTemplate_CRUDAndLocaleFallback(t *testing.T) {
	client := internaltest.GetDB(t)
	t.Cleanup(func() { internaltest.TruncateTables(t, client) })

	ctx := context.Background()
	repo, err := db.NewNotificationTemplateRepository(client)
	require.NoError(t, err)
	realmID := uuid.NewString()

	// A default-locale template + a Spanish override for the same name.
	_, err = repo.Create(ctx, domain.NewNotificationTemplate(realmID, "welcome",
		domain.WithTemplateSubject("Welcome {{.name}}"), domain.WithTemplateBody("Hi")))
	require.NoError(t, err)
	_, err = repo.Create(ctx, domain.NewNotificationTemplate(realmID, "welcome",
		domain.WithTemplateLocale("es"), domain.WithTemplateSubject("Hola {{.name}}"), domain.WithTemplateBody("Bienvenida")))
	require.NoError(t, err)

	// Exact locale match.
	es, err := repo.GetForRender(ctx, realmID, "welcome", "es")
	require.NoError(t, err)
	assert.Equal(t, "Hola {{.name}}", es.Subject())

	// Missing locale falls back to the default-locale template.
	fr, err := repo.GetForRender(ctx, realmID, "welcome", "fr")
	require.NoError(t, err)
	assert.Equal(t, "Welcome {{.name}}", fr.Subject())
}
