package log

import (
	"github.com/caser789/logger/internal/utils/env"
	"go.uber.org/zap"
)

var (
	defaultOptions = &option{
		LocalTime: false,
		Stdout:    false,
		Filename:  "customize",
		Ropt: rotateOptions{
			MaxSize:    100,
			MaxAge:     7,
			MaxBackups: 10,
			Compress:   false,
		},
		Lef: func(lvl LogLevel) bool {
			return lvl >= GetLevel()
		},
	}
)

type CustomizeOption func(*option)

// NewLogger will return a customized logger by CustomizeConfig.If config.LogFileName is empty,will write into customize.log.
func NewLogger(opts ...CustomizeOption) *zap.Logger {
	optCopy := defaultOptions
	for _, o := range opts {
		o(optCopy)
	}

	return newLogger(*optCopy)
}

func WithLogFileName(logPath, fileName string) CustomizeOption {
	if fileName == "" {
		fileName = "customize"
	}
	return func(o *option) {
		o.Filename = env.GetFilePath(logPath, fileName)
	}
}

func WithCompress(compress bool) CustomizeOption {
	return func(o *option) {
		o.Ropt.Compress = compress
	}
}

func WithPrintToStdout(printToStdout bool) CustomizeOption {
	return func(o *option) {
		o.Stdout = printToStdout
	}
}
