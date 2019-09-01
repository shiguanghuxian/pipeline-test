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
	// TODO task执行日志
}

func (c *Console) log(call goja.FunctionCall) goja.Value {
	if format, ok := goja.AssertFunction(c.util.Get("format")); ok {
		ret, err := format(c.util, call.Arguments...)
		if err != nil {
			panic(err)
		}

		log.Print(ret.String())
	} else {
		panic(c.runtime.NewTypeError("util.format is not a function"))
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
func NewConsole() *Console {
	return &Console{}
}
