package trace

import "time"

// Tracer provide functions to create spans.
type Tracer interface {
	// NewSpan returns a root Span whose span context is initiated with the given requestID.
	NewSpan(name string, spanContext SpanContext) (Span, error)

	// NewSpanWithOptions returns a root Span with options
	NewSpanWithOptions(name string, spanContext SpanContext, options ...NewSpanOption) (Span, error)
}

// NewSpanOptions saves options when span is created
type NewSpanOptions struct {
	StartTime time.Time
}

// NewSpanOption is a function that sets some options when new a span
type NewSpanOption func(opts *NewSpanOptions)

// StartTime sets start time for span
func StartTime(time time.Time) NewSpanOption {
	return func(opts *NewSpanOptions) {
		opts.StartTime = time
	}
}
