package db

import (
	"context"
	"time"

	"github.com/fromforgesoftware/go-kit/persistence/gormdb"
	"github.com/fromforgesoftware/go-kit/persistence/postgres"
	"gorm.io/gorm/clause"
)

type notificationPreferenceEntity struct {
	EID         string    `gorm:"column:id;type:uuid;default:uuid_generate_v4();primaryKey"`
	ECreatedAt  time.Time `gorm:"column:created_at;type:timestamptz;default:now()"`
	EUpdatedAt  time.Time `gorm:"column:updated_at;type:timestamptz;default:now()"`
	ERealmID    string    `gorm:"column:realm_id"`
	ERecipient  string    `gorm:"column:recipient"`
	EChannel    string    `gorm:"column:channel"`
	ESuppressed bool      `gorm:"column:suppressed"`
}

func (e *notificationPreferenceEntity) TableName() string { return "gjallarhorn.notification_preference" }

type notificationPreferenceRepo struct {
	db *gormdb.DBClient
}

func NewNotificationPreferenceRepository(db *gormdb.DBClient) *notificationPreferenceRepo {
	return &notificationPreferenceRepo{db: db}
}

// IsSuppressed reports whether the recipient has muted the channel for the
// realm. A missing row means "not suppressed" (opted in by default).
func (r *notificationPreferenceRepo) IsSuppressed(ctx context.Context, realmID, recipient, channel string) (bool, error) {
	var suppressed bool
	err := r.db.WithContext(ctx).
		Raw(`SELECT suppressed FROM gjallarhorn.notification_preference
		     WHERE realm_id = ? AND recipient = ? AND channel = ?`,
			realmID, recipient, channel).
		Scan(&suppressed).Error
	if err != nil {
		return false, postgres.NewErrUnknown(err)
	}
	return suppressed, nil
}

// Set upserts the recipient's preference for a channel.
func (r *notificationPreferenceRepo) Set(ctx context.Context, realmID, recipient, channel string, suppressed bool) error {
	e := &notificationPreferenceEntity{
		ERealmID: realmID, ERecipient: recipient, EChannel: channel, ESuppressed: suppressed,
	}
	err := r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "realm_id"}, {Name: "recipient"}, {Name: "channel"}},
		DoUpdates: clause.Assignments(map[string]any{"suppressed": suppressed, "updated_at": time.Now().UTC()}),
	}).Create(e).Error
	if err != nil {
		return postgres.NewErrUnknown(err)
	}
	return nil
}
