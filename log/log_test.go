package log

import (
	"os"
	"testing"
)

func TestTerminalfmt(t *testing.T) {
	// 不同级别使用不同颜色
	l := New("blockchain", "meta--")
	l.SetHandler(StreamHandler(os.Stdout, TerminalFormat(true)))
	l.Trace("trace logger")
	l.Debug("debug logger")
	l.Info("info logger")
	l.Warn("warn logger")
	l.Error("error logger")
	//l.Crit("crit logger")

	// 不使用颜色
	l.SetHandler(StreamHandler(os.Stdout, TerminalFormat(false)))
	l.Trace("trace logger")
	l.Debug("debug logger")
	l.Info("info logger")
	l.Warn("warn logger")
	l.Error("error logger")
	//l.Crit("crit logger")

	// 打印输出日志信息的代码位置
	PrintOrigins(true)
	l.Trace("trace logger")
	l.Debug("debug logger")
	l.Info("info logger")
	l.Warn("warn logger")
	l.Error("error logger")
	//l.Crit("crit logger")

	// Output:
	// TRACE[01-01|00:00:00.000] trace logger                             blockchain=meta--
	// DEBUG[01-01|00:00:00.000] debug logger                             blockchain=meta--
	// INFO*[01-01|00:00:00.000] info logger                              blockchain=meta--
	// WARN*[01-01|00:00:00.000] warn logger                              blockchain=meta--
	// ERROR[01-01|00:00:00.000] error logger                             blockchain=meta--
	// TRACE[01-01|00:00:00.000] trace logger                             blockchain=meta--
	// DEBUG[01-01|00:00:00.000] debug logger                             blockchain=meta--
	// INFO*[01-01|00:00:00.000] info logger                              blockchain=meta--
	// WARN*[01-01|00:00:00.000] warn logger                              blockchain=meta--
	// ERROR[01-01|00:00:00.000] error logger                             blockchain=meta--
	// TRACE[01-01|00:00:00.000|/log/log_test.go:30] trace logger                             blockchain=meta--
	// DEBUG[01-01|00:00:00.000|/log/log_test.go:31] debug logger                             blockchain=meta--
	// INFO*[01-01|00:00:00.000|/log/log_test.go:32] info logger                              blockchain=meta--
	// WARN*[01-01|00:00:00.000|/log/log_test.go:33] warn logger                              blockchain=meta--
	// ERROR[01-01|00:00:00.000|/log/log_test.go:34] error logger                             blockchain=meta--
}

func TestJSONfmt(t *testing.T) {
	l := New("blockchain", "meta--")
	l.SetHandler(StreamHandler(os.Stdout, JSONFormat()))
	l.Trace("trace logger")
	l.Debug("debug logger")
	l.Info("info logger")
	l.Warn("warn logger")
	l.Error("error logger")
	l.Crit("crit logger")

	// Output:
	// {"blockchain":"meta--","level":"trace","msg":"trace logger","time":"0001-01-01T00:00:00Z"}
	// {"blockchain":"meta--","level":"debug","msg":"debug logger","time":"0001-01-01T00:00:00Z"}
	// {"blockchain":"meta--","level":"info*","msg":"info logger","time":"0001-01-01T00:00:00Z"}
	// {"blockchain":"meta--","level":"warn*","msg":"warn logger","time":"0001-01-01T00:00:00Z"}
	// {"blockchain":"meta--","level":"error","msg":"error logger","time":"0001-01-01T00:00:00Z"}
	// {"blockchain":"meta--","level":"critical","msg":"crit logger","time":"0001-01-01T00:00:00Z"}
}

func TestLogfmt(t *testing.T) {
	l := New("blockchain", "meta--")
	l.SetHandler(StreamHandler(os.Stdout, LogfmtFormat()))
	l.Trace("trace logger")
	l.Debug("debug logger")
	l.Info("info logger")
	l.Warn("warn logger")
	l.Error("error logger")
	l.Crit("crit logger")

	// Output:
	// time=0001-01-01T00:00:00Z level=trace msg="trace logger" blockchain=meta--
	// time=0001-01-01T00:00:00Z level=debug msg="debug logger" blockchain=meta--
	// time=0001-01-01T00:00:00Z level=info* msg="info logger"  blockchain=meta--
	// time=0001-01-01T00:00:00Z level=warn* msg="warn logger"  blockchain=meta--
	// time=0001-01-01T00:00:00Z level=error msg="error logger" blockchain=meta--
	// time=0001-01-01T00:00:00Z level=critical msg="crit logger"  blockchain=meta--
}

func TestFilterLvl(t *testing.T) {
	l := New("blockchain", "meta--")
	l.SetHandler(LvlFilterHandler(LvlWarn, StreamHandler(os.Stdout, JSONFormat())))
	l.Trace("trace logger")
	l.Debug("debug logger")
	l.Info("info logger")
	l.Warn("warn logger")
	l.Error("error logger")
	l.Crit("crit logger")

	// Output:
	// {"blockchain":"meta--","level":"warn","msg":"warn logger","time":"0001-01-01T00:00:00Z"}
	// {"blockchain":"meta--","level":"eror","msg":"error logger","time":"0001-01-01T00:00:00Z"}
	// {"blockchain":"meta--","level":"crit","msg":"crit logger","time":"0001-01-01T00:00:00Z"}
}

func TestRedirectToFile(t *testing.T) {
	file, _ := os.OpenFile("text.log", os.O_CREATE|os.O_APPEND|os.O_RDWR, 0644)
	l := New("blockchain", "meta--")
	l.SetHandler(StreamHandler(file, TerminalFormat(false)))
	l.Trace("trace logger")
	l.Debug("debug logger")
	l.Info("info logger")
	l.Warn("warn logger")
	l.Error("error logger")
	l.Crit("crit logger")
}

func TestInherit(t *testing.T) {
	l := New("conn", "P2P/Connection:192.168.137.72:22")
	l.SetHandler(StreamHandler(os.Stdout, TerminalFormat(true)))
	l.Debug("hi")

	child := l.New("channel id", 0x01)
	child.Debug("hello")
}
