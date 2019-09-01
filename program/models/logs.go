package models

import (
	"github.com/jinzhu/gorm"
)

// LogsModel 测试日志 每个接口记录一条数据
type LogsModel struct {
	Id         int      `gorm:"column:id;primary_key" json:"id"`       //
	TaskId     string   `gorm:"column:task_id" json:"task_id"`         // 某次任务id 格式20060102150405
	PipelineId int      `gorm:"column:pipeline_id" json:"pipeline_id"` // 流水线id
	ApiId      int      `gorm:"column:api_id" json:"api_id"`           // api 接口id
	Req        string   `gorm:"column:req" json:"req"`                 // 请求参数
	Rsp        string   `gorm:"column:rsp" json:"rsp"`                 // 响应body
	HttpCode   int      `gorm:"column:http_code" json:"http_code"`     // http状态码
	Expect     int      `gorm:"column:expect" json:"expect"`           // 预期结果0 不符合预期 1符合预期
	Duration   string   `gorm:"column:duration" json:"duration"`       // 请求时长
	CreatedAt  JSONTime `gorm:"column:created_at" json:"created_at"`   // 创建时间
	UpdatedAt  JSONTime `gorm:"column:updated_at" json:"updated_at"`   // 修改时间
}

// TableName 获取表名
func (LogsModel) TableName() string {
	return gorm.DefaultTableNameHandler(nil, "logs")
}

func (m *LogsModel) Save() (err error) {
	err = client.Table(m.TableName()).Save(m).Error
	return
}
