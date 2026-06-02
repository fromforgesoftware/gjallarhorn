// Package app holds Gjallarhorn's usecases and the ports they depend on.
package app

import (
	"context"
	"time"

	"github.com/fromforgesoftware/go-kit/application/repository"
	"github.com/fromforgesoftware/go-kit/application/usecase"
	apierrors "github.com/fromforgesoftware/go-kit/errors"
	"github.com/fromforgesoftware/go-kit/filter"
	"github.com/fromforgesoftware/go-kit/search"
	"github.com/fromforgesoftware/go-kit/search/query"

	"github.com/fromforgesoftware/gjallarhorn/internal/domain"
	"github.com/fromforgesoftware/gjallarhorn/internal/fields"
)

const defaultQueryLimit = 100

// NotificationFilter narrows a notification list. Zero-value fields are not
// constrained.
type NotificationFilter struct {
	Status    string
	Recipient string
	Limit     int
}

// NotificationRepository persists notifications and mutates their status.
type NotificationRepository interface {
	repository.Creator[domain.Notification]
	repository.Getter[domain.Notification]
	repository.Lister[domain.Notification]
	UpdateStatus(ctx context.Context, id string, status domain.NotificationStatus, lastErr string) error
	// ClaimDue atomically claims due scheduled notifications (SCHEDULED→QUEUED)
	// for the dispatcher.
	ClaimDue(ctx context.Context, now time.Time, limit int) ([]domain.Notification, error)
}

// NotificationPreferenceRepository records per-recipient channel opt-outs and
// answers the suppression check on the send path.
type NotificationPreferenceRepository interface {
	IsSuppressed(ctx context.Context, realmID, recipient, channel string) (bool, error)
	Set(ctx context.Context, realmID, recipient, channel string, suppressed bool) error
}

// NotificationUsecase queues, schedules, dispatches and reads notifications.
type NotificationUsecase interface {
	repository.Getter[domain.Notification]
	repository.Lister[domain.Notification]
	Send(ctx context.Context, n domain.Notification) (domain.Notification, error)
	// Schedule persists a notification for future delivery; the dispatcher
	// sends it once scheduledAt is due.
	Schedule(ctx context.Context, n domain.Notification, at time.Time) (domain.Notification, error)
	// DispatchDue claims and sends up to limit due scheduled notifications,
	// returning how many were dispatched.
	DispatchDue(ctx context.Context, limit int) (int, error)
}

// BackoffPolicy bounds delivery retries with capped exponential backoff.
type BackoffPolicy struct {
	MaxAttempts int
	Base        time.Duration
	Max         time.Duration
}

func DefaultBackoffPolicy() BackoffPolicy {
	return BackoffPolicy{MaxAttempts: 3, Base: 100 * time.Millisecond, Max: 5 * time.Second}
}

func (p BackoffPolicy) delay(attempt int) time.Duration {
	d := p.Base << (attempt - 1)
	if d > p.Max || d <= 0 {
		return p.Max
	}
	return d
}

type notificationUsecase struct {
	usecase.Getter[domain.Notification]
	usecase.Lister[domain.Notification]

	notifications NotificationRepository
	attempts      DeliveryAttemptRepository
	channels      *ChannelRegistry
	renderer      TemplateRenderer
	preferences   NotificationPreferenceRepository
	backoff       BackoffPolicy
	sleep         func(context.Context, time.Duration) error
	now           func() time.Time
}

type UsecaseOption func(*notificationUsecase)

func WithBackoffPolicy(p BackoffPolicy) UsecaseOption {
	return func(uc *notificationUsecase) { uc.backoff = p }
}
func WithSleeper(s func(context.Context, time.Duration) error) UsecaseOption {
	return func(uc *notificationUsecase) { uc.sleep = s }
}
func WithClock(now func() time.Time) UsecaseOption {
	return func(uc *notificationUsecase) { uc.now = now }
}

// WithRenderer enables template rendering: a Send referencing a template name
// (with no explicit subject/body) has them rendered before dispatch.
func WithRenderer(r TemplateRenderer) UsecaseOption {
	return func(uc *notificationUsecase) { uc.renderer = r }
}

// WithPreferences enables the per-recipient channel opt-out check: a suppressed
// channel is persisted as SUPPRESSED and never dispatched.
func WithPreferences(p NotificationPreferenceRepository) UsecaseOption {
	return func(uc *notificationUsecase) { uc.preferences = p }
}

func NewNotificationUsecase(
	notifications NotificationRepository,
	attempts DeliveryAttemptRepository,
	channels *ChannelRegistry,
	opts ...UsecaseOption,
) NotificationUsecase {
	uc := &notificationUsecase{
		Getter:        usecase.NewGetter(notifications, domain.ResourceTypeNotification),
		Lister:        usecase.NewLister[domain.Notification](notifications),
		notifications: notifications,
		attempts:      attempts,
		channels:      channels,
		backoff:       DefaultBackoffPolicy(),
		sleep:         sleepCtx,
		now:           time.Now,
	}
	for _, opt := range opts {
		opt(uc)
	}
	return uc
}

// applyTemplate renders subject/body from the referenced template when a
// renderer is configured and the caller supplied a template name without an
// explicit subject/body. A no-op otherwise.
func (uc *notificationUsecase) applyTemplate(ctx context.Context, n domain.Notification) (domain.Notification, error) {
	if uc.renderer == nil || n.Template() == "" || n.Subject() != "" || n.Body() != "" {
		return n, nil
	}
	subject, body, err := uc.renderer.Render(ctx, n.RealmID(), n.Template(), n.Locale(), n.Data())
	if err != nil {
		return nil, err
	}
	return domain.NewNotification(n.Recipient(), n.Channel(),
		domain.WithNotificationRealmID(n.RealmID()),
		domain.WithNotificationTemplate(n.Template()),
		domain.WithNotificationLocale(n.Locale()),
		domain.WithNotificationData(n.Data()),
		domain.WithNotificationSubject(subject),
		domain.WithNotificationBody(body),
	), nil
}

func (uc *notificationUsecase) Send(ctx context.Context, n domain.Notification) (domain.Notification, error) {
	if err := validateSendable(n); err != nil {
		return nil, err
	}
	if _, err := uc.channels.Sender(n.Channel()); err != nil {
		return nil, err
	}
	if suppressed, err := uc.suppressed(ctx, n); err != nil {
		return nil, err
	} else if suppressed {
		return uc.notifications.Create(ctx, withStatus(n, domain.StatusSuppressed))
	}
	n, err := uc.applyTemplate(ctx, n)
	if err != nil {
		return nil, err
	}
	created, err := uc.notifications.Create(ctx, n)
	if err != nil {
		return nil, err
	}
	return uc.dispatch(ctx, created)
}

// suppressed reports whether the recipient has opted out of the channel, when
// preference checking is enabled.
func (uc *notificationUsecase) suppressed(ctx context.Context, n domain.Notification) (bool, error) {
	if uc.preferences == nil {
		return false, nil
	}
	return uc.preferences.IsSuppressed(ctx, n.RealmID(), n.Recipient(), n.Channel())
}

// withStatus rebuilds a notification carrying a fixed status, preserving its
// addressing and content.
func withStatus(n domain.Notification, status domain.NotificationStatus) domain.Notification {
	return domain.NewNotification(n.Recipient(), n.Channel(),
		domain.WithNotificationRealmID(n.RealmID()),
		domain.WithNotificationTemplate(n.Template()),
		domain.WithNotificationLocale(n.Locale()),
		domain.WithNotificationData(n.Data()),
		domain.WithNotificationSubject(n.Subject()),
		domain.WithNotificationBody(n.Body()),
		domain.WithNotificationStatus(status),
	)
}

// Schedule persists a notification as SCHEDULED for future delivery without
// dispatching it now. The dispatcher claims and sends it once at is due.
func (uc *notificationUsecase) Schedule(ctx context.Context, n domain.Notification, at time.Time) (domain.Notification, error) {
	if err := validateSendable(n); err != nil {
		return nil, err
	}
	if !at.After(uc.now()) {
		return nil, apierrors.InvalidArgument("scheduled time must be in the future")
	}
	if _, err := uc.channels.Sender(n.Channel()); err != nil {
		return nil, err
	}
	if suppressed, err := uc.suppressed(ctx, n); err != nil {
		return nil, err
	} else if suppressed {
		return uc.notifications.Create(ctx, withStatus(n, domain.StatusSuppressed))
	}
	n, err := uc.applyTemplate(ctx, n)
	if err != nil {
		return nil, err
	}
	scheduled := domain.NewNotification(n.Recipient(), n.Channel(),
		domain.WithNotificationRealmID(n.RealmID()),
		domain.WithNotificationTemplate(n.Template()),
		domain.WithNotificationLocale(n.Locale()),
		domain.WithNotificationData(n.Data()),
		domain.WithNotificationSubject(n.Subject()),
		domain.WithNotificationBody(n.Body()),
		domain.WithNotificationStatus(domain.StatusScheduled),
		domain.WithNotificationScheduledAt(&at),
	)
	return uc.notifications.Create(ctx, scheduled)
}

// DispatchDue claims due scheduled notifications and dispatches each, returning
// how many were sent (successfully or not — a failed dispatch is still
// counted as handled and left in its terminal status).
func (uc *notificationUsecase) DispatchDue(ctx context.Context, limit int) (int, error) {
	due, err := uc.notifications.ClaimDue(ctx, uc.now(), limit)
	if err != nil {
		return 0, err
	}
	for _, n := range due {
		if _, err := uc.dispatch(ctx, n); err != nil {
			return 0, err
		}
	}
	return len(due), nil
}

// dispatch runs the retry/backoff send loop over an already-persisted
// notification, updating its terminal status. Shared by immediate Send and the
// scheduled dispatcher.
func (uc *notificationUsecase) dispatch(ctx context.Context, created domain.Notification) (domain.Notification, error) {
	sender, err := uc.channels.Sender(created.Channel())
	if err != nil {
		return nil, err
	}
	var lastErr error
	for attempt := 1; attempt <= uc.backoff.MaxAttempts; attempt++ {
		if attempt > 1 {
			if err := uc.sleep(ctx, uc.backoff.delay(attempt)); err != nil {
				lastErr = err
				break
			}
		}
		sendErr := sender.Send(ctx, created)
		uc.recordAttempt(ctx, created.ID(), attempt, sendErr)
		if sendErr == nil {
			if err := uc.notifications.UpdateStatus(ctx, created.ID(), domain.StatusSent, ""); err != nil {
				return nil, err
			}
			return uc.reload(ctx, created.ID())
		}
		lastErr = sendErr
	}

	msg := ""
	if lastErr != nil {
		msg = lastErr.Error()
	}
	if err := uc.notifications.UpdateStatus(ctx, created.ID(), domain.StatusFailed, msg); err != nil {
		return nil, err
	}
	return uc.reload(ctx, created.ID())
}

func validateSendable(n domain.Notification) error {
	if n == nil || n.Recipient() == "" {
		return apierrors.InvalidArgument("recipient is required")
	}
	if n.Channel() == "" {
		return apierrors.InvalidArgument("channel is required")
	}
	return nil
}

func (uc *notificationUsecase) recordAttempt(ctx context.Context, id string, attempt int, sendErr error) {
	status := domain.StatusSent
	errMsg := ""
	if sendErr != nil {
		status = domain.StatusFailed
		errMsg = sendErr.Error()
	}
	_, _ = uc.attempts.Create(ctx, domain.NewDeliveryAttempt(id, attempt, status,
		domain.WithDeliveryAttemptError(errMsg),
		domain.WithDeliveryAttemptAt(uc.now()),
	))
}

func (uc *notificationUsecase) reload(ctx context.Context, id string) (domain.Notification, error) {
	return uc.notifications.Get(ctx, search.WithQueryOpts(query.FilterBy(filter.OpEq, fields.ID, id)))
}

func sleepCtx(ctx context.Context, d time.Duration) error {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}

// QueryOptions translates a filter into kit search options. Newest-first and a
// defaulted limit are part of the read contract.
func (f NotificationFilter) QueryOptions() []search.Option {
	opts := []query.Option{query.SortBy(fields.CreatedAt, query.SortDesc)}
	if f.Status != "" {
		opts = append(opts, query.FilterBy(filter.OpEq, fields.Status, f.Status))
	}
	if f.Recipient != "" {
		opts = append(opts, query.FilterBy(filter.OpEq, fields.Recipient, f.Recipient))
	}
	limit := f.Limit
	if limit <= 0 {
		limit = defaultQueryLimit
	}
	opts = append(opts, query.Pagination(limit, 0))
	return []search.Option{search.WithQueryOpts(opts...)}
}
