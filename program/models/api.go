package models

import (
	"github.com/jinzhu/gorm"
)

// ApiModel 项目下的api列表
type ApiModel struct {
	Id        int      `gorm:"column:id;primary_key" json:"id"`     //
	Uri       string   `gorm:"column:uri" json:"uri"`               // 请求地址uri
	Name      string   `gorm:"column:name" json:"name"`             // 接口名
	Method    string   `gorm:"column:method" json:"method"`         // 请求Method方法类型
	BodyType  int      `gorm:"column:body_type" json:"body_type"`   // 1 json 2 from
	GroupId   int      `gorm:"column:group_id" json:"group_id"`     // 分组id
	CreatedAt JSONTime `gorm:"column:created_at" json:"created_at"` // 创建时间
	UpdatedAt JSONTime `gorm:"column:updated_at" json:"updated_at"` // 修改时间

}

// TableName 获取表名
func (ApiModel) TableName() string {
	return gorm.DefaultTableNameHandler(nil, "api")
}

// QueryByIds 根据id列表查询api列表
func (m *ApiModel) QueryByIds(ids []int) (list []*ApiModel, err error) {
	err = client.Table(m.TableName()).Where("id in (?)", ids).Scan(&list).Error
	return
}
