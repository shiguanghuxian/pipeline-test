package models

import (
	"github.com/jinzhu/gorm"
)

// ApiGroupModel api分组
type ApiGroupModel struct {
	Id        int      `gorm:"column:id;primary_key" json:"id"`     //
	Name      string   `gorm:"column:name" json:"name"`             // 分组名
	ParentId  int      `gorm:"column:parent_id" json:"parent_id"`   // 父级id
	CreatedAt JSONTime `gorm:"column:created_at" json:"created_at"` // 创建时间
	UpdatedAt JSONTime `gorm:"column:updated_at" json:"updated_at"` // 修改时间

}

// TableName 获取表名
func (ApiGroupModel) TableName() string {
	return gorm.DefaultTableNameHandler(nil, "api_group")
}
