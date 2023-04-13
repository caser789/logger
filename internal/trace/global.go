package trace

import (
	"go.uber.org/atomic"
)

var globalTracer atomic.Value

type registeredTracer struct {
	tracer Tracer
}

func init() {
	globalTracer.Store(registeredTracer{tracer: NewNoopTracer()})
}

// SetGlobalTracer sets the [singleton] Tracer returned by GlobalTracer().
func SetGlobalTracer(tracer Tracer) {
	globalTracer.Store(registeredTracer{tracer: tracer})
}

// GlobalTracer returns the global singleton `Tracer` implementation.
// Before `SetGlobalTracer()` is called, the `GlobalTracer()` is a noop implementation that drops all data handed to it.
func GlobalTracer() Tracer {
	if regTracer, ok := globalTracer.Load().(registeredTracer); ok {
		return regTracer.tracer
	}
	return NewNoopTracer()
}
