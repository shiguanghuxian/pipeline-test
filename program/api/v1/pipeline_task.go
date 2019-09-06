package v1

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/dop251/goja"
	"github.com/dop251/goja_nodejs/require"
	gin "github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/shiguanghuxian/pipeline-test/program/common"
	"github.com/shiguanghuxian/pipeline-test/program/logger"
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
	taskId := fmt.Sprint(time.Now().UnixNano()) // 本地任务id
	taskLog := common.NewTaskLog(taskId)        // 日志对象
	taskLogMap.Store(taskId, taskLog)           // 记录日志对象到map

	// 逐条执行api测试列表
	go func() {
		for _, v := range pipelineTasks {
			apiOne, ok := apiMap[v.ApiId]
			if ok == false {
				logger.Log.Warnw("流水线中不存在api", "api_id", v.ApiId)
				taskLog.Log(fmt.Sprintf("流水线中不存在api api_id:%d", v.ApiId))
				continue
			}
			proper, err := api.runOneTask(httpTool, v, apiOne, taskLog, taskId)
			if err != nil {
				logger.Log.Errorw("终止执行，调用接口遇到错误", "err", err)
				taskLog.Log(fmt.Sprintf("终止执行，调用接口遇到错误 err:%s", err.Error()))
				return
			}
			if proper == false {
				logger.Log.Errorw("终止执行，执行一个api调用未达到预期", "err", err)
				taskLog.Log(fmt.Sprintf("终止执行，执行一个api调用未达到预期 name: %s pipeline_task_id:%d", v.Name, v.Id))
				return
			}
		}
		// 归档所有日志到MySQL
		logs := taskLog.AllLogs()
		logsBytes, _ := json.Marshal(logs)
		err = (&models.TaskLogPlaceModel{
			TaskId:     taskId,
			PipelineId: pipelineId,
			Log:        string(logsBytes),
		}).Save()
		if err != nil {
			logger.Log.Errorw("归档日志到数据库错误", "err", err)
			taskLog.Log("归档日志到数据库错误: " + err.Error())
		}
		// 执行完毕关闭日志
		taskLog.Close()
	}()

	c.JSON(http.StatusOK, gin.H{
		"task_id": taskId,
		"tasks":   pipelineTasks,
	})
}

const (
	// 默认断言判断
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

var (
	taskLogMap = new(sync.Map) // 存储task日志对象
)

func (api *PipelineTaskController) runOneTask(httpTool *common.HttpTool, pipelineTask *models.PipelineTaskModel, apiOne *models.ApiModel, taskLog *common.TaskLog, taskId string) (proper bool, err error) {
	if pipelineTask == nil || apiOne == nil {
		err = errors.New("runOneTask 参数不能为nil")
		proper = false
		return
	}
	time.Sleep(time.Second / 2)
	// 构建js解析器
	vm := goja.New()
	new(require.Registry).Enable(vm)
	// 日志对象
	common.NewConsole(taskLog).Enable(vm)

	// 设置请求前后钩子
	httpTool.SetBefore(api.caHttpBefore(vm, pipelineTask.Before, taskLog))
	httpTool.SetAfter(api.caHttpAfter(vm, pipelineTask.After, taskLog))
	log.Println("开始请求接口", apiOne.Uri)

	// 执行编号

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
	if err != nil {
		return
	}

	// 预期验证逻辑
	if pipelineTask.Expect == "" {
		// 默认验证状态码
		pipelineTask.Expect = DefaultExpect
	}
	v, err := vm.RunString(pipelineTask.Expect)
	if err != nil {
		log.Println("运行预期js错误", err.Error())
		proper = false
		taskLog.Log(fmt.Sprintf("计算预期，运行预期js错误 name: %s", pipelineTask.Name))
	}
	expect := 0 // 是否达到预期
	if v == nil {
		log.Println("执行预期js返回 v=nil")
		proper = false
		taskLog.Log(fmt.Sprintf("计算预期，运行js返回 v=nil name: %s", pipelineTask.Name))
	} else {
		obj := v.Export()
		if val, ok := obj.(bool); ok {
			if val == true {
				log.Println("执行预期js返回true")
				expect = 1
				proper = true
			} else {
				log.Println("执行预期js返回false")
				proper = false
				taskLog.Log(fmt.Sprintf("计算预期，运行js返回false name: %s", pipelineTask.Name))
			}
			// 发送接口达到预期消息
			expectMsg := make(map[string]interface{}, 0)
			expectMsg["api_id"] = pipelineTask.ApiId
			expectMsg["id"] = pipelineTask.Id
			expectMsg["expect"] = val
			expectMsgBytes, _ := json.Marshal(expectMsg)
			taskLog.Log(string(expectMsgBytes), "task_state")
		} else {
			log.Println("执行预期未返回true或false js结果: ", obj)
			err = fmt.Errorf("执行预期未返回true或false js结果: %v", obj)
			proper = false
			taskLog.Log(fmt.Sprintf("执行预期未返回true或false name: %s", pipelineTask.Name))
			return
		}
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
		Expect:     expect,
		Duration:   end.Sub(start).String(),
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	err = apiLog.Save()
	if err != nil {
		logger.Log.Errorw("保存日志错误", "err", err)
		proper = false
		taskLog.Log(fmt.Sprintf("保存api执行日志错误 name: %s", pipelineTask.Name))
		return
	}
	return
}

// 请求之前做一件事
func (api *PipelineTaskController) caHttpBefore(vm *goja.Runtime, js string, taskLog *common.TaskLog) common.BeforeFunc {
	if len(js) == 0 {
		log.Println("调用请求之前钩子，js代码段为空")
		return common.DefaultBeforeFunc
	}
	return func(tool *common.HttpTool, req *http.Request) {
		// 构造js执行对象
		jsReq := common.NewJsReq(req, tool, taskLog)
		jsReq.Enable(vm)

		v, err := vm.RunString(js)
		if err != nil {
			taskLog.Log(fmt.Sprintf("运行调用前钩子错误 err:%s", err.Error()))
		}
		obj := v.Export()
		if val, ok := obj.(bool); ok {
			if val == true {
				taskLog.Log("运行调用前钩子返回true")
			} else {
				taskLog.Log("运行调用前钩子返回false")
			}
		} else {
			taskLog.Log(fmt.Sprintf("运行调用前钩子返回 %v", obj))
		}
	}
}

// 请求之后做一件事
func (api *PipelineTaskController) caHttpAfter(vm *goja.Runtime, js string, taskLog *common.TaskLog) common.AfterFunc {
	if len(js) == 0 {
		log.Println("调用请求之前钩子，js代码段为空")
		return common.DefaultAfterFunc
	}
	return func(tool *common.HttpTool, req *http.Response) {
		// 构造js执行对象
		jsRes := common.NewJsRes(req, tool, taskLog)
		jsRes.Enable(vm)

		v, err := vm.RunString(js)
		if err != nil {
			log.Println("运行js错误", err.Error())
		}
		obj := v.Export()
		if val, ok := obj.(bool); ok {
			if val == true {
				taskLog.Log("运行调用后钩子返回true")
			} else {
				taskLog.Log("运行调用后钩子返回false")
			}
		} else {
			taskLog.Log(fmt.Sprintf("运行调用后钩子返回 %v", obj))
		}
	}
}

// WsConsumerData websocket 订阅日志消息
type WsConsumerData struct {
	Typ     string `json:"type"` // 消息类型 ping | consumer | unconsumer
	Payload string `json:"payload"`
}

// ConsumerData 订阅消息
type ConsumerData struct {
	TaskId string `json:"task_id"` // task id 订阅的执行查询日志
}

var wsupgrader = websocket.Upgrader{
	ReadBufferSize:   1024,
	WriteBufferSize:  1024,
	HandshakeTimeout: 5 * time.Second,
	// 取消ws跨域校验
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// WsConsumer 订阅task日志输出
func (api *PipelineTaskController) WsConsumer(c *gin.Context) {
	var conn *websocket.Conn
	var err error
	conn, err = wsupgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		logger.Log.Errorw("未成功升级未为websocket", "err", err)
		return
	}
	// 连接关闭时，删除连接订阅
	defer func() {
		taskLogMap.Range(func(key, value interface{}) bool {
			if taskLog, ok := value.(*common.TaskLog); ok {
				taskLog.RemoveConn(conn)
			}
			return true
		})
	}()

	for {
		//读取ws中的数据
		_, message, err := conn.ReadMessage()
		if err != nil {
			logger.Log.Errorw("websocket接收消息错误", "err", err)
			break
		}
		// log.Println("收到消息", string(message))
		// 解析消息
		wsData := new(WsConsumerData)
		err = json.Unmarshal(message, wsData)
		if err != nil {
			logger.Log.Errorw("解析websocket消息错误", "err", err)
			continue
		}
		if wsData.Typ == "ping" {
			// log.Println("收到客户端ping")
		} else if wsData.Typ == "consumer" {
			cData := new(ConsumerData)
			err = json.Unmarshal([]byte(wsData.Payload), cData)
			if err != nil {
				logger.Log.Errorw("订阅消息解析错误", "err", err)
			}
			// 将websocket连接放入日志连接列表
			if taskOne, ok := taskLogMap.Load(cData.TaskId); ok == true {
				if taskLog, ok := taskOne.(*common.TaskLog); ok == true {
					taskLog.AppendConn(conn)
				} else {
					logger.Log.Warnw("taskLogMap取值类型断言粗我")
				}
			} else {
				logger.Log.Warnw("订阅的task_id不存在")
			}
		} else if wsData.Typ == "unconsumer" {
			cData := new(ConsumerData)
			err = json.Unmarshal([]byte(wsData.Payload), cData)
			if err != nil {
				logger.Log.Errorw("取消订阅消息解析错误", "err", err)
			}
		} else {
			logger.Log.Errorw("未知消息类型")
		}
	}
}
