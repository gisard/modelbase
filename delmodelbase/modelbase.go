package delmodelbase

import (
	"context"
	"reflect"
	"time"

	"github.com/gisard/modelbase"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

type modelBase[K comparable, T DataObjecter[K]] struct {
	modelbase.ModelBase[K, T]
}

func (m *modelBase[K, T]) GetModelBaseDB() modelbase.ModelBase[K, T] {
	return m.ModelBase
}

func (m *modelBase[K, T]) Insert(ctx context.Context, ts ...T) error {
	now := time.Now()
	for _, t := range ts {
		t.SetCreateTime(now)
		t.SetUpdateTime(now)
	}
	return m.ModelBase.Insert(ctx, ts...)
}

func (m *modelBase[K, T]) Upsert(ctx context.Context, ts ...T) error {
	now := time.Now()
	for _, t := range ts {
		if t.GetCreateTime().IsZero() {
			t.SetCreateTime(now)
		}
		t.SetUpdateTime(now)
	}
	return m.ModelBase.Upsert(ctx, ts...)
}

func (m *modelBase[K, T]) Get(ctx context.Context, id K) (T, error) {
	return m.GetBy(ctx, "`id` = ?", id)
}

func (m *modelBase[K, T]) GetWithLock(ctx context.Context, lock modelbase.Lock, id K) (T, error) {
	return m.GetWithLockBy(ctx, lock, "`id` = ?", id)
}

func (m *modelBase[K, T]) GetBy(ctx context.Context, where string, values ...any) (T, error) {
	if where != "" {
		where += " AND "
	}
	where += "`is_deleted` = 0"
	return m.ModelBase.GetBy(ctx, where, values...)
}

func (m *modelBase[K, T]) GetWithLockBy(ctx context.Context, lock modelbase.Lock, where string, values ...any) (T, error) {
	if where != "" {
		where += " AND "
	}
	where += "`is_deleted` = 0"
	return m.ModelBase.GetWithLockBy(ctx, lock, where, values...)
}

func (m *modelBase[K, T]) Update(ctx context.Context, t T) error {
	now := time.Now()
	if t.GetCreateTime().IsZero() {
		t.SetCreateTime(now)
	}
	t.SetUpdateTime(now)
	return m.GetModelBaseDB().Update(ctx, t)
}

func (m *modelBase[K, T]) UpdateAllFields(ctx context.Context, t T) error {
	now := time.Now()
	if t.GetCreateTime().IsZero() {
		t.SetCreateTime(now)
	}
	t.SetUpdateTime(now)
	return m.GetModelBaseDB().UpdateAllFields(ctx, t)
}

func (m *modelBase[K, T]) UpdateBatch(ctx context.Context, params map[string]any, where string, values ...any) error {
	if len(params) == 0 {
		return nil
	}
	if params["update_time"] == nil {
		params["update_time"] = time.Now()
	}
	var t T
	return errors.WithStack(m.GetDB(ctx).Model(&t).Where(where, values...).Updates(params).Error)
}

func (m *modelBase[K, T]) List(ctx context.Context, where string, values ...any) ([]T, error) {
	return m.ListOpts(ctx, modelbase.WhereOpt(where, values...))
}

func (m *modelBase[K, T]) ListMap(ctx context.Context, where string, values ...any) (map[K]T, error) {
	return m.ListMapOpts(ctx, modelbase.WhereOpt(where, values...))
}

func (m *modelBase[K, T]) ListOpts(ctx context.Context, opts ...modelbase.ListOpt) ([]T, error) {
	opts = append(opts, modelbase.WhereOpt("`is_deleted` = 0"))
	return m.ModelBase.ListOpts(ctx, opts...)
}

func (m *modelBase[K, T]) ListMapOpts(ctx context.Context, opts ...modelbase.ListOpt) (map[K]T, error) {
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
	return m.ListOpts(ctx, modelbase.WhereOpt("`id` IN (?)", ids))
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

func (m *modelBase[K, T]) ListIDs(ctx context.Context, where string, values ...any) ([]K, error) {
	if where != "" {
		where += " AND "
	}
	where += "`is_deleted` = 0"
	return m.ModelBase.ListIDs(ctx, where, values...)
}

func (m *modelBase[K, T]) Exist(ctx context.Context, where string, values ...any) (bool, error) {
	if where != "" {
		where += " AND "
	}
	where += "`is_deleted` = 0"
	return m.ModelBase.Exist(ctx, where, values...)
}

func (m *modelBase[K, T]) Count(ctx context.Context, where string, values ...any) (int64, error) {
	return m.CountOpts(ctx, modelbase.WhereOpt(where, values...))
}

func (m *modelBase[K, T]) CountOpts(ctx context.Context, opts ...modelbase.ListOpt) (int64, error) {
	opts = append(opts, modelbase.WhereOpt("`is_deleted` = 0"))
	return m.ModelBase.CountOpts(ctx, opts...)
}

func (m *modelBase[K, T]) Delete(ctx context.Context, id K) error {
	return m.DeleteBatch(ctx, "`id` = ?", id)
}

func (m *modelBase[K, T]) DeleteBatch(ctx context.Context, where string, values ...any) error {
	return m.GetModelBaseDB().UpdateBatch(ctx, map[string]any{"is_deleted": 1}, where, values...)
}

func NewModelBase[K comparable, T DataObjecter[K]](db *gorm.DB) ModelBase[K, T] {
	var t T
	if reflect.TypeOf(t).Kind() != reflect.Pointer {
		panic(errors.Errorf("ModelBase should inject point type: %T", t))
	}
	return &modelBase[K, T]{ModelBase: modelbase.NewModelBase[K, T](db)}
}
