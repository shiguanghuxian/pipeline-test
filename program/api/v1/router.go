package v1

import (
	gin "github.com/gin-gonic/gin"
)

// APIV1 v1版接口
type APIV1 struct {
}

// Register 注册路由
func (v1 *APIV1) Register(router *gin.RouterGroup) {
	ptg := router.Group("/keys")
	pipelineTaskController := new(PipelineTaskController)
	ptg.GET("/runTask", pipelineTaskController.RunTask) // 运行流水线
}
