package models

import (
	"github.com/jinzhu/gorm"
)

// PipelineTaskModel 流水线任务
type PipelineTaskModel struct {
	Id         int      `gorm:"column:id;primary_key" json:"id"`       //
	ApiId      int      `gorm:"column:api_id" json:"api_id"`           // api接口id
	Name       string   `gorm:"column:name" json:"name"`               // 显示名
	Header     string   `gorm:"column:header" json:"header"`           // 附加http头信息
	Thread     int      `gorm:"column:thread" json:"thread"`           // 1 不启用多线程 8启用多线程
	Task       int      `gorm:"column:task" json:"task"`               // 多线程时的线程数
	Expect     string   `gorm:"column:expect" json:"expect"`           // 预期验证，返回true则为通过，false失败 默认比较http状态码
	RelyOn     int      `gorm:"column:rely_on" json:"rely_on"`         // 依赖本次流水线，之前某步骤的数据，默认上一步
	PipelineId int      `gorm:"column:pipeline_id" json:"pipeline_id"` // 流水线id
	Sort       int      `gorm:"column:sort" json:"sort"`               // 排序
	Req        string   `gorm:"column:req" json:"req"`                 // 请求参数
	Before     string   `gorm:"column:before" json:"before"`           // 请求钩子 js代码段
	After      string   `gorm:"column:after" json:"after"`             // 请求之后钩子
	CreatedAt  JSONTime `gorm:"column:created_at" json:"created_at"`   // 创建时间
	UpdatedAt  JSONTime `gorm:"column:updated_at" json:"updated_at"`   // 修改时间
}

// TableName 获取表名
func (PipelineTaskModel) TableName() string {
	return gorm.DefaultTableNameHandler(nil, "pipeline_task")
}

// QueryByPipelineId 根据流水线id 查询任务列表
func (m *PipelineTaskModel) QueryByPipelineId(pipelineId int) (list []*PipelineTaskModel, err error) {
	err = client.Table(m.TableName()).Where("pipeline_id = ?", pipelineId).Order("sort asc, id asc").Scan(&list).Error
	return
}
