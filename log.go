package log

import (
	"context"
	"go.uber.org/zap"
)

const TraceKey = "@jiao_trace_id"

// Log Interfaces

func Debug(ctx context.Context, msg string, fields ...zap.Field) {
	GetTraceLogFromCtx(ctx).Debug(msg, fields...)
}

func Debugf(ctx context.Context, args ...interface{}) {
	GetTraceLogFromCtx(ctx).Sugar().Debug(args)
}

func Info(ctx context.Context, msg string, fields ...zap.Field) {
	GetTraceLogFromCtx(ctx).Info(msg, fields...)
}

func Infof(ctx context.Context, args ...interface{}) {
	GetTraceLogFromCtx(ctx).Sugar().Info(args)
}

func Warn(ctx context.Context, msg string, fields ...zap.Field) {
	GetTraceLogFromCtx(ctx).Warn(msg, fields...)
}

func Warnf(ctx context.Context, args ...interface{}) {
	GetTraceLogFromCtx(ctx).Sugar().Warn(args)
}

func Error(ctx context.Context, msg string, fields ...zap.Field) {
	GetTraceLogFromCtx(ctx).Error(msg, fields...)
}

func Errorf(ctx context.Context, args ...interface{}) {
	GetTraceLogFromCtx(ctx).Sugar().Error(args)
}

func DPanic(ctx context.Context, msg string, fields ...zap.Field) {
	GetTraceLogFromCtx(ctx).DPanic(msg, fields...)
}

func DPanicf(ctx context.Context, args ...interface{}) {
	GetTraceLogFromCtx(ctx).Sugar().DPanic(args)
}

func Panic(ctx context.Context, msg string, fields ...zap.Field) {
	GetTraceLogFromCtx(ctx).Panic(msg, fields...)
}

func Panicf(ctx context.Context, args ...interface{}) {
	GetTraceLogFromCtx(ctx).Sugar().Panic(args)
}

func Fatal(ctx context.Context, msg string, fields ...zap.Field) {
	GetTraceLogFromCtx(ctx).Fatal(msg, fields...)
}

func Fatalf(ctx context.Context, args ...interface{}) {
	GetTraceLogFromCtx(ctx).Sugar().Fatal(args)
}

// System log interface

// SysDebug - System log in DebugLvl level.
func SysDebug(ctx context.Context, msg string, fields ...zap.Field) {
	getSysLogger(ctx).Debug(msg, fields...)
}

// SysDebugf - System sugar log in DebugLvl level.
func SysDebugf(ctx context.Context, args ...interface{}) {
	getSysLogger(ctx).Sugar().Debug(args)
}

// SysInfo - System log in InfoLvl level.
func SysInfo(ctx context.Context, msg string, fields ...zap.Field) {
	getSysLogger(ctx).Info(msg, fields...)
}

// SysInfof - System sugar log in InfoLvl level.
func SysInfof(ctx context.Context, args ...interface{}) {
	getSysLogger(ctx).Sugar().Info(args)
}

// SysWarn - System log in WarnLvl level.
func SysWarn(ctx context.Context, msg string, fields ...zap.Field) {
	getSysLogger(ctx).Warn(msg, fields...)
}

// SysWarnf - System sugar log in WarnLvl level.
func SysWarnf(ctx context.Context, args ...interface{}) {
	getSysLogger(ctx).Sugar().Warn(args)
}

// SysError - System log in ErrorLvl level.
func SysError(ctx context.Context, msg string, fields ...zap.Field) {
	getSysLogger(ctx).Error(msg, fields...)
}

// SysErrorf - System sugar log in ErrorLvl level.
func SysErrorf(ctx context.Context, args ...interface{}) {
	getSysLogger(ctx).Sugar().Error(args)
}

// SysDPanic - System log in DPanicLvl level.
func SysDPanic(ctx context.Context, msg string, fields ...zap.Field) {
	getSysLogger(ctx).DPanic(msg, fields...)
}

// SysDPanicf - System sugar log in DPanicLvl level.
func SysDPanicf(ctx context.Context, args ...interface{}) {
	getSysLogger(ctx).Sugar().DPanic(args)
}

// SysPanic - System log in PanicLvl level.
func SysPanic(ctx context.Context, msg string, fields ...zap.Field) {
	getSysLogger(ctx).Panic(msg, fields...)
}

// SysPanicf - System sugar log in PanicLvl level.
func SysPanicf(ctx context.Context, args ...interface{}) {
	getSysLogger(ctx).Sugar().Panic(args)
}

// SysFatal - System log in FatalLvl level.
func SysFatal(ctx context.Context, msg string, fields ...zap.Field) {
	getSysLogger(ctx).Fatal(msg, fields...)
}

// SysFatalf - System sugar log in FatalLvl level.
func SysFatalf(ctx context.Context, args ...interface{}) {
	getSysLogger(ctx).Sugar().Fatal(args)
}

func getSysLogger(ctx context.Context) *zap.Logger {
	traceID := GetTraceIDFromCtx(ctx)
	return GetSysLogger().With(zap.String(TraceKey, traceID))
}

func getTracingLogger(ctx context.Context) *zap.Logger {
	traceID := GetTraceIDFromCtx(ctx)
	return GetTracingLogger().With(zap.String(TraceKey, traceID))
}

// Tracing Log Interface

// Tracing write log,info level only
func Tracing(ctx context.Context, msg string, fields ...zap.Field) {
	getTracingLogger(ctx).Info(msg, fields...)
}

// Tracingf tracing sugar log,info level only
func Tracingf(ctx context.Context, args ...interface{}) {
	getTracingLogger(ctx).Sugar().Info(args)
}

func TracingDebug(ctx context.Context, msg string, fields ...zap.Field) {
	getTracingLogger(ctx).Debug(msg, fields...)
}

func TracingDebugf(ctx context.Context, args ...interface{}) {
	getTracingLogger(ctx).Sugar().Debug(args)
}
