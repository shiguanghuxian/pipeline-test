package common

import (
	"log"

	"github.com/dop251/goja"
	"github.com/dop251/goja_nodejs/require"
	_ "github.com/dop251/goja_nodejs/util"
)

type Console struct {
	runtime *goja.Runtime
	util    *goja.Object
	logger  *TaskLog // 日志对象，会输出到websocket
}

func (c *Console) log(call goja.FunctionCall) goja.Value {
	if format, ok := goja.AssertFunction(c.util.Get("format")); ok {
		ret, err := format(c.util, call.Arguments...)
		if err != nil {
			c.logger.Log("日志输出错误: " + err.Error())
			return nil
		}
		log.Println(ret.String())
	} else {
		c.logger.Log("日志输出错误: util.format is not a function")
	}

	return nil
}

func (c *Console) Require(runtime *goja.Runtime, module *goja.Object) {
	c.util = require.Require(runtime, "util").(*goja.Object)

	o := module.Get("exports").(*goja.Object)
	o.Set("log", c.log)
	o.Set("error", c.log)
	o.Set("warn", c.log)

}

func (c *Console) Enable(runtime *goja.Runtime) {
	c.runtime = runtime
	require.RegisterNativeModule("console", c.Require)
	runtime.Set("console", require.Require(runtime, "console"))
}

// NewConsole 创建日志对象
func NewConsole(logger *TaskLog) *Console {
	return &Console{
		logger: logger,
	}
}
