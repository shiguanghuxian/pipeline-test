package common

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/dop251/goja"
	"github.com/dop251/goja_nodejs/require"
	_ "github.com/dop251/goja_nodejs/util"
)

/* 用于设置请求前后和预期判断的js库 */

// JsRes 响应钩子
type JsRes struct {
	runtime  *goja.Runtime
	util     *goja.Object
	res      *http.Response
	httpTool *HttpTool
	logger   *TaskLog // 日志对象，会输出到websocket
}

// js调用获取响应头信息
func (c *JsRes) getHeader(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) < 1 {
		c.logger.Log("获取头信息，参数header名不能为空")
		return c.runtime.ToValue("")
	}
	name := call.Arguments[0].String()
	val := c.res.Header.Get(name)
	c.logger.Log(fmt.Sprintf("获取头信息成功 %s:%s", name, val))
	return c.runtime.ToValue(val)
}

// js 调用获取cookie列表
func (c *JsRes) cookies(call goja.FunctionCall) goja.Value {
	cookieMap := make(map[string]string, 0)
	for _, v := range c.res.Cookies() {
		cookieMap[v.Name] = v.Value
	}
	return c.runtime.ToValue(cookieMap)
}

// js 调用获取一个cookie
func (c *JsRes) getCookie(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) < 1 {
		c.logger.Log("获取cookie，参数cookie名不能为空")
		return c.runtime.ToValue("")
	}
	name := call.Arguments[0].String()
	for _, v := range c.res.Cookies() {
		if name == v.Name {
			c.logger.Log(fmt.Sprintf("获取cookie成功 %s:%s", name, v.Value))
			return c.runtime.ToValue(v.Value)
		}
	}
	c.logger.Log(fmt.Sprintf("获取cookie成功 %s:空", name))
	return c.runtime.ToValue("")
}

// getBody js获取响应数据
func (c *JsRes) getBody(call goja.FunctionCall) goja.Value {
	body, _ := ioutil.ReadAll(c.res.Body)
	c.res.Body = ioutil.NopCloser(bytes.NewReader(body))
	// c.logger.Log("获取body成功: " + string(body))
	return c.runtime.ToValue(string(body))
}

// getBody js获取响应数据
func (c *JsRes) getStatusCode(call goja.FunctionCall) goja.Value {
	c.logger.Log(fmt.Sprintf("获取状态码成功: %d", c.res.StatusCode))
	return c.runtime.ToValue(c.res.StatusCode)
}

// Require 注册方法
func (c *JsRes) Require(runtime *goja.Runtime, module *goja.Object) {
	c.runtime = runtime
	c.util = require.Require(runtime, "util").(*goja.Object)

	o := module.Get("exports").(*goja.Object)
	o.Set("getHeader", c.getHeader)
	o.Set("cookies", c.cookies)
	o.Set("getCookie", c.getCookie)
	o.Set("getBody", c.getBody)
	o.Set("getStatusCode", c.getStatusCode)
}

// Enable 启用扩张
func (c *JsRes) Enable(runtime *goja.Runtime) {
	require.RegisterNativeModule("response", c.Require)
	runtime.Set("response", require.Require(runtime, "response"))
}

// NewJsRes 创建请求插件
func NewJsRes(res *http.Response, httpTool *HttpTool, logger *TaskLog) *JsRes {
	return &JsRes{
		res:      res,
		httpTool: httpTool,
		logger:   logger,
	}
}
