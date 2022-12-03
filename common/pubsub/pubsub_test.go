package pubsub

import (
	"github.com/232425wxy/meta--/log"
	"os"
	"testing"
)

func TestSetLoggerForServer(t *testing.T) {
	_s := NewServer()
	s := *_s
	l := log.New()
	l.SetHandler(log.StreamHandler(os.Stdout, log.TerminalFormat(true)))
	// s里的BaseService虽然是个值，不是指针，但是BaseService里的Logger是指针类型，在这里对其修改依然有效
	// 尽管s也是个值，不是一个指针
	s.SetLogger(l)

	s.Logger.Debug("hello, people")
}
