package trace

import (
	"encoding/hex"
	"errors"
	"math/rand"
	"strings"
	"sync"
	"time"
)

const (
	traceIDSize = 16
	spanIDSize  = 8
	// totalIDSize total byte length of traceID, spanID and parentID
	totalIDSize = traceIDSize + spanIDSize*2

	// indicates the total number of bits being used for components of the flag
	flagBitsTotal         = 8
	flagBitsTypeMarker    = 3
	flagBitsNonTypeMarker = flagBitsTotal - flagBitsTypeMarker

	traceFlagOldDebug      = 1
	traceFlagSampled       = 1 << 1
	traceFlagCritical      = 1 << 2
	typeMarkerForOldFormat = 0 << 5

	typeMarkerForDebug      = 2 << 5
	typeMarkerForStressTest = 3 << 5
	typeMarkerForShadow     = 4 << 5

	typeMarkerMask uint8 = 7 << 5

	// Human readable request types as indicated by their type marker

	// ReqTypeOldFormat indicates requests using the old span context type marker 000xxxxx
	ReqTypeOldFormat = "old_format"
	// ReqTypeNormal indicates requests from regular traffic using type marker 001xxxxx
	ReqTypeNormal = "normal"
	// ReqTypeDebug indicates requests with debug flag set using type marker 010xxxxx
	ReqTypeDebug = "debug"
	// ReqTypeStressTest indicates requests with stress test flag set using type marker 011xxxxx
	ReqTypeStressTest = "stress_test"
	// ReqTypeShadow indicates requests with shadow flag set, using type marker 100xxxxx
	ReqTypeShadow = "shadow"
	// ReqTypeUnknown indicates requests without a known type marker
	ReqTypeUnknown = "unknown"
)

var (
	errConvertToSpanContext = errors.New("cannot convert to SpanContext")
)

// SpanContext represent context info of a Span, it is used to connect different Spans together into one specific trace.
type SpanContext interface {
	// TraceID get byte slice of traceID, DO NOT modify the returned slice
	TraceID() []byte
	// TraceIDString get hex string of traceID
	TraceIDString() string
	// SpanID get byte slice of spanID, DO NOT modify the returned slice
	SpanID() []byte
	// SpanIDString get hex string of spanID
	SpanIDString() string
	// ParentID get byte slice of parentID, DO NOT modify the returned slice
	ParentID() []byte
	// ParentIDString get hex string of parentID
	ParentIDString() string
	// Bytes get byte slice of SpanContext: {traceID}{spanID}{parentID}, DO NOT modify the returned slice
	Bytes() []byte
	// String get string of SpanContext: {traceIDString}:{spanIDString}:{parentIDString}
	String() string
	// IsDebug indicates whether it is on debug mode
	// Deprecated, use func IsSpanContextDebug instead.
	IsDebug() bool
	// IsSampled indicates whether it is to be sampled, can be sampled when either sampling flag or debug flag is true
	// Deprecated, use func IsSpanContextSampled instead.
	IsSampled() bool
	// IsFromStressTest indicates whether the request is part of a stress test
	// Deprecated, use func IsSpanContextFromStressTest instead.
	IsFromStressTest() bool
	// NewChildSpanContext generate a child SpanContext based on current one
	NewChildSpanContext() SpanContext
	// GetTypeMarker returns the type marker indicating the type of request.
	// Deprecated, use func GetTypeMarker instead.
	GetTypeMarker() int
}

type spanContext struct {
	mutex           sync.Mutex
	childSequenceID uint16 // next SequenceID of children
	id              [totalIDSize]byte
}

// TraceID get byte slice of traceID
// DO NOT modify the returned slice
func (sc *spanContext) TraceID() []byte {
	return sc.id[:traceIDSize]
}

// SpanID get byte slice of spanID
// DO NOT modify the returned slice
func (sc *spanContext) SpanID() []byte {
	return sc.id[traceIDSize : traceIDSize+spanIDSize]
}

// ParentID get byte slice of parentID
// DO NOT modify the returned slice
func (sc *spanContext) ParentID() []byte {
	return sc.id[traceIDSize+spanIDSize:]
}

// TraceIDString get hex string of traceID
func (sc *spanContext) TraceIDString() string {
	var dst [traceIDSize * 2]byte
	hex.Encode(dst[:], sc.TraceID())
	return string(dst[:])
}

// SpanIDString get hex string of spanID
func (sc *spanContext) SpanIDString() string {
	var dst [spanIDSize * 2]byte
	hex.Encode(dst[:], sc.SpanID())
	return string(dst[:])
}

// ParentIDString get hex string of parentID
func (sc *spanContext) ParentIDString() string {
	var dst [spanIDSize * 2]byte
	hex.Encode(dst[:], sc.ParentID())
	return string(dst[:])
}

// Bytes get byte slice of SpanContext: {traceID}{spanID}{parentID}
// DO NOT modify the returned slice
func (sc *spanContext) Bytes() []byte {
	return sc.id[:]
}

// String get string of SpanContext: {traceID}:{spanID}:{parentID}
func (sc *spanContext) String() string {
	var rBytes [totalIDSize*2 + 2]byte

	hex.Encode(rBytes[:], sc.TraceID())
	rBytes[hex.EncodedLen(traceIDSize)] = ':'

	hex.Encode(rBytes[hex.EncodedLen(traceIDSize)+1:], sc.SpanID())
	rBytes[hex.EncodedLen(traceIDSize+spanIDSize)+1] = ':'

	hex.Encode(rBytes[hex.EncodedLen(traceIDSize+spanIDSize)+2:], sc.ParentID())

	return string(rBytes[:])
}

// IsDebug indicates whether it is on debug mode
func (sc *spanContext) IsDebug() bool {
	if isOldFormat(sc) {
		// old format, checking debug bit
		if sc.id[traceIDSize-1]&traceFlagOldDebug == traceFlagOldDebug {
			return true
		}
	}
	// new format, checking type marker
	return sc.id[traceIDSize-1]&typeMarkerMask == typeMarkerForDebug
}

// IsFromStressTest indicates whether the request is part of a stress test
func (sc *spanContext) IsFromStressTest() bool {
	return sc.id[traceIDSize-1]&typeMarkerMask == typeMarkerForStressTest
}

// IsSampled indicates whether it is to be sampled, can be sampled when either sampling flag or debug flag is true
func (sc *spanContext) IsSampled() bool {
	// SpanContextID generated without this lib could have random flags which leads to the high probability of sampling
	// This is a rough version check to reduce the probability of this kind of sampling
	return sc.id[traceIDSize-1]&traceFlagSampled == traceFlagSampled || sc.IsDebug()
}

func (sc *spanContext) level() uint8 {
	return sc.id[traceIDSize]
}

// NewChildSpanContext generate a child SpanContext based on current one
func (sc *spanContext) NewChildSpanContext() SpanContext {
	sc.mutex.Lock()
	currentSequenceID := sc.childSequenceID
	sc.childSequenceID++
	sc.mutex.Unlock()

	var childID [totalIDSize]byte
	copy(childID[:], sc.TraceID())
	newSpanID(childID[traceIDSize:traceIDSize+spanIDSize], sc.level()+1, currentSequenceID)
	copy(childID[traceIDSize+spanIDSize:], sc.SpanID())

	childSC := &spanContext{
		id: childID,
	}
	return childSC
}

// GetTypeMarker return the type marker for the request
func (sc *spanContext) GetTypeMarker() int {
	return int((sc.id[traceIDSize-1] & typeMarkerMask) >> flagBitsNonTypeMarker)
}

// NewSpanContextFromBytes reconstructs the SpanContext encoded in the byte slice
func NewSpanContextFromBytes(data []byte) (SpanContext, error) {
	if len(data) != totalIDSize {
		return nil, errConvertToSpanContext
	}
	var tmp [totalIDSize]byte
	copy(tmp[:], data)
	return &spanContext{id: tmp}, nil
}

// NewSpanContextFromString reconstructs the SpanContext encoded in a string
func NewSpanContextFromString(data string) (SpanContext, error) {
	if len(data) == 0 {
		return nil, errConvertToSpanContext
	}
	firstColonIdx := strings.Index(data, ":")
	lastColonIdx := strings.LastIndex(data, ":")
	if firstColonIdx >= lastColonIdx {
		return nil, errConvertToSpanContext
	}

	var hBytes [totalIDSize]byte
	_, err := hex.Decode(hBytes[:], []byte(data[:firstColonIdx]))
	if err != nil {
		return nil, err
	}

	_, err = hex.Decode(hBytes[traceIDSize:], []byte(data[firstColonIdx+1:lastColonIdx]))
	if err != nil {
		return nil, err
	}

	_, err = hex.Decode(hBytes[traceIDSize+spanIDSize:], []byte(data[lastColonIdx+1:]))
	if err != nil {
		return nil, err
	}

	return &spanContext{id: hBytes}, nil
}

// SpanContextFromBytesToString converts the bytes which represents a SpanContext to a string
func SpanContextFromBytesToString(bytes []byte) string {
	spanContext, err := NewSpanContextFromBytes(bytes)
	if err != nil {
		return ""
	}

	return spanContext.String()
}

func getRandomBytes(bs []byte) {
	if len(bs) == 0 {
		return
	}

	bsLen, _ := rand.Read(bs)
	// avoid confuse with ancestor's parent span id
	if bs[bsLen-1] == 0 {
		isAllZero := true
		for i := 0; i < bsLen-1 && bs[i] != 0; i++ {
			isAllZero = false
		}
		if isAllZero {
			bs[bsLen-1] = 1
		}
	}
}

// newSpanID create a SpanID
func newSpanID(spanID []byte, level uint8, sequenceID uint16) {
	spanID[0] = level
	if sequenceID > 0 {
		spanID[1], spanID[2] = byte(sequenceID>>8), byte(sequenceID)
	}
	getRandomBytes(spanID[3:spanIDSize])
}

// Span represents an execution sequence in your code logic, from where it is started to the place its Finish been called.
// You can set tags and additional logs on it, both are key value pairs and will be displayed later on tracing ui.
type Span interface {

	// Context returns SpanContext of current Span.
	Context() SpanContext

	// NewChildSpan creates and returns a child Span of current Span.
	NewChildSpan(name string) (Span, error)

	// SetTag adds a tag to the span.
	// The key of Tag is unique and applies to the whole lifecycle of the Span
	SetTag(key string, value interface{}) Span

	// SetTags set tags to current Span
	// The parameters must appear as key-value pairs, each key-value pair is a tag. Like:
	// 		SetTags(
	// 			"type", "http",
	//			"debug" true,
	// 		)
	// The keys must all be strings
	SetTags(keyValues ...interface{}) Span

	// SetDebugTags is similar with SetTags, the only difference is these tags will only be applicable
	// when debug flag of current Span is toggled on.
	SetDebugTags(keyValues ...interface{}) Span

	// LogFields uses key:value logging data to print a log entry to current Span.
	// The parameters must appear as key-value pairs, each key-value pair is a logging field. For example:
	//		LogFields(
	//			"event", "soft error",
	//			"type", "cache timeout",
	//			"waited.millis", 1500)
	// Every Log entry has a specific timestamp, it can be used to print intermediate results, data or debug infos.
	LogFields(keyValues ...interface{}) Span

	// LogDebugFields is similar with LogFields, the only difference is these logs will only be applicable
	// when debug flag of current Span is toggled on.
	LogDebugFields(keyValues ...interface{}) Span

	// Finish will end the current Span and report it to tracing platform.
	Finish()

	// FinishWithOptions will end the current Span and report it to tracing platform with options.
	FinishWithOptions(options ...FinishSpanOption)
}

// FinishSpanOptions saves options when span is finished
type FinishSpanOptions struct {
	FinishTime time.Time
}

// FinishSpanOption is a function that sets some options when finish a span
type FinishSpanOption func(opts *FinishSpanOptions)

// FinishTime sets finish time for span
func FinishTime(time time.Time) FinishSpanOption {
	return func(opts *FinishSpanOptions) {
		opts.FinishTime = time
	}
}
