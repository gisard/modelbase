package modelbase

import (
	"context"
	"reflect"
	"strings"

	"github.com/pkg/errors"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type modelBase[K comparable, T DataObjecter[K]] struct {
	db *gorm.DB
}

func (m *modelBase[K, T]) GetDB() *gorm.DB {
	return m.db
}

func (m *modelBase[K, T]) InsertBatch(ctx context.Context, ts ...T) error {
	if len(ts) == 0 {
		return nil
	}
	return errors.WithStack(m.db.WithContext(ctx).Create(ts).Error)
}

func (m *modelBase[K, T]) Upsert(ctx context.Context, t T) error {
	err := m.db.WithContext(ctx).Create(t).Error
	if err != nil {
		if strings.Contains(err.Error(), "Duplicate entry") {
			err = m.db.WithContext(ctx).Updates(t).Error
		}
	}
	return errors.WithStack(err)
}

func (m *modelBase[K, T]) Get(ctx context.Context, id K) (T, error) {
	var object T
	return object, errors.WithStack(m.db.WithContext(ctx).First(&object, id).Error)
}

func (m *modelBase[K, T]) GetWithLock(ctx context.Context, lock Lock, id K) (T, error) {
	var object T
	return object, errors.WithStack(m.db.WithContext(ctx).Clauses(
		clause.Locking{Strength: lock.ToString()}).First(&object, id).Error)
}

func (m *modelBase[K, T]) GetBy(ctx context.Context, where string, values ...any) (T, error) {
	var object T
	return object, errors.WithStack(m.db.WithContext(ctx).
		Where(where, values...).First(&object).Error)
}

func (m *modelBase[K, T]) GetWithLockBy(ctx context.Context, lock Lock, where string, values ...any) (T, error) {
	var object T
	return object, errors.WithStack(m.db.WithContext(ctx).Clauses(
		clause.Locking{Strength: lock.ToString()}).Where(where, values...).First(&object).Error)
}

func (m *modelBase[K, T]) Update(ctx context.Context, t T) error {
	return errors.WithStack(m.db.WithContext(ctx).Updates(t).Error)
}

func (m *modelBase[K, T]) UpdateBatch(ctx context.Context, params map[string]any, where string, values ...any) error {
	var t T
	return errors.WithStack(m.db.WithContext(ctx).Model(&t).Where(where, values...).Updates(params).Error)
}

func (m *modelBase[K, T]) List(ctx context.Context, opts ...ListOpt) ([]T, error) {
	listOpts := &listOpt{}
	for _, opt := range opts {
		opt(listOpts)
	}
	var objects []T
	db := m.db.WithContext(ctx).Where(listOpts.where, listOpts.values...).
		Offset(listOpts.offset).Limit(listOpts.limit)
	for _, order := range listOpts.orders {
		db = db.Order(clause.OrderByColumn{
			Column: clause.Column{Name: order.sortField},
			Desc:   order.sort == DESC,
		})
	}
	return objects, errors.WithStack(db.Find(&objects).Error)
}

func (m *modelBase[K, T]) ListMap(ctx context.Context, where string, values ...any) (map[K]T, error) {
	var objects []T
	err := m.db.WithContext(ctx).Where(where, values...).Find(&objects).Error
	if err != nil {
		return nil, errors.WithStack(err)
	}
	if len(objects) == 0 {
		return nil, nil
	}
	tMap := make(map[K]T)
	for _, object := range objects {
		tMap[object.GetID()] = object
	}
	return tMap, nil
}

func (m *modelBase[K, T]) ListMapByIDs(ctx context.Context, ids []K) (map[K]T, error) {
	var objects []T
	err := m.db.WithContext(ctx).Where("`id` IN (?)", ids).Find(&objects).Error
	if err != nil {
		return nil, errors.WithStack(err)
	}
	if len(objects) == 0 {
		return nil, nil
	}
	tMap := make(map[K]T)
	for _, object := range objects {
		tMap[object.GetID()] = object
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

func (m *modelBase[K, T]) Count(ctx context.Context, where string, values ...any) (int64, error) {
	var (
		count int64
		t     T
	)
	return count, errors.WithStack(m.db.WithContext(ctx).Model(&t).
		Where(where, values...).Count(&count).Error)
}

func (m *modelBase[K, T]) Delete(ctx context.Context, t T) error {
	return errors.WithStack(m.db.WithContext(ctx).Delete(&t).Error)
}

func (m *modelBase[K, T]) DeleteBatch(ctx context.Context, where string, values ...any) error {
	var object T
	return errors.WithStack(m.db.WithContext(ctx).Delete(&object, where, values).Error)
}

func NewModelBase[K comparable, T DataObjecter[K]](db *gorm.DB) ModelBase[K, T] {
	var object T
	if reflect.TypeOf(object).Kind() != reflect.Pointer {
		panic(errors.Errorf("ModelBase should inject point type: %T", object))
	}
	return &modelBase[K, T]{db: db}
}
