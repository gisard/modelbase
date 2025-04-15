package delmodelbase

import (
	"time"
)

type DelObject struct {
	CreateTime time.Time `gorm:"column:create_time;type:datetime"` // 创建时间
	UpdateTime time.Time `gorm:"column:update_time;type:datetime"` // 更新时间
	IsDeleted  bool      `gorm:"column:is_deleted;type:tinyint"`   // 是否删除
}

func (d *DelObject) GetDeletedField() string {
	return "is_deleted"
}

func (d *DelObject) GetCreateTime() time.Time {
	return d.CreateTime
}

func (d *DelObject) SetCreateTime(t time.Time) {
	d.CreateTime = t
}

func (d *DelObject) GetUpdateTime() time.Time {
	return d.UpdateTime
}

func (d *DelObject) SetUpdateTime(t time.Time) {
	d.UpdateTime = t
}
