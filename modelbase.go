package modelbase

import (
	"context"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"reflect"
)

type modelBase[K comparable, T DataObjecter[K]] struct {
	db *gorm.DB
}

func (m *modelBase[K, T]) GetDB() *gorm.DB {
	return m.db
}

func (m *modelBase[K, T]) Insert(ctx context.Context, ts ...T) error {
	if len(ts) == 0 {
		return nil
	}
	return errors.WithStack(m.db.WithContext(ctx).Create(ts).Error)
}

func (m *modelBase[K, T]) Upsert(ctx context.Context, t T) error {
	return errors.WithStack(m.db.WithContext(ctx).Save(t).Error)
}

func (m *modelBase[K, T]) Get(ctx context.Context, id K) (T, error) {
	var t T
	if err := m.db.WithContext(ctx).First(&t, id).Error; err != nil {
		// If an error occurs, return an empty object
		var t1 T
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return t1, nil
		}
		return t1, errors.WithStack(err)
	}
	return t, nil
}

func (m *modelBase[K, T]) GetWithLock(ctx context.Context, lock Lock, id K) (T, error) {
	var t T
	if err := m.db.WithContext(ctx).Clauses(
		clause.Locking{Strength: lock.ToString()}).First(&t, id).Error; err != nil {
		// If an error occurs, return an empty object
		var t1 T
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return t1, nil
		}
		return t1, errors.WithStack(err)
	}
	return t, nil
}

func (m *modelBase[K, T]) GetBy(ctx context.Context, where string, values ...any) (T, error) {
	var t T
	if err := m.db.WithContext(ctx).
		Where(where, values...).First(&t).Error; err != nil {
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
	if err := m.db.WithContext(ctx).Clauses(
		clause.Locking{Strength: lock.ToString()}).Where(where, values...).First(&t).Error; err != nil {
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
	return errors.WithStack(m.db.WithContext(ctx).Updates(t).Error)
}

func (m *modelBase[K, T]) UpdateBatch(ctx context.Context, params map[string]any, where string, values ...any) error {
	var t T
	return errors.WithStack(m.db.WithContext(ctx).Model(&t).Where(where, values...).Updates(params).Error)
}

func (m *modelBase[K, T]) List(ctx context.Context, opts ...ListOpt) ([]T, error) {
	var ts []T
	db := m.db.WithContext(ctx)
	for _, opt := range opts {
		db = opt.Apply(db)
	}
	return ts, errors.WithStack(db.Find(&ts).Error)
}

func (m *modelBase[K, T]) ListMap(ctx context.Context, opts ...ListOpt) (map[K]T, error) {
	ts, err := m.List(ctx, opts...)
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
	var ts []T
	err := m.db.WithContext(ctx).Where("`id` IN (?)", ids).Find(&ts).Error
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return ts, nil
}

func (m *modelBase[K, T]) ListMapByIDs(ctx context.Context, ids []K) (map[K]T, error) {
	ts, err := m.ListByIDs(ctx, ids)
	if err != nil {
		return nil, err
	}
	if len(ts) == 0 {
		return nil, nil
	}
	tMap := make(map[K]T)
	for _, t := range ts {
		tMap[t.GetID()] = t
	}
	return tMap, nil
}

func (m *modelBase[K, T]) Exist(ctx context.Context, where string, values ...any) (bool, error) {
	_, err := m.GetBy(ctx, where, values...)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, errors.WithStack(err)
	}
	return true, nil
}

func (m *modelBase[K, T]) Count(ctx context.Context, opts ...ListOpt) (int64, error) {
	var (
		count int64
		t     T
	)
	db := m.db.WithContext(ctx).Model(&t)
	for _, opt := range opts {
		if !opt.IsCountOpt() {
			continue
		}
		db = opt.Apply(db)
	}
	return count, errors.WithStack(db.Count(&count).Error)
}

func (m *modelBase[K, T]) Delete(ctx context.Context, t T) error {
	return errors.WithStack(m.db.WithContext(ctx).Delete(&t).Error)
}

func (m *modelBase[K, T]) DeleteBatch(ctx context.Context, where string, values ...any) error {
	var t T
	return errors.WithStack(m.db.WithContext(ctx).Where(where, values...).Delete(&t).Error)
}

func NewModelBase[K comparable, T DataObjecter[K]](db *gorm.DB) ModelBase[K, T] {
	var t T
	if reflect.TypeOf(t).Kind() != reflect.Pointer {
		panic(errors.Errorf("ModelBase should inject point type: %T", t))
	}
	return &modelBase[K, T]{db: db}
}
