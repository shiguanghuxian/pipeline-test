package common

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"strings"
)

/* http 请求库 */

type HttpTool struct {
	baseUrl string         // 基础路径
	Cookies []*http.Cookie // 上次请求cookie，下次携带上
	before  BeforeFunc     // 请求之前做一件事
	after   AfterFunc      // 响应拦截器
}

// BeforeFunc 请求之前事件钩子
type BeforeFunc func(tool *HttpTool, req *http.Request)

// DefaultBeforeFunc 默认请求之前钩子
var DefaultBeforeFunc = func(tool *HttpTool, req *http.Request) {}

// AfterFunc 请求之前事件钩子
type AfterFunc func(tool *HttpTool, res *http.Response)

// DefaultAfterFunc 默认请求之前钩子
var DefaultAfterFunc = func(tool *HttpTool, res *http.Response) {}

// NewHttpTool 创建一个http请求对象
func NewHttpTool(baseUrl string) *HttpTool {
	return &HttpTool{
		baseUrl: baseUrl,
		before:  DefaultBeforeFunc,
		after:   DefaultAfterFunc,
	}
}

// SetBefore 设置请求之前做的事
func (tool *HttpTool) SetBefore(before BeforeFunc) {
	tool.before = before
}

// SetAfter 设置请求之前做的事
func (tool *HttpTool) SetAfter(after AfterFunc) {
	tool.after = after
}

// Request 统一处理请求
func (tool *HttpTool) Request(req *http.Request, header map[string]string) (body []byte, httpCode int, err error) {
	// 请求钩子
	tool.before(tool, req)
	// 设置cookies
	for _, v := range tool.Cookies {
		req.AddCookie(v)
	}
	// 设置头信息
	if header != nil {
		for k, v := range header {
			req.Header.Add(k, v)
		}
	}
	// 发送请求
	res, err := http.DefaultClient.Do(req)
	defer res.Body.Close()
	if err != nil {
		return
	}
	// 读取body信息
	body, err = ioutil.ReadAll(res.Body)
	if err != nil {
		return
	}
	res.Body = ioutil.NopCloser(bytes.NewReader(body))
	// 读取cookie
	tool.Cookies = res.Cookies()
	// 响应钩子
	tool.after(tool, res)
	// http状态码
	httpCode = res.StatusCode
	return
}

// Post post请求
func (tool *HttpTool) Post(uri string, data string, header map[string]string, bodyType int) (body []byte, httpCode int, err error) {
	url := tool.baseUrl + uri
	payload := strings.NewReader(data)
	req, err := http.NewRequest("POST", url, payload)
	if err != nil {
		return
	}
	// 请求体类型
	if bodyType == 1 {
		req.Header.Add("content-type", "application/json")
	} else {
		req.Header.Add("content-type", "multipart/form-data")
	}
	// 发送请求
	body, httpCode, err = tool.Request(req, header)
	return
}

// Get 请求
func (tool *HttpTool) Get(uri string, data string, header map[string]string) (body []byte, httpCode int, err error) {
	url := tool.baseUrl + uri + "?" + data
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return
	}
	// 发送请求
	body, httpCode, err = tool.Request(req, header)
	return
}

// Put put请求
func (tool *HttpTool) Put(uri string, data string, header map[string]string, bodyType int) (body []byte, httpCode int, err error) {
	url := tool.baseUrl + uri
	payload := strings.NewReader(data)
	req, err := http.NewRequest("PUT", url, payload)
	if err != nil {
		return
	}
	// 请求体类型
	if bodyType == 1 {
		req.Header.Add("content-type", "application/json")
	} else {
		req.Header.Add("content-type", "multipart/form-data")
	}
	// 发送请求
	body, httpCode, err = tool.Request(req, header)
	return
}

// Delete 请求
func (tool *HttpTool) Delete(uri string, data string, header map[string]string) (body []byte, httpCode int, err error) {
	url := tool.baseUrl + uri + "?" + data
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return
	}
	// 发送请求
	body, httpCode, err = tool.Request(req, header)
	return
}
