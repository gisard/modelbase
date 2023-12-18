package modelbase

import (
	"context"

	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

type DataObjecter[K comparable] interface {
	schema.Tabler
	GetID() K
}

type ModelBase[K comparable, T DataObjecter[K]] interface {
	GetDB() *gorm.DB

	InsertBatch(ctx context.Context, ts ...T) error
	Upsert(ctx context.Context, t T) error
	Get(ctx context.Context, id K) (T, error)
	GetWithLock(ctx context.Context, lock Lock, id K) (T, error)
	GetBy(ctx context.Context, where string, values ...any) (T, error)
	GetWithLockBy(ctx context.Context, lock Lock, where string, values ...any) (T, error)
	Update(ctx context.Context, t T) error
	UpdateBatch(ctx context.Context, params map[string]any, where string, values ...any) error
	List(ctx context.Context, listOpts ...ListOpt) ([]T, error)
	ListMap(ctx context.Context, where string, values ...any) (map[K]T, error)
	ListMapByIDs(ctx context.Context, ids []K) (map[K]T, error)
	Exist(ctx context.Context, where string, values ...any) (bool, error)
	Count(ctx context.Context, where string, values ...any) (int64, error)
	Delete(ctx context.Context, t T) error
	DeleteBatch(ctx context.Context, where string, values ...any) error
}

type ListOpt func(*listOpt)

func PageOpt(pageNo, pageSize int) ListOpt {
	return func(opt *listOpt) {
		opt.offset = (pageNo - 1) * pageSize
		opt.limit = pageSize
	}
}

func SortOpt(sortField string, sort Sort) ListOpt {
	return func(opt *listOpt) {
		opt.orders = append(opt.orders, order{
			sortField: sortField,
			sort:      sort,
		})
	}
}

func OffsetOpt(offset, limit int) ListOpt {
	return func(opt *listOpt) {
		opt.offset = offset
		opt.limit = limit
	}
}

func WhereOpt(where string, values ...any) ListOpt {
	return func(opt *listOpt) {
		opt.where = where
		opt.values = values
	}
}

type listOpt struct {
	offset int
	limit  int
	orders []order
	where  string
	values []any
}

type order struct {
	sortField string
	sort      Sort
}
