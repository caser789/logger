package log

import (
	"io"
	"os"
	"sync"
	"time"

	"github.com/caser789/logger/internal/extension"
	"github.com/caser789/logger/internal/lumberjack"
	"github.com/caser789/logger/internal/utils/env"
	grpczap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	"github.com/hashicorp/go-multierror"
	"go.uber.org/atomic"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type LogLevel = zapcore.Level
type SplitLevel string
type PrintToStd uint8

const (
	DebugLvl                      = zapcore.DebugLevel
	InfoLvl                       = zapcore.InfoLevel
	WarnLvl                       = zapcore.WarnLevel
	ErrorLvl                      = zapcore.ErrorLevel
	PanicLvl                      = zapcore.PanicLevel
	DPanicLvl                     = zapcore.DPanicLevel
	FatalLvl                      = zapcore.FatalLevel
	SplitError         SplitLevel = "error"
	SplitWarn          SplitLevel = "warn"
	SplitInfo          SplitLevel = "info"
	SplitDebug         SplitLevel = "debug"
	SplitNone          SplitLevel = "none"
	PrintToStd_NONE    PrintToStd = 0
	PrintToStd_USERLOG PrintToStd = 1
	PrintToStd_SYSLOG  PrintToStd = 2
	PrintToStd_TRACING PrintToStd = 4
	PrintToStd_ALL     PrintToStd = 7
)

const (
	customTimeLayout       = "2006-01-02 15:04:05.999999-07:00"
	maxResetLvlDur         = 48 * time.Hour
	maxResetVer            = 10000
	SysLogFileName         = "sys"
	SysErrorLogFileName    = "sys_error"
	DefaultLogFileName     = "server"
	DefaultTracingFileName = "traffic_recording"
)

var (
	logger        *zap.Logger
	sysLogger     *zap.Logger
	tracingLogger *zap.Logger

	loggerInitOnce        sync.Once
	sysLoggerInitOnce     sync.Once
	tracingLoggerInitOnce sync.Once

	logLevel        atomic.Int32
	initialLogLevel LogLevel

	resetVer atomic.Int64

	levelMap = map[SplitLevel]LogLevel{SplitDebug: DebugLvl, SplitInfo: InfoLvl, SplitWarn: WarnLvl, SplitError: ErrorLvl}
	nameMap  = map[LogLevel]string{DebugLvl: "debug", InfoLvl: "info", WarnLvl: "warn", ErrorLvl: "error"}
)

type Config struct {
	// Level - Initial default log level.
	// By default, DebugLvl is set for non-live environment, while InfoLvl is set for live environment.
	Level LogLevel
	//Deprecated -Print all logs into stdout.Deprecated,can use PrintToStd to make sure which kind log you want to see in stdout.
	PrintToStdout bool
	// PrintToStdout - Which kind log you want print into stdout,default none.Only effect in the non-live environment
	PrintToStd PrintToStd
	Compress   bool
	// Path - Customized log file path.Only effect in K8S. Log files will be created under ./log dir if not specified.
	Path string
	// LogFileName - Customized log file name. It will be server.log if not specified.
	LogFileName string
	// SplitLevel -The minimum level of logs to be split. Logs greater than this level will write into different file.
	//Logs less than this level will write into server.log,and all log will write into server.log if not set.
	SplitLevel SplitLevel
	//TracingLogFileName -Customized tracing log file.It will be traffic_recording.log if not specified
	TracingLogFileName string
}

// InitLogger - Initialize the logger and system logger.
// This function should only run once.
func InitLogger(config *Config) {
	initLogLevel(config)

	loggerInitOnce.Do(func() {
		// init default logger
		initDefaultLogger(config)
	})

	sysLoggerInitOnce.Do(func() {
		// init system logger
		initSystemLogger(config)
	})

	tracingLoggerInitOnce.Do(func() {
		// init tracing logger
		initTracingLogger(config)
	})
}

func initLogLevel(config *Config) {
	lvl := defaultLevel()
	if config.Level > lvl {
		lvl = config.Level
	}
	initialLogLevel = lvl
	SetLevel(lvl, 0)
}

// GetLogger - Return the logger. The output log will be in
// the ./log/error.log and./log/server.log file.
func GetLogger() *zap.Logger {
	loggerInitOnce.Do(func() {
		config := getDefaultConfig()
		initLogLevel(config)
		initDefaultLogger(config)
	})
	return logger
}

// GetSysLogger - Return system logger.
// The output log will be in the ./log/sys_error.log and./log/sys.log file.
// This logger should only be used for system framework. Please use GetLogger() for your business log.
func GetSysLogger() *zap.Logger {
	sysLoggerInitOnce.Do(
		func() {
			config := &Config{
				Level:      InfoLvl,
				PrintToStd: PrintToStd_NONE,
			}
			initSystemLogger(config)
		},
	)

	return sysLogger
}

func GetTracingLogger() *zap.Logger {
	tracingLoggerInitOnce.Do(
		func() {
			config := &Config{
				Level:              InfoLvl,
				TracingLogFileName: DefaultTracingFileName,
			}
			initTracingLogger(config)
		})
	return tracingLogger
}

func Sync() error {
	var res *multierror.Error
	if err := GetLogger().Sync(); err != nil {
		res = multierror.Append(res, err)
	}
	if err := GetSysLogger().Sync(); err != nil {
		res = multierror.Append(res, err)
	}
	if err := GetTracingLogger().Sync(); err != nil {
		res = multierror.Append(res, err)
	}
	return res
}

func initTracingLogger(config *Config) {
	printToStd := config.PrintToStd
	if config.TracingLogFileName == "" {
		config.TracingLogFileName = DefaultTracingFileName
	}
	for _, s := range nameMap {
		if config.TracingLogFileName == s {
			config.TracingLogFileName = DefaultTracingFileName
		}
	}
	var opts []option
	if (printToStd == PrintToStd_TRACING || printToStd == PrintToStd_ALL || config.PrintToStdout) && !env.IsLive() {
		opts = append(opts, option{
			Stdout: true,
			Lef: func(level zapcore.Level) bool {
				return level >= GetLevel()
			},
		})
	} else {
		opts = append(opts, getOption(config, config.TracingLogFileName, func(level zapcore.Level) bool {
			return level >= GetLevel()
		}))
	}
	tracingLogger = newLogger(opts...)
}

func initSystemLogger(config *Config) {
	var opts []option
	printToStd := config.PrintToStd
	if (printToStd == PrintToStd_SYSLOG || printToStd == PrintToStd_ALL || config.PrintToStdout) && !env.IsLive() {
		opts = append(opts, option{
			Stdout: true,
			Lef: func(lvl LogLevel) bool {
				return lvl >= GetLevel()
			},
		})
	} else {
		opts = append(opts, getOption(config, SysErrorLogFileName, func(lvl LogLevel) bool {
			return lvl >= ErrorLvl
		}))
		opts = append(opts, getOption(config, SysLogFileName, func(lvl LogLevel) bool {
			return lvl >= GetLevel()
		}))
	}

	sysLogger = newLogger(opts...)
	grpczap.ReplaceGrpcLoggerV2(sysLogger)
}

func initDefaultLogger(config *Config) {
	if config.LogFileName == "" {
		config.LogFileName = DefaultLogFileName
	}
	for _, name := range nameMap {
		if config.LogFileName == name {
			config.LogFileName = DefaultLogFileName
		}
	}

	var opts []option
	printToStd := config.PrintToStd
	if (printToStd == PrintToStd_USERLOG || printToStd == PrintToStd_ALL || config.PrintToStdout) && !env.IsLive() {
		printToStdOut(config)
		return
	}

	splitLevel, ok := checkLevel(config.SplitLevel)

	if ok {
		opts = getSplitOpt(config, splitLevel)
	} else {
		opts = getDefaultOpt(config)
	}

	logger = newLogger(opts...)
	zap.ReplaceGlobals(logger)
}

func printToStdOut(config *Config) {
	opt := option{
		Stdout: true,
		Lef: func(lvl LogLevel) bool {
			return lvl >= GetLevel()
		},
	}
	logger = newLogger(opt)
}

func checkLevel(splitLevel SplitLevel) (LogLevel, bool) {
	if splitLevel == "" || splitLevel == SplitNone {
		return 0, false
	}
	if level, ok := levelMap[splitLevel]; ok {
		return level, true
	}
	return 0, false
}

func getSplitOpt(config *Config, splitLevel LogLevel) []option {
	var opts []option
	//server.log
	if splitLevel != DebugLvl {
		opts = append(opts, getOption(config, config.LogFileName, func(lvl LogLevel) bool {
			return lvl >= GetLevel() && lvl < splitLevel
		}))
	}
	//split log to different file
	for level, s := range nameMap {
		if level >= splitLevel && level != ErrorLvl {
			l := level
			opts = append(opts, getOption(config, s, func(lvl LogLevel) bool {
				return lvl >= GetLevel() && lvl == l
			}))
		}
	}
	//error log contains all logs that loglevel > error
	opts = append(opts, getOption(config, nameMap[ErrorLvl], func(lvl LogLevel) bool {
		return lvl >= GetLevel() && lvl >= ErrorLvl
	}))
	return opts
}

func getOption(config *Config, fileName string, enablerFunc zap.LevelEnablerFunc) option {
	return option{
		Filename: env.GetFilePath(config.Path, fileName),
		Ropt: rotateOptions{
			MaxSize:    100,
			MaxAge:     7,
			MaxBackups: 10,
			Compress:   config.Compress,
		},
		Lef: enablerFunc,
	}
}

func getDefaultOpt(config *Config) []option {
	var opts []option
	opts = append(opts, getOption(config, config.LogFileName, func(lvl LogLevel) bool {
		return lvl >= GetLevel()
	}))
	return opts
}

type rotateOptions struct {
	MaxSize    int
	MaxAge     int
	MaxBackups int
	Compress   bool
}

type option struct {
	LocalTime bool
	Stdout    bool
	Filename  string
	Ropt      rotateOptions
	Lef       zap.LevelEnablerFunc
}

func newLogger(opts ...option) *zap.Logger {
	var cores []zapcore.Core
	encCfg := extension.NewProductionEncoderConfig()
	encCfg.EncodeTime = zapcore.TimeEncoderOfLayout(customTimeLayout)
	encCfg.EncodeDuration = zapcore.MillisDurationEncoder
	encCfg.ConsoleSeparator = "|"
	encoder := extension.NewConsoleEncoder(encCfg)

	for _, opt := range opts {
		core := newCore(encoder, opt)
		cores = append(cores, core)
	}

	logger := zap.New(zapcore.NewTee(cores...), zap.AddCaller(), zap.AddStacktrace(zap.PanicLevel))
	logger = logger.With(zap.String(TraceKey, "-"))

	return logger
}

func newCore(encoder zapcore.Encoder, opt option) zapcore.Core {
	lv := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return opt.Lef(lvl)
	})

	var syncer io.Writer
	if opt.Stdout {
		syncer = os.Stdout
	} else {
		syncer = &lumberjack.Logger{
			LocalTime:  opt.LocalTime,
			Filename:   opt.Filename,
			MaxSize:    opt.Ropt.MaxSize,
			MaxBackups: opt.Ropt.MaxBackups,
			MaxAge:     opt.Ropt.MaxAge,
			Compress:   opt.Ropt.Compress,
		}
	}
	w := zapcore.AddSync(syncer)
	core := zapcore.NewCore(
		encoder,
		zapcore.AddSync(w),
		lv,
	)
	return core
}

func defaultLevel() zapcore.Level {
	if env.IsLive() {
		return zap.InfoLevel
	}
	return zap.DebugLevel
}

// SetLevel - Dynamically set the log level. Time duration only works when
// setting log level to debug on live environment, that is, when the log level
// is dynamically set to debug level in live env, it will be reset
// to the initial log level configuration (default InfoLvl if not specified initially)
// after the time duration.
func SetLevel(level zapcore.Level, duration time.Duration) {
	if env.IsLive() && level < zap.InfoLevel {
		if duration > maxResetLvlDur {
			duration = maxResetLvlDur
		}

		ver := resetVer.Add(1)
		if ver > maxResetVer {
			ver = ver % maxResetVer
			resetVer.Store(ver)
		}

		time.AfterFunc(duration, func() {
			resetLevel(ver)
		})
	}

	logLevel.Store(int32(level))
}

func resetLevel(ver int64) {
	if GetLevel() < InfoLvl && resetVer.Load() == ver {
		SetLevel(initialLogLevel, 0)
		return
	}
}

func GetLevel() zapcore.Level {
	return zapcore.Level(int8(logLevel.Load()))
}

// SetLogFileName Set the filename of logs that split by level.
// Return true if newName is valid and set success,return false if newName is duplicate with other log file and set fail.
// Should be called before logger initialization.
// E.g. SetLogFileName(log.DebugLvl,"my_debug") then all the debug log will write into my_debug.log.
func SetLogFileName(level LogLevel, newName string) bool {
	if !checkLogFileNameValid(level, newName) {
		return false
	}
	nameMap[level] = newName
	return true
}

func checkLogFileNameValid(level LogLevel, newName string) bool {
	if newName == "" || newName == SysLogFileName || newName == SysErrorLogFileName || newName == DefaultLogFileName || newName == DefaultTracingFileName {
		return false
	}
	for l, s := range nameMap {
		if l != level && newName == s {
			return false
		}
	}
	return true
}
