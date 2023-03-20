# golang常用工具，包括

* [与java兼容的字符串](jcomp/README.md)
* [cityhash golang实现](hash/README.md)
* [常用功能函数集](utils/README.md)
* [golang版本并发库封装](gocc/README.md)
* [tcp通信框架](nfour/README.md)

# 日志支持

logger.Logger接口抽象类saber框架需要的日志功能，提供debug、info、error三种级别的日志输出，缺省实现调用标准库的log包输出，你可以实现自己的日志输出。

```
type Logger interface {
	Debug(format string, v ...any)
	DebugLn(v ...any)

	Info(format string, v ...any)
	InfoLn(v ...any)

	Error(format string, v ...any)
	ErrorLn(v ...any)
}
```