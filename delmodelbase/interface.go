package delmodelbase

import (
	"time"

	"github.com/gisard/modelbase"
)

// Soft delete model, users don't need to care about fields like is_deleted, create_time, update_time
// To update these fields, please use methods from GetModelBaseDB
type DataObjecter[K comparable] interface {
	modelbase.DataObjecter[K]
	GetDeletedField() string
	GetCreateTime() time.Time
	SetCreateTime(time.Time)
	GetUpdateTime() time.Time
	SetUpdateTime(time.Time)
}

type ModelBase[K comparable, T DataObjecter[K]] interface {
	modelbase.ModelBase[K, T]
	GetModelBaseDB() modelbase.ModelBase[K, T]
}
