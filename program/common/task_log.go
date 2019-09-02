package common

import (
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// TaskLog 每个测试运行时的日志输出 输出到控制台和websocket
type TaskLog struct {
	content []*WsData         // 存储本次日志全部内容 关闭时清空 - 用于首次订阅消息输出历史日志
	ws      []*websocket.Conn // 要输出的websocket连接列表
	lock    *sync.Mutex
	close   bool // 是否已经关闭 - 关闭后，再有订阅只输出历史日志
}

// WsData ws 输出消息
type WsData struct {
	Timestamp int64  `json:"timestamp"`
	Msg       string `json:"msg"`
}

// NewTaskLog 创建task日志对象
func NewTaskLog() *TaskLog {
	return &TaskLog{
		content: make([]*WsData, 0),
		ws:      make([]*websocket.Conn, 0),
		lock:    &sync.Mutex{},
		close:   false,
	}
}

// Close 关闭日志对象
func (t *TaskLog) Close() error {
	t.lock.Lock()
	defer t.lock.Unlock()
	t.content = nil
	t.ws = nil
	t.close = true
	return nil
}

// AllLogs 获取所有日志
func (t *TaskLog) AllLogs() []*WsData {
	return t.content
}

// Log 输出日志
func (t *TaskLog) Log(msg string) {
	t.lock.Lock()
	defer t.lock.Unlock()

	log.Println(msg)

	now := time.Now().Unix()
	wsData := &WsData{Timestamp: now, Msg: msg}
	t.content = append(t.content, wsData)
	// 输出到websocket
	for _, ws := range t.ws {
		if ws == nil {
			continue
		}
		if len(t.content) > 0 {
			for _, v := range t.content {
				err := ws.WriteJSON(v)
				if err != nil {
					log.Println("写客户端溜错误，停止此客户端写数据1")
					t.removeConn(ws)
					continue
				}
			}
		}
		err := ws.WriteJSON(wsData)
		if err != nil {
			log.Println("写客户端溜错误，停止此客户端写数据2")
			t.removeConn(ws)
			continue
		}
	}
}

// AppendConn 添加websocket到连接列表
func (t *TaskLog) AppendConn(ws *websocket.Conn) {
	t.lock.Lock()
	defer t.lock.Unlock()
	if t.close == true {
		return
	}
	t.ws = append(t.ws, ws)
}

// RemoveConn 移除连接
func (t *TaskLog) RemoveConn(ws *websocket.Conn) {
	t.lock.Lock()
	defer t.lock.Unlock()
	t.removeConn(ws)
}

func (t *TaskLog) removeConn(ws *websocket.Conn) {
	if t.close == true {
		return
	}
	tws := make([]*websocket.Conn, 0)
	for _, v := range t.ws {
		if v != ws {
			tws = append(tws, v)
		}
	}
	t.ws = tws
}
