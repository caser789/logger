package trace

// A NoopTracer is a trivial, minimum overhead implementation of Tracer
// for which all operations are no-ops. This is used by other projects to mock tracing dependency
// when doing unit test or toggle off tracing functionality urgently in runtime if anything goes wrong.
type NoopTracer struct {
}

// NoopSpan structure definition
type NoopSpan struct {
	ctx    SpanContext
	tracer Tracer
}

// Context ... returns SpanContext of current NoopSpan.
func (ns *NoopSpan) Context() SpanContext {
	return ns.ctx
}

// NewChildSpan ... creates and returns a child Span of current NoopSpan
func (ns *NoopSpan) NewChildSpan(name string) (Span, error) {
	childSpanContext := ns.ctx.NewChildSpanContext()
	newSpan, err := ns.tracer.NewSpan(name, childSpanContext)
	if err != nil {
		return nil, err
	}
	return newSpan, nil
}

// SetTag ... NoopSpan implementation of this function does nothing
func (ns *NoopSpan) SetTag(key string, value interface{}) Span {
	return ns
}

// SetTags ... NoopSpan implementation of this function does nothing
func (ns *NoopSpan) SetTags(tags ...interface{}) Span {
	return ns
}

// SetDebugTags ... NoopSpan implementation of this function does nothing
func (ns *NoopSpan) SetDebugTags(tags ...interface{}) Span {
	return ns
}

// LogFields ... NoopSpan implementation of this function does nothing
func (ns *NoopSpan) LogFields(fields ...interface{}) Span {
	return ns
}

// LogDebugFields ... NoopSpan implementation of this function does nothing
func (ns *NoopSpan) LogDebugFields(fields ...interface{}) Span {
	return ns
}

// Finish ... NoopSpan implementation of this function does nothing
func (ns *NoopSpan) Finish() {
}

// FinishWithOptions ... NoopSpan implementation of this function does nothing
func (ns *NoopSpan) FinishWithOptions(option ...FinishSpanOption) {
}

// NewSpan ... returns the new instance of NoopSpan
func (nt *NoopTracer) NewSpan(name string, spanContext SpanContext) (Span, error) {
	return nt.NewSpanWithOptions(name, spanContext)
}

// NewSpanWithOptions ... returns the new instance of NoopSpan with options
func (nt *NoopTracer) NewSpanWithOptions(name string, spanContext SpanContext, option ...NewSpanOption) (Span, error) {
	return &NoopSpan{
		ctx:    spanContext,
		tracer: nt,
	}, nil
}

// NewNoopTracer ... returns the new instance of NoopTracer
func NewNoopTracer() Tracer {
	return &NoopTracer{}
}
