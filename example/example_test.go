package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	log "github.com/caser789/logger"
	"github.com/caser789/logger/internal/extension"
	"github.com/hashicorp/go-multierror"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestNewLogger(t *testing.T) {
	newLogger := log.NewLogger(log.WithLogFileName("", "request"),
		log.WithCompress(false))
	defer newLogger.Sync()
	newLogger.Info("test")
	testLogger := log.NewLogger(log.WithLogFileName("test", "file"))
	testLogger.Error("tests")
	defer testLogger.Sync()
	cLogger := log.NewLogger(log.WithLogFileName("", "request"))
	defer cLogger.Sync()
	cLogger.Error("sfasf")
}

func TestMultiError(t *testing.T) {
	var err *multierror.Error
	e1 := errors.New("error1---test")
	e2 := errors.New("error2")
	e3 := errors.New("error3")
	fmt.Println(e1.Error())
	err = multierror.Append(err, e1)
	fmt.Println(err.Error())
	err = multierror.Append(err, e2)
	fmt.Println(err.Error())
	err = multierror.Append(err, e3)
	fmt.Println(err.Error())
}

func TestInitTracingLog(t *testing.T) {
	ok := log.SetLogFileName(log.InfoLvl, "traffic_recording")
	assert.Equal(t, false, ok)
	ok = log.SetLogFileName(log.InfoLvl, "tracing")
	assert.Equal(t, true, ok)
	log.InitLogger(&log.Config{
		Level:              log.DebugLvl,
		TracingLogFileName: "tracing",
		SplitLevel:         log.SplitDebug,
	})
	ctx := context.Background()
	logger := log.GetLogger()
	logger.Info("info")
	log.Tracing(ctx, "test", zap.String("name", "xiaofeng.liu"))
	tracingLogger := log.GetTracingLogger()
	defer func() {
		logger.Sync()
		tracingLogger.Sync()
	}()
}

func TestGetTracingLog(t *testing.T) {
	log.GetTracingLogger()
	ctx := context.Background()
	log.Tracing(ctx, "ahha")
}

func TestSetLogName(t *testing.T) {
	ok := log.SetLogFileName(log.WarnLvl, "server")
	assert.Equal(t, false, ok)
	ok = log.SetLogFileName(log.DebugLvl, "info")
	assert.Equal(t, false, ok)
	ok = log.SetLogFileName(log.DebugLvl, "sys")
	assert.Equal(t, false, ok)
	ok = log.SetLogFileName(log.DebugLvl, "sys_error")
	assert.Equal(t, false, ok)
	ok = log.SetLogFileName(log.InfoLvl, "myInfo")
	assert.Equal(t, true, ok)
	ok = log.SetLogFileName(log.WarnLvl, "myInfo")
	assert.Equal(t, false, ok)
	ok = log.SetLogFileName(log.WarnLvl, "myWarn")
	assert.Equal(t, true, ok)
	log.InitLogger(&log.Config{
		LogFileName: "myWarn",
		Level:       log.DebugLvl,
		SplitLevel:  log.SplitDebug,
	})
	logger := log.GetLogger()
	defer func() {
		logger.Sync()
	}()
	logger.Info("info")
	logger.Warn("warn")
	logger.Debug("debug")
}

func TestInitLog(t *testing.T) {
	log.InitLogger(&log.Config{
		Level: log.InfoLvl,
	})
	logger := log.GetLogger()
	defer func() {
		logger.Sync()
	}()
	logger.Info("haha")
}

func TestPrintToStd(t *testing.T) {
	log.InitLogger(&log.Config{
		Level:      log.DebugLvl,
		SplitLevel: log.SplitDebug,
		PrintToStd: log.PrintToStd_ALL,
	})

	logger := log.GetLogger()
	sysLogger := log.GetSysLogger()
	defer func() {
		sysLogger.Sync()
		logger.Sync()
	}()
	logger.Info("test user log")
	logger.Debug("debug")
	log.SetLevel(log.DebugLvl, time.Second*1)
	logger.Error("test default error")
	sysLogger.Error("test sys log")
	logger.Debug("debug")
}

func TestWithoutSplitLog(t *testing.T) {
	log.InitLogger(&log.Config{
		LogFileName: "test",
		Level:       log.InfoLvl,
	})
	sysLogger := log.GetSysLogger()
	logger := log.GetLogger()
	defer func() {
		sysLogger.Sync()
		logger.Sync()
	}()
	sysLogger.Info("sys_info")
	sysLogger.Error("sys_error")
	logger.Debug("user_debug")
	logger.Info("user_info")
	logger.Error("user_error")
	logger.Panic("panic")
}

func TestSplitLog(t *testing.T) {
	log.InitLogger(&log.Config{
		Level:      log.DebugLvl,
		SplitLevel: log.SplitDebug,
		PrintToStd: log.PrintToStd_ALL,
	})
	sysLogger := log.GetSysLogger()
	logger := log.GetLogger()
	defer func() {
		sysLogger.Sync()
		logger.Sync()
	}()
	sysLogger.Error("sys_error")
	logger.Debug("debug")
	logger.Info("info")
	logger.Warn("warn")
	logger.Error("error")
	log.SetLevel(log.DebugLvl, time.Second*3)
	logger.Debug("run timing debug")
	//	logger.DPanic("DPanic")
	//logger.Panic("Panic")
	//	logger.Fatal("Fatal")
}

func TestLog(t *testing.T) {
	logger := log.GetSysLogger()
	//logger.Info("hello", zap.String("test", "test"))
	//
	//// sys log example
	l := logger.With(zap.String(extension.TraceKey, "123454"))
	l.Info("hello", zap.Int("a", 123))

	logger.Sync()

}

func TestLazyInit(t *testing.T) {
	for i := 0; i < 10; i++ {
		go func() {
			logger := log.GetLogger()
			logger.Info("test", zap.String("info", "init"))
			defer func() {
				logger.Sync()
			}()
		}()
	}
	time.Sleep(5 * time.Second)
}

func TestDebug(t *testing.T) {
	splitLogConfig := &log.Config{
		Level:      log.InfoLvl,
		SplitLevel: log.SplitDebug,
		PrintToStd: log.PrintToStd_NONE,
	}
	log.InitLogger(splitLogConfig)
	logger := log.GetLogger()
	defer logger.Sync()
	log.SetLevel(log.DebugLvl, time.Second*30)
	logger.Debug("debug")
	logger.Info("info")
}

func TestResetLog(t *testing.T) {
	logger := log.GetLogger()

	logger.Debug("hi")
	logger.Info("hi")
	logger.Warn("hi")
	logger.Error("hi")

	log.SetLevel(zap.DebugLevel, 5*time.Second)
	log.Info(context.TODO(), "set debug level")
	logger.Debug("set debug level")
	logger.Info("set debug level")
	logger.Warn("set debug level")
	logger.Error("set debug level")

	log.SetLevel(zap.WarnLevel, 5*time.Second)

	//log.SetLevel(zap.InfoLevel, 5*time.Second)
	log.Info(context.TODO(), "live env set to warn level")
	logger.Debug("env set to warn level")
	logger.Info("env set to warn level")
	logger.Warn("env set to warn level")
	logger.Error("env set to warn level")

	for i := 1; i < 3; i++ {
		go func() {
			log.SetLevel(zap.DebugLevel, 2*time.Second)
		}()
	}

	log.Info(context.TODO(), "multiple reset")
	logger.Debug("multiple reset")
	logger.Info("multiple reset")
	logger.Warn("multiple reset")
	logger.Error("multiple reset")

	time.Sleep(10 * time.Second)
	log.Info(context.TODO(), "time reach")
	logger.Debug("time reach")
	logger.Info("time reach")
	logger.Warn("time reach")
	logger.Error("time reach")

	logger.Sync()
}

func TestSplitLogDefault(t *testing.T) {
	os.Setenv("SPLIT_LOG", "true")
	log.Debug(context.TODO(), "test debug")
	log.Info(context.TODO(), "test info")
	log.Warn(context.TODO(), "test warn")
	log.Error(context.TODO(), "test error")
	log.Sync()
}

func TestLogDefault(t *testing.T) {
	log.Debug(context.TODO(), "test debug")
	log.Info(context.TODO(), "test info")
	log.Warn(context.TODO(), "test warn")
	log.Error(context.TODO(), "test error")
	log.Sync()
}
