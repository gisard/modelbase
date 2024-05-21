package modelbase

import (
	"context"
	"gorm.io/gorm/clause"

	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

type DataObjecter[K comparable] interface {
	schema.Tabler
	GetID() K
}

type ModelBase[K comparable, T DataObjecter[K]] interface {
	GetDB(ctx context.Context) *gorm.DB

	Insert(ctx context.Context, ts ...T) error
	Upsert(ctx context.Context, t T) error
	Get(ctx context.Context, id K) (T, error)
	GetWithLock(ctx context.Context, lock Lock, id K) (T, error)
	GetBy(ctx context.Context, where string, values ...any) (T, error)
	GetWithLockBy(ctx context.Context, lock Lock, where string, values ...any) (T, error)
	Update(ctx context.Context, t T) error
	UpdateBatch(ctx context.Context, params map[string]any, where string, values ...any) error
	List(ctx context.Context, opts ...ListOpt) ([]T, error)
	ListMap(ctx context.Context, opts ...ListOpt) (map[K]T, error)
	ListByIDs(ctx context.Context, ids []K) ([]T, error)
	ListMapByIDs(ctx context.Context, ids []K) (map[K]T, error)
	Exist(ctx context.Context, where string, values ...any) (bool, error)
	Count(ctx context.Context, opts ...ListOpt) (int64, error)
	Delete(ctx context.Context, t T) error
	DeleteBatch(ctx context.Context, where string, values ...any) error
}

type ListOpt interface {
	Apply(db *gorm.DB) *gorm.DB
	IsCountOpt() bool
}

type pageOpt struct {
	offset int
	limit  int
}

func (l *pageOpt) Apply(db *gorm.DB) *gorm.DB {
	return db.Offset(l.offset).Limit(l.limit)
}

func (l *pageOpt) IsCountOpt() bool {
	return false
}

type sortOpt struct {
	sortField string
	sort      Sort
}

func (s *sortOpt) Apply(db *gorm.DB) *gorm.DB {
	return db.Order(clause.OrderByColumn{
		Column: clause.Column{Name: s.sortField},
		Desc:   s.sort == DESC,
	})
}

func (s *sortOpt) IsCountOpt() bool {
	return false
}

type whereOpt struct {
	where  string
	values []any
}

func (w *whereOpt) Apply(db *gorm.DB) *gorm.DB {
	return db.Where(w.where, w.values...)
}

func (w *whereOpt) IsCountOpt() bool {
	return true
}

func PageOpt(pageNo, pageSize int) ListOpt {
	return &pageOpt{
		offset: (pageNo - 1) * pageSize,
		limit:  pageSize,
	}
}

func SortOpt(sortField string, sort Sort) ListOpt {
	return &sortOpt{
		sortField: sortField,
		sort:      sort,
	}
}

func WhereOpt(where string, values ...any) ListOpt {
	return &whereOpt{
		where:  where,
		values: values,
	}
}
