package common

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/dop251/goja"
	"github.com/dop251/goja_nodejs/require"
	_ "github.com/dop251/goja_nodejs/util"
)

/* 用于设置请求前后和预期判断的js库 */

// JsReq 请求钩子
type JsReq struct {
	runtime  *goja.Runtime
	util     *goja.Object
	req      *http.Request
	httpTool *HttpTool
	logger   *TaskLog // 日志对象，会输出到websocket
}

// js调用设置头信息
func (c *JsReq) header(call goja.FunctionCall) goja.Value {
	// 读取调用参数
	if len(call.Arguments) != 2 {
		c.logger.Log("设置头信息，参数必须是2个")
		return c.runtime.ToValue(false)
	}
	key := call.Arguments[0].String()
	val := call.Arguments[1].String()
	// log.Println(key, val)
	// 设置头信息
	c.req.Header.Set(key, val)
	c.logger.Log("头信息设置成功")
	return c.runtime.ToValue(true)
}

// js 调用获取cookie列表
func (c *JsReq) cookies(call goja.FunctionCall) goja.Value {
	cookieMap := make(map[string]string, 0)
	list := c.httpTool.Cookies
	for _, v := range list {
		cookieMap[v.Name] = v.Value
	}
	return c.runtime.ToValue(cookieMap)
}

// js 调用获取一个cookie
func (c *JsReq) getCookie(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) < 1 {
		c.logger.Log("获取cookie，参数cookie名不能为空")
		return c.runtime.ToValue("")
	}
	name := call.Arguments[0].String()
	for _, v := range c.httpTool.Cookies {
		if name == v.Name {
			c.logger.Log(fmt.Sprintf("获取cookie值 %s:%s", name, v.Value))
			return c.runtime.ToValue(v.Value)
		}
	}
	c.logger.Log(fmt.Sprintf("获取cookie值 %s:空", name))
	return c.runtime.ToValue("")
}

// 第一个参数为body内容，第二参数为请求体类型 0 json 1 from 默认0
func (c *JsReq) setBody(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) < 1 {
		c.logger.Log("设置body，参数必传body内容")
		return c.runtime.ToValue(false)
	}
	// 如果body是form类型，则必须传递map类型参数
	if len(call.Arguments) >= 2 && call.Arguments[1].String() == "1" {
		// TODO 实现表单
	} else {
		bodyStr := call.Arguments[0].String()
		c.req.Body = ioutil.NopCloser(strings.NewReader(bodyStr))
	}
	c.logger.Log("body设置成功")
	return c.runtime.ToValue(true)
}

// 设置BasicAuth
func (c *JsReq) setBasicAuth(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) < 2 {
		c.logger.Log("设置BasicAuth，参数用户名和密码不能为空")
		return c.runtime.ToValue(false)
	}
	username := call.Arguments[0].String()
	password := call.Arguments[1].String()
	c.req.SetBasicAuth(username, password)
	c.logger.Log(fmt.Sprintf("设置BasicAuth成功 username:%s password:%s", username, password))
	return c.runtime.ToValue(true)
}

// Require 注册方法
func (c *JsReq) Require(runtime *goja.Runtime, module *goja.Object) {
	c.runtime = runtime
	c.util = require.Require(runtime, "util").(*goja.Object)

	o := module.Get("exports").(*goja.Object)
	o.Set("header", c.header)
	o.Set("cookies", c.cookies)
	o.Set("getCookie", c.getCookie)
}

// Enable 启用扩张
func (c *JsReq) Enable(runtime *goja.Runtime) {
	require.RegisterNativeModule("request", c.Require)
	runtime.Set("request", require.Require(runtime, "request"))
}

// NewJsReq 创建请求插件
func NewJsReq(req *http.Request, httpTool *HttpTool, logger *TaskLog) *JsReq {
	return &JsReq{
		req:      req,
		httpTool: httpTool,
		logger:   logger,
	}
}
