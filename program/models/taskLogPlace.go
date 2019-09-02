package models

import (
	"github.com/jinzhu/gorm"
)

// TaskLogPlaceModel Task执行完后日志归档表
type TaskLogPlaceModel struct {
	Id         int    `gorm:"column:id;primary_key" json:"id"`       //
	TaskId     string `gorm:"column:task_id" json:"task_id"`         // task生成的id
	PipelineId int    `gorm:"column:pipeline_id" json:"pipeline_id"` // 流水线id
	Log        string `gorm:"column:log" json:"log"`                 // 日志内容

}

// TableName 获取表名
func (TaskLogPlaceModel) TableName() string {
	return gorm.DefaultTableNameHandler(nil, "task_log_place")
}

// Save 保存数据
func (m *TaskLogPlaceModel) Save() (err error) {
	err = client.Table(m.TableName()).Save(m).Error
	return
}
