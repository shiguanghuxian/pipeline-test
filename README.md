# 流水线测试平台

## 备注
1. pipeline形式设置测试路线
2. 每步可设置特殊头信息和cookie，且支持从上一步请求拿数据
3. 支持从任意之前步骤获取本次请求参数绑定
4. 调用结果支持验证用例是否通过，验证规则内置几种常用的 预期http code 预期返回数据code
5. 预期结果验证需支持一直嵌入式脚本语言自定义验证
6. 每不请求可设置请求数、并发数、。。。
7. 记录请求日志，提供查询界面，提供测试pipeline执行结果（错误步骤终止运行，切给出日志）
8. 实现日志打印库，将日志输出到web界面，实时查看测试进度

一期只实现流程，不实现压测

req请求参数可以${之前的请求id.xx.xx}
${id.data.xxx_id}

执行时长 - 单位毫秒

控制台输出全局拦截， 重写组件，尝试组件调用控制台输出

生成指定长度随机数文字，测试  中英文
