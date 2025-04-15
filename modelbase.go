package modelbase

import (
	"context"
	"reflect"

	"github.com/pkg/errors"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type modelBase[K comparable, T DataObjecter[K]] struct {
	db *gorm.DB
}

func (m *modelBase[K, T]) GetDB(ctx context.Context) *gorm.DB {
	return m.db.WithContext(ctx)
}

func (m *modelBase[K, T]) Insert(ctx context.Context, ts ...T) error {
	if len(ts) == 0 {
		return nil
	}
	return errors.WithStack(m.GetDB(ctx).Create(ts).Error)
}

// Upsert on primary key and unique index update
func (m *modelBase[K, T]) Upsert(ctx context.Context, ts ...T) error {
	if len(ts) == 0 {
		return nil
	}
	return errors.WithStack(m.GetDB(ctx).Save(ts).Error)
}

func (m *modelBase[K, T]) Get(ctx context.Context, id K) (T, error) {
	return m.GetBy(ctx, "`id` = ?", id)
}

func (m *modelBase[K, T]) GetWithLock(ctx context.Context, lock Lock, id K) (T, error) {
	return m.GetWithLockBy(ctx, lock, "`id` = ?", id)
}

func (m *modelBase[K, T]) GetBy(ctx context.Context, where string, values ...any) (T, error) {
	var t T
	if err := m.GetDB(ctx).
		Where(where, values...).Take(&t).Error; err != nil {
		// If an error occurs, return an empty object
		var t1 T
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return t1, nil
		}
		return t1, errors.WithStack(err)
	}
	return t, nil
}

func (m *modelBase[K, T]) GetWithLockBy(ctx context.Context, lock Lock, where string, values ...any) (T, error) {
	var t T
	if err := m.GetDB(ctx).Clauses(
		clause.Locking{Strength: lock.ToString()}).Where(where, values...).Take(&t).Error; err != nil {
		// If an error occurs, return an empty object
		var t1 T
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return t1, nil
		}
		return t1, errors.WithStack(err)
	}
	return t, nil
}

func (m *modelBase[K, T]) Update(ctx context.Context, t T) error {
	t1, err := m.Get(ctx, t.GetID())
	if err != nil {
		return err
	}
	if !m.checkObjectIsValid(t1) {
		return nil
	}
	return errors.WithStack(m.GetDB(ctx).Save(t).Error)
}

func (m *modelBase[K, T]) checkObjectIsValid(t T) bool {
	v := reflect.ValueOf(t)
	if !v.IsValid() {
		return false
	}
	// If it is a pointer type
	if v.Kind() == reflect.Pointer {
		return !v.IsNil()
	}
	// If it is a non-pointer type, check if it is the zero value of the type
	return !v.IsZero()
}

func (m *modelBase[K, T]) UpdateBatch(ctx context.Context, params map[string]any, where string, values ...any) error {
	if len(params) == 0 {
		return nil
	}
	var t T
	return errors.WithStack(m.GetDB(ctx).Model(&t).Where(where, values...).Updates(params).Error)
}

func (m *modelBase[K, T]) List(ctx context.Context, where string, values ...any) ([]T, error) {
	return m.ListOpts(ctx, WhereOpt(where, values...))
}

func (m *modelBase[K, T]) ListMap(ctx context.Context, where string, values ...any) (map[K]T, error) {
	return m.ListOptsMap(ctx, WhereOpt(where, values...))
}

func (m *modelBase[K, T]) ListOpts(ctx context.Context, opts ...ListOpt) ([]T, error) {
	var ts []T
	db := m.GetDB(ctx)
	for _, opt := range opts {
		db = opt.Apply(db)
	}
	return ts, errors.WithStack(db.Find(&ts).Error)
}

func (m *modelBase[K, T]) ListOptsMap(ctx context.Context, opts ...ListOpt) (map[K]T, error) {
	ts, err := m.ListOpts(ctx, opts...)
	if err != nil {
		return nil, err
	}
	tMap := make(map[K]T)
	for _, t := range ts {
		tMap[t.GetID()] = t
	}
	return tMap, nil
}

func (m *modelBase[K, T]) ListByIDs(ctx context.Context, ids []K) ([]T, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	return m.ListOpts(ctx, WhereOpt("`id` IN (?)", ids))
}

func (m *modelBase[K, T]) ListMapByIDs(ctx context.Context, ids []K) (map[K]T, error) {
	return m.ListOptsMap(ctx, WhereOpt("`id` IN (?)", ids))
}

func (m *modelBase[K, T]) Exist(ctx context.Context, where string, values ...any) (bool, error) {
	t, err := m.GetBy(ctx, where, values...)
	if err != nil {
		return false, errors.WithStack(err)
	}

	return m.checkObjectIsValid(t), nil
}

func (m *modelBase[K, T]) Count(ctx context.Context, opts ...ListOpt) (int64, error) {
	var (
		count int64
		t     T
	)
	db := m.GetDB(ctx).Model(&t)
	for _, opt := range opts {
		if !opt.IsCountOpt() {
			continue
		}
		db = opt.Apply(db)
	}
	return count, errors.WithStack(db.Count(&count).Error)
}

func (m *modelBase[K, T]) Delete(ctx context.Context, id K) error {
	return m.DeleteBatch(ctx, "`id` = ?", id)
}

func (m *modelBase[K, T]) DeleteBatch(ctx context.Context, where string, values ...any) error {
	var t T
	return errors.WithStack(m.GetDB(ctx).Where(where, values...).Delete(&t).Error)
}

func NewModelBase[K comparable, T DataObjecter[K]](db *gorm.DB) ModelBase[K, T] {
	var t T
	if reflect.TypeOf(t).Kind() != reflect.Pointer {
		panic(errors.Errorf("ModelBase should inject point type: %T", t))
	}
	return &modelBase[K, T]{db: db}
}
