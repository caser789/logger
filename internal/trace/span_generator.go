package trace

import (
	"context"
	"crypto/md5"
	"math/rand"
	"os"
	"time"
)

// SpanContextGenerator a generator to produce new SpanContext
type SpanContextGenerator interface {
	NewSpanContext(options ...SpanContextOption) SpanContext
}

type cachedSpanContextGenerator struct {
	serviceInstanceID string
	instanceIDHash    [4]byte
	sampler           Sampler
}

// NewSpanContext produce SpanContext with options
func (scg *cachedSpanContextGenerator) NewSpanContext(options ...SpanContextOption) SpanContext {
	sco := SpanContextOptions{
		IsDebug:          false,
		IsFromStressTest: false,
		IsShadow:         false,
		IsCritical:       false,
	}
	for _, f := range options {
		f(&sco)
	}

	var traceFlag byte
	traceFlag = setTypeMarker(sco, traceFlag)
	traceFlag = setSingleFlags(sco, traceFlag, scg)

	sc := spanContext{
		childSequenceID: 0,
	}
	scg.newSpanContextID(sc.id[:], traceFlag)

	return &sc
}

func setTypeMarker(sco SpanContextOptions, traceFlag byte) byte {
	if sco.IsDebug {
		// Temporarily use the old debug flag until all services have migrated to v1.4+
		traceFlag |= traceFlagOldDebug
	} else if sco.IsFromStressTest {
		traceFlag = traceFlag&(^typeMarkerMask) | typeMarkerForStressTest
	} else if sco.IsShadow {
		traceFlag = traceFlag&(^typeMarkerMask) | typeMarkerForShadow
	}
	// Temporarily suppress until all services have migrated to v1.4+
	// else {
	// 	traceFlag = traceFlag&(^typeMarkerMask) | typeMarkerForNormal
	// }

	return traceFlag
}

func setSingleFlags(sco SpanContextOptions, traceFlag byte, scg *cachedSpanContextGenerator) byte {
	// 1. handle sampling flag
	if sco.IsSampled != nil {
		if *sco.IsSampled {
			traceFlag |= traceFlagSampled
		}
	} else if scg.sampler.IsSampled(context.Background()) {
		traceFlag |= traceFlagSampled
	}

	// 2. handle critical flag
	if sco.IsCritical {
		traceFlag |= traceFlagCritical
	}

	return traceFlag
}

// newSpanContextID create a new id for SpanContext
func (scg *cachedSpanContextGenerator) newSpanContextID(scID []byte, flag byte) {
	// Generate trace ID
	// 4 bytes serviceHash
	copy(scID[:], scg.instanceIDHash[:])
	// 6 bytes timestamp
	timestamp := time.Now().UnixNano() / 1000
	scID[4] = byte(timestamp >> 40)
	scID[5] = byte(timestamp >> 32)
	scID[6] = byte(timestamp >> 24)
	scID[7] = byte(timestamp >> 16)
	scID[8] = byte(timestamp >> 8)
	scID[9] = byte(timestamp)
	// 5 bytes randomID
	getRandomBytes(scID[10 : traceIDSize-1])
	// 1 byte special flag
	scID[traceIDSize-1] = flag

	// Generate spanID
	newSpanID(scID[traceIDSize:], 0, 0)
}

// GeneratorOptions are options to create a new SpanContextGenerator
type GeneratorOptions struct {
	sampler Sampler
}

// GeneratorOption is modifier to update GeneratorOptions
type GeneratorOption func(options *GeneratorOptions)

// WithSampler sets GeneratorOptions.sampler
func WithSampler(sampler Sampler) GeneratorOption {
	return func(options *GeneratorOptions) {
		options.sampler = sampler
	}
}

// NewSpanContextGenerator construct a SpanContextGenerator with cashed instanceID hash
func NewSpanContextGenerator(serviceInstanceID string, options ...GeneratorOption) SpanContextGenerator {
	// combine pid and timestamp as seed
	rand.Seed((int64(os.Getpid()) << 32) + time.Now().UnixNano())

	ops := GeneratorOptions{}
	for _, op := range options {
		op(&ops)
	}

	var siHash [4]byte
	if len(serviceInstanceID) == 0 {
		rand.Read(siHash[:])
	} else {
		siMD5 := md5.Sum([]byte(serviceInstanceID))
		copy(siHash[:], siMD5[md5.Size-4:])
	}

	sampler := ops.sampler
	if sampler == nil {
		sampler = NewProbabilisticSampler(defaultSamplingProbability)
	}

	return &cachedSpanContextGenerator{
		serviceInstanceID: serviceInstanceID,
		instanceIDHash:    siHash,
		sampler:           sampler,
	}
}

// SpanContextOptions are options to create a new SpanContext
type SpanContextOptions struct {
	IsDebug          bool
	IsFromStressTest bool
	IsShadow         bool

	// IsSampled contains the sampling decision
	// If it's nil, the default sampling strategy applies when creating new SpanContext
	IsSampled  *bool
	IsCritical bool
}

// SpanContextOption is modifier to update SpanContextOptions
type SpanContextOption func(options *SpanContextOptions)

// IsDebug sets SpanContextOption.IsDebug
func IsDebug(isDebug bool) SpanContextOption {
	return func(options *SpanContextOptions) {
		options.IsDebug = isDebug
	}
}

// IsFromStressTest sets SpanContextOption.IsFromStressTest
func IsFromStressTest(isFromStressTest bool) SpanContextOption {
	return func(options *SpanContextOptions) {
		options.IsFromStressTest = isFromStressTest
	}
}

// IsShadow sets SpanContextOption.IsShadow
func IsShadow(isShadow bool) SpanContextOption {
	return func(options *SpanContextOptions) {
		options.IsShadow = isShadow
	}
}

// IsSampled sets SpanContextOption.IsSampled
func IsSampled(isSampled *bool) SpanContextOption {
	return func(options *SpanContextOptions) {
		options.IsSampled = isSampled
	}
}

// IsCritical sets SpanContextOption.IsCritical
func IsCritical(isCritical bool) SpanContextOption {
	return func(options *SpanContextOptions) {
		options.IsCritical = isCritical
	}
}
