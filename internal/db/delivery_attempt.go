package db

import (
	"context"
	"time"

	"github.com/fromforgesoftware/go-kit/filter"
	"github.com/fromforgesoftware/go-kit/persistence/gormdb"
	"github.com/fromforgesoftware/go-kit/persistence/postgres"
	"github.com/fromforgesoftware/go-kit/resource"
	"github.com/fromforgesoftware/go-kit/search"
	"github.com/fromforgesoftware/go-kit/search/query"
	"github.com/fromforgesoftware/go-kit/slicesx"

	"github.com/fromforgesoftware/gjallarhorn/internal/domain"
	"github.com/fromforgesoftware/gjallarhorn/internal/fields"
)

var deliveryAttemptFieldMapping = map[string]string{
	fields.ID:             "id",
	fields.CreatedAt:      "created_at",
	fields.NotificationID: "notification_id",
	fields.Attempt:        "attempt",
	fields.Status:         "status",
}

type deliveryAttemptEntity struct {
	EID             string    `gorm:"column:id;type:uuid;default:uuid_generate_v4();primaryKey"`
	ECreatedAt      time.Time `gorm:"column:created_at;type:timestamptz;default:now()"`
	ENotificationID string    `gorm:"column:notification_id;type:uuid"`
	EAttempt        int       `gorm:"column:attempt"`
	EStatus         string    `gorm:"column:status"`
	EError          string    `gorm:"column:error"`
	EAttemptedAt    time.Time `gorm:"column:attempted_at;type:timestamptz"`
}

func (e *deliveryAttemptEntity) TableName() string     { return "gjallarhorn.delivery_attempt" }
func (e *deliveryAttemptEntity) Type() resource.Type   { return domain.ResourceTypeDeliveryAttempt }
func (e *deliveryAttemptEntity) ID() string            { return e.EID }
func (e *deliveryAttemptEntity) LID() string           { return "" }
func (e *deliveryAttemptEntity) CreatedAt() time.Time  { return e.ECreatedAt }
func (e *deliveryAttemptEntity) UpdatedAt() time.Time  { return e.ECreatedAt }
func (e *deliveryAttemptEntity) DeletedAt() *time.Time { return nil }

func (e *deliveryAttemptEntity) NotificationID() string { return e.ENotificationID }
func (e *deliveryAttemptEntity) Attempt() int           { return e.EAttempt }
func (e *deliveryAttemptEntity) Status() domain.NotificationStatus {
	return domain.NotificationStatus(e.EStatus)
}
func (e *deliveryAttemptEntity) Error() string          { return e.EError }
func (e *deliveryAttemptEntity) AttemptedAt() time.Time { return e.EAttemptedAt }

func deliveryAttemptToEntity(a domain.DeliveryAttempt) *deliveryAttemptEntity {
	return &deliveryAttemptEntity{
		EID:             a.ID(),
		ENotificationID: a.NotificationID(),
		EAttempt:        a.Attempt(),
		EStatus:         a.Status().String(),
		EError:          a.Error(),
		EAttemptedAt:    a.AttemptedAt(),
	}
}

type deliveryAttemptRepo struct {
	*postgres.Repo
}

func NewDeliveryAttemptRepository(db *gormdb.DBClient) (*deliveryAttemptRepo, error) {
	r, err := postgres.NewRepo(db, deliveryAttemptFieldMapping)
	if err != nil {
		return nil, err
	}
	return &deliveryAttemptRepo{Repo: r}, nil
}

func (r *deliveryAttemptRepo) Create(ctx context.Context, a domain.DeliveryAttempt) (domain.DeliveryAttempt, error) {
	e := deliveryAttemptToEntity(a)
	tx := r.DB.WithContext(ctx)
	if e.EID == "" {
		tx = tx.Omit("id")
	}
	if e.ECreatedAt.IsZero() {
		tx = tx.Omit("created_at")
	}
	if err := tx.Create(e).Error; err != nil {
		return nil, postgres.NewErrUnknown(err)
	}
	return e, nil
}

func (r *deliveryAttemptRepo) ListByNotification(ctx context.Context, notificationID string) ([]domain.DeliveryAttempt, error) {
	q := search.New(search.WithQueryOpts(
		query.FilterBy(filter.OpEq, fields.NotificationID, notificationID),
		query.SortBy(fields.Attempt, query.SortAsc),
	)).Query()
	var found []*deliveryAttemptEntity
	if err := r.QueryApply(ctx, q).Find(&found).Error; err != nil {
		return nil, postgres.NewErrUnknown(err)
	}
	return slicesx.Map(found, func(e *deliveryAttemptEntity) domain.DeliveryAttempt { return e }), nil
}
