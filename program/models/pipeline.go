package models

import (
	"github.com/jinzhu/gorm"
)

// PipelineModel 流水线
type PipelineModel struct {
	Id        int      `gorm:"column:id;primary_key" json:"id"`     //
	Name      string   `gorm:"column:name" json:"name"`             // 流水线名
	ProjectId int      `gorm:"column:project_id" json:"project_id"` // 所属项目
	Desc      string   `gorm:"column:desc" json:"desc"`             // 备注
	CreatedAt JSONTime `gorm:"column:created_at" json:"created_at"` // 创建时间
	UpdatedAt JSONTime `gorm:"column:updated_at" json:"updated_at"` // 修改时间

}

// TableName 获取表名
func (PipelineModel) TableName() string {
	return gorm.DefaultTableNameHandler(nil, "pipeline")
}
