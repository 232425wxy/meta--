# 日志记录器

## 概述

`log`包提供了**6**种级别的日志记录模式，分别是：

- Trace
- Debug
- Info
- Warn
- Error
- Crit

以上六个级别的日志等级从上到下逐渐递增。

此外，还支持将日志信息重定向到两种输出通道：

- 控制台
- 文件

最后，支持三种日志输出格式：

- 控制台格式
- JSON格式
- 普通日志格式

## 使用方法

### 控制台格式输出日志

如果我们是在包外使用本包中定义的日志记录器，首先需要导入本包，然后按照下面的代码实例化一个日志记录器：

```go
l := log.New("blockchain", "meta--")
l.SetHandler(StreamHandler(os.Stdout, TerminalFormat(true)))
```

上面代码里的`"blockchain"`和`"meta+"`作为是一对键值对，以后每次使用`logger`输出日志时，都会打印这对键值对，然后第二行代码是用来设置输出
日志的处理器，这里我们设置将日志信息输出到操作系统的标准输出里，并且以控制台显示的格式输出，然后对于不同日志等级还会显式不同的颜色：

```go
l.Info("start service")
```

>输出：
>
>INFO*[01-01|00:00:00.000] start service                            blockchain=meta--
>
>ERROR[01-01|00:00:00.000] start service                            blockchain=meta--

### JSON格式输出日志

实例化一个以JSON格式输出日志信息的日志记录器：

```go
l := New("blockchain", "meta--")
l.SetHandler(StreamHandler(os.Stdout, JSONFormat()))
l.Info("start service")
l.Error("start service")
```

>输出
>
>{"blockchain":"meta--","level":"info*","msg":"start service","time":"0001-01-01T00:00:00Z"}
>
>{"blockchain":"meta--","level":"error","msg":"start service","time":"0001-01-01T00:00:00Z"}

### 普通日志格式

`LogfmtFormat()`函数定义了将日志信息按照普通日志格式打印的逻辑：

```go
l := New("blockchain", "meta--")
l.SetHandler(StreamHandler(os.Stdout, LogfmtFormat()))
l.Trace("trace logger")
```

>输出：
>
> time=0001-01-01T00:00:00Z level=trace msg="trace logger" blockchain=meta--

### 将日志信息打印到文件里

```go
l := New("blockchain", "meta--")
file, _ := os.OpenFile("text.log", os.O_CREATE|os.O_APPEND|os.O_RDWR, 0664)
l.SetHandler(StreamHandler(file, TerminalFormat(false)))
l.Info("start service")
l.Error("start service")
```

结果：

![image-20221124212439854](https://gitee.com/Sagaya815/assets/raw/master/image-20221124212439854.png)

### 设置打印日志的级别

在下面的例子里，我们要求最多只打印`Warn`这一级别的日志，也就是说，`Trace Debug Info`这三个级别的日志不会被打印

```go
l := New("blockchain", "meta--")
l.SetHandler(LvlFilterHandler(LvlWarn, StreamHandler(os.Stdout, TerminalFormat(true))))
l.Info("info logger")
l.Warn("warn logger")
l.Error("error logger")
```

>输出：
>
>WARN*[01-01|00:00:00.000] warn logger                              blockchain=meta--
>
>ERROR[01-01|00:00:00.000] error logger                             blockchain=meta--

### 调试代码时输出日志

调试代码时输出的日志信息要想包含"file:line"这样的位置信息，打印日志的格式需要设置成控制台格式才能有效：

```go
PrintOrigins(true)
l := New("blockchain", "meta--")
l.SetHandler(StreamHandler(os.Stdout, TerminalFormat(true)))
l.Trace("trace logger")
```

`PrintOrigins(true)`函数里的`true`参数会把输出位置信息的开关打开。

>输出：
>
> TRACE[01-01|00:00:00.000|/log/log_test.go:24] trace logger                             blockchain=meta--