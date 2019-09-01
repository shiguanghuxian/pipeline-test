package v1

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/dop251/goja"
	"github.com/dop251/goja_nodejs/require"
	"github.com/shiguanghuxian/pipeline-test/program/logger"

	gin "github.com/gin-gonic/gin"
	"github.com/shiguanghuxian/pipeline-test/program/common"
	"github.com/shiguanghuxian/pipeline-test/program/models"
)

/* 流水线任务 */

type PipelineTaskController struct {
}

// RunTask 运行一个测试流水线
func (api *PipelineTaskController) RunTask(c *gin.Context) {
	pipelineId := common.GetHttpToInt(c, "pipeline_id")
	projectId := common.GetHttpToInt(c, "project_id")
	if pipelineId == 0 || projectId == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}
	// 查询此流水线task列表
	pipelineTasks, err := new(models.PipelineTaskModel).QueryByPipelineId(pipelineId)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}
	// 获取api列表
	apiIds := make([]int, 0)
	for _, v := range pipelineTasks {
		apiIds = append(apiIds, v.ApiId)
	}
	apis, err := new(models.ApiModel).QueryByIds(apiIds)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}
	apiMap := make(map[int]*models.ApiModel, 0)
	for _, v := range apis {
		apiMap[v.Id] = v
	}
	// 查询项目信息
	projectInfo := new(models.ProjectModel)
	err = projectInfo.FirstById(projectId)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}
	// 遍历执行api测试
	httpTool := common.NewHttpTool(projectInfo.BaseUrl)
	// logBuffer := bytes.NewBuffer(nil) // TODO 记录执行日志 开始请求xxx等
	for _, v := range pipelineTasks {
		apiOne, ok := apiMap[v.ApiId]
		if ok == false {
			logger.Log.Warnw("流水线中不存在api", "api_id", v.ApiId)
			continue
		}
		proper, err := api.runOneTask(httpTool, v, apiOne)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err})
			return
		}
		if proper == false {
			c.JSON(http.StatusBadRequest, gin.H{"error": "预期验证返回false"})
			return
		}
	}
	log.Println(pipelineTasks)
	c.JSON(http.StatusOK, "ok")
}

const (
	DefaultExpect = `
	function expect() {
		var code = response.getStatusCode()
		if (code == 200) {
			return true
		}
		return false
	}
	expect()
	`
)

func (api *PipelineTaskController) runOneTask(httpTool *common.HttpTool, pipelineTask *models.PipelineTaskModel, apiOne *models.ApiModel) (proper bool, err error) {
	if pipelineTask == nil || apiOne == nil {
		err = errors.New("参数不能为nil")
		proper = false
		return
	}

	// 构建js解析器
	vm := goja.New()
	new(require.Registry).Enable(vm)
	// 日志对象
	common.NewConsole().Enable(vm)

	// 设置请求前后钩子
	httpTool.SetBefore(api.caHttpBefore(vm, pipelineTask.Before))
	httpTool.SetAfter(api.caHttpAfter(vm, pipelineTask.After))
	log.Println("开始请求接口", apiOne.Uri)

	// 执行编号
	taskId := fmt.Sprint(time.Now().UnixNano())
	var httpCode int
	var body []byte
	start := time.Now()

	// 执行一个请求
	if apiOne.Method == "POST" {
		body, httpCode, err = httpTool.Post(apiOne.Uri, pipelineTask.Req, nil, 1)
	} else if apiOne.Method == "GET" {
		body, httpCode, err = httpTool.Get(apiOne.Uri, pipelineTask.Req, nil)
	} else {
		log.Println("不支持的请求类型")
		err = errors.New("不支持的请求类型")
		proper = false
		return
	}
	end := time.Now()

	// 预期验证逻辑
	if pipelineTask.Expect == "" {
		// 默认验证状态码
		pipelineTask.Expect = DefaultExpect
	}
	v, err := vm.RunString(pipelineTask.Expect)
	if err != nil {
		log.Println("运行预期js错误", err.Error())
	}
	obj := v.Export()
	if val, ok := obj.(bool); ok {
		if val == true {
			log.Println("执行预期js返回true")
		} else {
			log.Println("执行预期js返回false")
		}
	} else {
		log.Println("执行预期未返回true或false js结果: ", obj)
		err = fmt.Errorf("执行预期未返回true或false js结果: %v", obj)
		proper = false
		return
	}

	// 写执行日志
	now := models.JSONTime(time.Now())
	apiLog := &models.LogsModel{
		TaskId:     taskId,
		PipelineId: pipelineTask.Id,
		ApiId:      apiOne.Id,
		Req:        pipelineTask.Req,
		Rsp:        string(body),
		HttpCode:   httpCode,
		Expect:     1,
		Duration:   end.Sub(start).String(),
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	err = apiLog.Save()
	if err != nil {
		logger.Log.Errorw("保存日志错误", "err", err)
		proper = false
		return
	}
	return
}

// 请求之前做一件事
func (api *PipelineTaskController) caHttpBefore(vm *goja.Runtime, js string) common.BeforeFunc {
	if len(js) == 0 {
		log.Println("调用请求之前钩子，js代码段为空")
		return common.DefaultBeforeFunc
	}
	return func(tool *common.HttpTool, req *http.Request) {
		// 构造js执行对象
		jsReq := common.NewJsReq(req, tool)
		jsReq.Enable(vm)

		v, err := vm.RunString(js)
		if err != nil {
			log.Println("运行js错误", err.Error())
		}
		obj := v.Export()
		if val, ok := obj.(bool); ok {
			if val == true {
				log.Println("执行js返回true")
			} else {
				log.Println("执行js返回false")
			}
		} else {
			log.Println("执行js结果: ", obj)
		}
	}
}

// 请求之后做一件事
func (api *PipelineTaskController) caHttpAfter(vm *goja.Runtime, js string) common.AfterFunc {
	if len(js) == 0 {
		log.Println("调用请求之前钩子，js代码段为空")
		return common.DefaultAfterFunc
	}
	return func(tool *common.HttpTool, req *http.Response) {
		// 构造js执行对象
		jsRes := common.NewJsRes(req, tool)
		jsRes.Enable(vm)

		v, err := vm.RunString(js)
		if err != nil {
			log.Println("运行js错误", err.Error())
		}
		obj := v.Export()
		if val, ok := obj.(bool); ok {
			if val == true {
				log.Println("执行js返回true")
			} else {
				log.Println("执行js返回false")
			}
		} else {
			log.Println("执行js结果: ", obj)
		}
	}
}
