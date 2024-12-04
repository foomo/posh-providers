package migrate

import (
	"github.com/foomo/posh/pkg/log"
)

type logger struct {
	l log.Logger
}

func (l *logger) Printf(format string, v ...interface{}) {
	l.l.Infof(format, v...)
}

func (l *logger) Verbose() bool {
	return l.l.IsLevel(log.LevelDebug)
}
