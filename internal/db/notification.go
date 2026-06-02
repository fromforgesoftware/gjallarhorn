// Package db holds Gjallarhorn's Postgres repositories.
package db

import (
	"context"
	"errors"
	"time"

	apierrors "github.com/fromforgesoftware/go-kit/errors"
	"github.com/fromforgesoftware/go-kit/persistence/gormdb"
	"github.com/fromforgesoftware/go-kit/persistence/postgres"
	"github.com/fromforgesoftware/go-kit/resource"
	"github.com/fromforgesoftware/go-kit/search"
	"github.com/fromforgesoftware/go-kit/slicesx"
	"gorm.io/gorm"

	"github.com/fromforgesoftware/gjallarhorn/internal/domain"
	"github.com/fromforgesoftware/gjallarhorn/internal/fields"
)

var notificationFieldMapping = map[string]string{
	fields.ID:        "id",
	fields.CreatedAt: "created_at",
	fields.UpdatedAt: "updated_at",
	fields.RealmID:   "realm_id",
	fields.Recipient: "recipient",
	fields.Channel:   "channel",
	fields.Template:  "template",
	fields.Status:    "status",
}

type notificationEntity struct {
	EID          string         `gorm:"column:id;type:uuid;default:uuid_generate_v4();primaryKey"`
	ECreatedAt   time.Time      `gorm:"column:created_at;type:timestamptz;default:now()"`
	EUpdatedAt   time.Time      `gorm:"column:updated_at;type:timestamptz;default:now()"`
	ERealmID     *string        `gorm:"column:realm_id;type:uuid"`
	ERecipient   string         `gorm:"column:recipient"`
	EChannel     string         `gorm:"column:channel"`
	ETemplate    string         `gorm:"column:template"`
	ELocale      string         `gorm:"column:locale"`
	EData        map[string]any `gorm:"column:data;type:jsonb;serializer:json"`
	ESubject     string         `gorm:"column:subject"`
	EBody        string         `gorm:"column:body"`
	EStatus      string         `gorm:"column:status"`
	ELastError   string         `gorm:"column:last_error"`
	EScheduledAt *time.Time     `gorm:"column:scheduled_at"`
}

func (e *notificationEntity) TableName() string     { return "gjallarhorn.notification" }
func (e *notificationEntity) Type() resource.Type   { return domain.ResourceTypeNotification }
func (e *notificationEntity) ID() string            { return e.EID }
func (e *notificationEntity) LID() string           { return "" }
func (e *notificationEntity) CreatedAt() time.Time  { return e.ECreatedAt }
func (e *notificationEntity) UpdatedAt() time.Time  { return e.EUpdatedAt }
func (e *notificationEntity) DeletedAt() *time.Time { return nil }

func (e *notificationEntity) Recipient() string    { return e.ERecipient }
func (e *notificationEntity) Channel() string      { return e.EChannel }
func (e *notificationEntity) Template() string     { return e.ETemplate }
func (e *notificationEntity) Locale() string       { return e.ELocale }
func (e *notificationEntity) Data() map[string]any { return e.EData }
func (e *notificationEntity) Subject() string      { return e.ESubject }
func (e *notificationEntity) Body() string         { return e.EBody }
func (e *notificationEntity) Status() domain.NotificationStatus {
	return domain.NotificationStatus(e.EStatus)
}
func (e *notificationEntity) LastError() string       { return e.ELastError }
func (e *notificationEntity) ScheduledAt() *time.Time { return e.EScheduledAt }

func (e *notificationEntity) RealmID() string {
	if e.ERealmID == nil {
		return ""
	}
	return *e.ERealmID
}

func notificationToEntity(n domain.Notification) *notificationEntity {
	e := &notificationEntity{
		EID:          n.ID(),
		ERecipient:   n.Recipient(),
		EChannel:     n.Channel(),
		ETemplate:    n.Template(),
		ELocale:      n.Locale(),
		EData:        n.Data(),
		ESubject:     n.Subject(),
		EBody:        n.Body(),
		EStatus:      n.Status().String(),
		ELastError:   n.LastError(),
		EScheduledAt: n.ScheduledAt(),
	}
	if rid := n.RealmID(); rid != "" {
		e.ERealmID = &rid
	}
	return e
}

type notificationRepo struct {
	*postgres.Repo
}

func NewNotificationRepository(db *gormdb.DBClient) (*notificationRepo, error) {
	r, err := postgres.NewRepo(db, notificationFieldMapping)
	if err != nil {
		return nil, err
	}
	return &notificationRepo{Repo: r}, nil
}

func (r *notificationRepo) Create(ctx context.Context, n domain.Notification) (domain.Notification, error) {
	e := notificationToEntity(n)
	tx := r.DB.WithContext(ctx)
	if e.EID == "" {
		tx = tx.Omit("id")
	}
	if e.ECreatedAt.IsZero() {
		tx = tx.Omit("created_at", "updated_at")
	}
	if err := tx.Create(e).Error; err != nil {
		return nil, postgres.NewErrUnknown(err)
	}
	return e, nil
}

func (r *notificationRepo) Get(ctx context.Context, opts ...search.Option) (domain.Notification, error) {
	s := search.New(opts...)
	var e notificationEntity
	if err := r.QueryApply(ctx, s.Query()).First(&e).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apierrors.NotFound("notification", "")
		}
		return nil, postgres.NewErrUnknown(err)
	}
	return &e, nil
}

func (r *notificationRepo) List(ctx context.Context, opts ...search.Option) (resource.ListResponse[domain.Notification], error) {
	q := search.New(opts...).Query()
	var found []*notificationEntity
	if err := r.QueryApply(ctx, q).Find(&found).Error; err != nil {
		return nil, postgres.NewErrUnknown(err)
	}
	var total int64
	if err := r.CountApply(ctx, new(notificationEntity), q).Count(&total).Error; err != nil {
		return nil, postgres.NewErrUnknown(err)
	}
	out := slicesx.Map(found, func(e *notificationEntity) domain.Notification { return e })
	return resource.NewListResponse(out, int(total)), nil
}

func (r *notificationRepo) UpdateStatus(ctx context.Context, id string, status domain.NotificationStatus, lastErr string) error {
	res := r.DB.WithContext(ctx).Exec(
		`UPDATE gjallarhorn.notification SET status = ?, last_error = ?, updated_at = now() WHERE id = ?`,
		status.String(), lastErr, id,
	)
	if res.Error != nil {
		return postgres.NewErrUnknown(res.Error)
	}
	if res.RowsAffected == 0 {
		return apierrors.NotFound("notification", id)
	}
	return nil
}

// ClaimDue atomically claims up to limit due scheduled notifications, flipping
// them SCHEDULED→QUEUED and returning them. FOR UPDATE SKIP LOCKED lets
// multiple dispatcher instances run without claiming the same row twice.
func (r *notificationRepo) ClaimDue(ctx context.Context, now time.Time, limit int) ([]domain.Notification, error) {
	var found []*notificationEntity
	err := r.DB.WithContext(ctx).Raw(
		`UPDATE gjallarhorn.notification SET status = ?, updated_at = now()
		 WHERE id IN (
		     SELECT id FROM gjallarhorn.notification
		     WHERE status = ? AND scheduled_at IS NOT NULL AND scheduled_at <= ?
		     ORDER BY scheduled_at
		     FOR UPDATE SKIP LOCKED
		     LIMIT ?
		 )
		 RETURNING *`,
		domain.StatusQueued.String(), domain.StatusScheduled.String(), now, limit,
	).Scan(&found).Error
	if err != nil {
		return nil, postgres.NewErrUnknown(err)
	}
	return slicesx.Map(found, func(e *notificationEntity) domain.Notification { return e }), nil
}
