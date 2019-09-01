package common

import (
	"io/ioutil"
	"log"
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
	// TODO 日志写每次生命周期
}

// js调用设置头信息
func (c *JsReq) header(call goja.FunctionCall) goja.Value {
	// 读取调用参数
	if len(call.Arguments) != 2 {
		log.Println("参数必须是两个")
		return c.runtime.ToValue(false)
	}
	key := call.Arguments[0].String()
	val := call.Arguments[1].String()
	// log.Println(key, val)
	// 设置头信息
	c.req.Header.Set(key, val)
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
		log.Println("参数cookie名不能为空")
		return c.runtime.ToValue("")
	}
	for _, v := range c.httpTool.Cookies {
		if call.Arguments[0].String() == v.Name {
			if v == nil {
				return c.runtime.ToValue("")
			}
			return c.runtime.ToValue(v.Value)
		}
	}
	return c.runtime.ToValue("")
}

// 第一个参数为body内容，第二参数为请求体类型 0 json 1 from 默认0
func (c *JsReq) setBody(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) < 1 {
		log.Println("参数body必传")
		return c.runtime.ToValue(false)
	}
	// 如果body是form类型，则必须传递map类型参数
	if len(call.Arguments) >= 2 && call.Arguments[1].String() == "1" {
		// TODO 实现表单
	} else {
		bodyStr := call.Arguments[0].String()
		c.req.Body = ioutil.NopCloser(strings.NewReader(bodyStr))
	}
	return c.runtime.ToValue(true)
}

// 设置BasicAuth
func (c *JsReq) setBasicAuth(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) < 2 {
		log.Println("参数用户名和密码不能为空")
		return c.runtime.ToValue(false)
	}
	username := call.Arguments[0].String()
	password := call.Arguments[1].String()
	c.req.SetBasicAuth(username, password)
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
func NewJsReq(req *http.Request, httpTool *HttpTool) *JsReq {
	return &JsReq{
		req:      req,
		httpTool: httpTool,
	}
}
