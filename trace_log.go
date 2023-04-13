package log

import (
	"context"

	"github.com/caser789/logger/internal/extension"
	"github.com/caser789/logger/internal/trace"
	ctxzap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"go.uber.org/zap"
)

func WithNewTraceLog(operationName string, ctx context.Context) (context.Context, trace.Span) {
	spanCtx := GetSpanContext(ctx)
	if spanCtx == nil {
		spanCtx = trace.NewSpanContextGenerator("").NewSpanContext()
	}
	span, _ := trace.GlobalTracer().NewSpan(operationName, spanCtx)
	ctx = WithSpanContext(ctx, spanCtx)
	newLogger := GetLogger().With(zap.String(extension.TraceKey, spanCtx.String()))
	ctx = ctxzap.ToContext(ctx, newLogger)
	return ctx, span
}

func GetTraceIDFromCtx(ctx context.Context) string {
	if spanCtx := GetSpanContext(ctx); spanCtx != nil {
		return spanCtx.String()
	}
	return ""
}

func GetTraceLogFromCtx(ctx context.Context) *zap.Logger {
	l := ctxzap.Extract(ctx)
	if l.Core().Enabled(zap.FatalLevel) {
		return l
	}
	return GetLogger()
}

func WithLogger(ctx context.Context, logger *zap.Logger) context.Context {
	return ctxzap.ToContext(ctx, logger)
}

// GetSpanContext will return the current SpanContext from context
func GetSpanContext(ctx context.Context) trace.SpanContext {
	if ctx == nil {
		return nil
	}

	spanContext, ok := ctx.Value(contextKeyForSpanContext).(trace.SpanContext)
	if ok {
		return spanContext
	}

	return nil
}

type spanContextCtxKey string

const (
	// contextKeyForSpanContext is the key in the context for SpanContext
	contextKeyForSpanContext = spanContextCtxKey("sc")
)

// WithSpanContext sets the SpanContext in context
func WithSpanContext(ctx context.Context, spanContext trace.SpanContext) context.Context {
	return context.WithValue(ctx, contextKeyForSpanContext, spanContext)
}
