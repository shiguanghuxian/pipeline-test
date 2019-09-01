package common

import (
	"io/ioutil"
	"log"
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
	// TODO 日志写每次生命周期
}

// js调用获取响应头信息
func (c *JsRes) getHeader(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) < 1 {
		log.Println("参数cookie名不能为空")
		return c.runtime.ToValue("")
	}
	val := c.res.Header.Get(call.Arguments[0].String())
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
		log.Println("参数cookie名不能为空")
		return c.runtime.ToValue("")
	}
	for _, v := range c.res.Cookies() {
		if call.Arguments[0].String() == v.Name {
			if v == nil {
				return c.runtime.ToValue("")
			}
			return c.runtime.ToValue(v.Value)
		}
	}
	return c.runtime.ToValue("")
}

// getBody js获取响应数据
func (c *JsRes) getBody(call goja.FunctionCall) goja.Value {
	body, _ := ioutil.ReadAll(c.res.Body)
	return c.runtime.ToValue(string(body))
}

// getBody js获取响应数据
func (c *JsRes) getStatusCode(call goja.FunctionCall) goja.Value {
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
func NewJsRes(res *http.Response, httpTool *HttpTool) *JsRes {
	return &JsRes{
		res:      res,
		httpTool: httpTool,
	}
}
