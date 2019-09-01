package models

import (
	"github.com/jinzhu/gorm"
)

// ProjectModel 项目管理
type ProjectModel struct {
	Id        int      `gorm:"column:id;primary_key" json:"id"`     //
	Name      string   `gorm:"column:name" json:"name"`             // 项目名
	Desc      string   `gorm:"column:desc" json:"desc"`             // 简述
	BaseUrl   string   `gorm:"column:base_url" json:"base_url"`     // 项目api跟路径
	CreatedAt JSONTime `gorm:"column:created_at" json:"created_at"` // 创建时间
	UpdatedAt JSONTime `gorm:"column:updated_at" json:"updated_at"` // 修改时间

}

// TableName 获取表名
func (ProjectModel) TableName() string {
	return gorm.DefaultTableNameHandler(nil, "project")
}

func (m *ProjectModel) FirstById(id int) (err error) {
	err = client.Table(m.TableName()).Where("id = ?", id).First(m).Error
	return
}
