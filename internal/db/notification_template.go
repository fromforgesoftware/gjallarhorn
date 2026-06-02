package db

import (
	"context"
	"errors"
	"time"

	"github.com/fromforgesoftware/go-kit/application/repository"
	apierrors "github.com/fromforgesoftware/go-kit/errors"
	"github.com/fromforgesoftware/go-kit/filter"
	"github.com/fromforgesoftware/go-kit/persistence/gormdb"
	"github.com/fromforgesoftware/go-kit/persistence/postgres"
	"github.com/fromforgesoftware/go-kit/resource"
	"github.com/fromforgesoftware/go-kit/search"
	"github.com/fromforgesoftware/go-kit/search/query"
	"github.com/fromforgesoftware/go-kit/slicesx"
	"gorm.io/gorm"

	"github.com/fromforgesoftware/gjallarhorn/internal/domain"
	"github.com/fromforgesoftware/gjallarhorn/internal/fields"
)

var templateFieldMapping = map[string]string{
	fields.ID:      "id",
	fields.RealmID: "realm_id",
	fields.Name:    "name",
	fields.Locale:  "locale",
}

type templateEntity struct {
	EID        string    `gorm:"column:id;type:uuid;default:uuid_generate_v4();primaryKey"`
	ECreatedAt time.Time `gorm:"column:created_at;type:timestamptz;default:now()"`
	EUpdatedAt time.Time `gorm:"column:updated_at;type:timestamptz;default:now()"`
	ERealmID   string    `gorm:"column:realm_id;type:uuid"`
	EName      string    `gorm:"column:name"`
	EChannel   string    `gorm:"column:channel"`
	ELocale    string    `gorm:"column:locale"`
	ESubject   string    `gorm:"column:subject"`
	EBody      string    `gorm:"column:body"`
}

func (e *templateEntity) TableName() string     { return "gjallarhorn.notification_template" }
func (e *templateEntity) Type() resource.Type   { return domain.ResourceTypeNotificationTemplate }
func (e *templateEntity) ID() string            { return e.EID }
func (e *templateEntity) LID() string           { return "" }
func (e *templateEntity) CreatedAt() time.Time  { return e.ECreatedAt }
func (e *templateEntity) UpdatedAt() time.Time  { return e.EUpdatedAt }
func (e *templateEntity) DeletedAt() *time.Time { return nil }
func (e *templateEntity) RealmID() string       { return e.ERealmID }
func (e *templateEntity) Name() string          { return e.EName }
func (e *templateEntity) Channel() string       { return e.EChannel }
func (e *templateEntity) Locale() string        { return e.ELocale }
func (e *templateEntity) Subject() string       { return e.ESubject }
func (e *templateEntity) Body() string          { return e.EBody }

func templateToEntity(t domain.NotificationTemplate) *templateEntity {
	return &templateEntity{
		EID:      t.ID(),
		ERealmID: t.RealmID(),
		EName:    t.Name(),
		EChannel: t.Channel(),
		ELocale:  t.Locale(),
		ESubject: t.Subject(),
		EBody:    t.Body(),
	}
}

type templateRepo struct {
	*postgres.Repo
}

func NewNotificationTemplateRepository(db *gormdb.DBClient) (*templateRepo, error) {
	r, err := postgres.NewRepo(db, templateFieldMapping)
	if err != nil {
		return nil, err
	}
	return &templateRepo{Repo: r}, nil
}

func (r *templateRepo) Create(ctx context.Context, t domain.NotificationTemplate) (domain.NotificationTemplate, error) {
	e := templateToEntity(t)
	tx := r.DB.WithContext(ctx)
	if e.EID == "" {
		tx = tx.Omit("id", "created_at", "updated_at")
	}
	if err := tx.Create(e).Error; err != nil {
		if postgres.ErrorIs(err, "23505") {
			return nil, apierrors.AlreadyExists("notification template", t.Name())
		}
		return nil, postgres.NewErrUnknown(err)
	}
	return e, nil
}

func (r *templateRepo) Get(ctx context.Context, opts ...search.Option) (domain.NotificationTemplate, error) {
	s := search.New(opts...)
	var e templateEntity
	if err := r.QueryApply(ctx, s.Query()).First(&e).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apierrors.NotFound("notification template", "")
		}
		return nil, postgres.NewErrUnknown(err)
	}
	return &e, nil
}

func (r *templateRepo) List(ctx context.Context, opts ...search.Option) (resource.ListResponse[domain.NotificationTemplate], error) {
	s := search.New(opts...)
	var found []*templateEntity
	if err := r.QueryApply(ctx, s.Query()).Find(&found).Error; err != nil {
		return nil, postgres.NewErrUnknown(err)
	}
	var total int64
	if err := r.CountApply(ctx, new(templateEntity), s.Query()).Count(&total).Error; err != nil {
		return nil, postgres.NewErrUnknown(err)
	}
	out := slicesx.Map(found, func(e *templateEntity) domain.NotificationTemplate { return e })
	return resource.NewListResponse(out, int(total)), nil
}

func (r *templateRepo) Delete(ctx context.Context, delType repository.DeleteType, opts ...search.Option) error {
	q := search.New(opts...).Query()
	if err := query.Validate(q, query.MandatoryFilters(fields.ID)); err != nil {
		return err
	}
	if err := r.QueryApply(ctx, q).Delete(&templateEntity{}).Error; err != nil {
		return postgres.NewErrUnknown(err)
	}
	return nil
}

// GetForRender resolves the template for (realm, name, locale), falling back to
// the realm's default-locale template when the requested locale is absent.
func (r *templateRepo) GetForRender(ctx context.Context, realmID, name, locale string) (domain.NotificationTemplate, error) {
	t, err := r.byLocale(ctx, realmID, name, locale)
	if err == nil {
		return t, nil
	}
	if !apierrors.Is(err, apierrors.CodeNotFound) || locale == domain.DefaultLocale {
		return nil, err
	}
	return r.byLocale(ctx, realmID, name, domain.DefaultLocale)
}

func (r *templateRepo) byLocale(ctx context.Context, realmID, name, locale string) (domain.NotificationTemplate, error) {
	return r.Get(ctx, search.WithQueryOpts(
		query.FilterBy(filter.OpEq, fields.RealmID, realmID),
		query.FilterBy(filter.OpEq, fields.Name, name),
		query.FilterBy(filter.OpEq, fields.Locale, locale),
	))
}
