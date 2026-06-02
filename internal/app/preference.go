package app

import (
	"context"

	apierrors "github.com/fromforgesoftware/go-kit/errors"
)

// PreferenceUsecase is the admin surface for per-recipient channel opt-outs.
type PreferenceUsecase interface {
	Set(ctx context.Context, realmID, recipient, channel string, suppressed bool) error
}

type preferenceUsecase struct {
	prefs NotificationPreferenceRepository
}

func NewPreferenceUsecase(prefs NotificationPreferenceRepository) PreferenceUsecase {
	return &preferenceUsecase{prefs: prefs}
}

func (uc *preferenceUsecase) Set(ctx context.Context, realmID, recipient, channel string, suppressed bool) error {
	if recipient == "" || channel == "" {
		return apierrors.InvalidArgument("recipient and channel are required")
	}
	return uc.prefs.Set(ctx, realmID, recipient, channel, suppressed)
}
