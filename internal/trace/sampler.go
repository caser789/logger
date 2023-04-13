package trace

import (
	"context"
	"fmt"
	"math"
	"math/rand"
)

const defaultSamplingProbability = 0.001

// Sampler decides whether a new trace should be sampled or not.
type Sampler interface {
	// IsSampled decides whether a trace with given `context` should be sampled.
	IsSampled(ctx context.Context) bool

	// Close does a clean shutdown of the sampler, stopping any background go-routines it may have started.
	Close()
}

// ProbabilisticSampler is a sampler that randomly samples a certain percentage of traces specified by the
// samplingRate, in the range between 0.0 and 1.0.
type ProbabilisticSampler struct {
	samplingRate float64
}

// IsSampled implements IsSampled() of Sampler.
func (s *ProbabilisticSampler) IsSampled(_ context.Context) bool {
	return rand.Float64() < s.samplingRate
}

// Close implements Close() of Sampler.
func (s *ProbabilisticSampler) Close() {}

// SamplingRate returns the sampling probability this sampled was constructed with.
func (s *ProbabilisticSampler) SamplingRate() float64 {
	return s.samplingRate
}

// String is used to log sampler details.
func (s *ProbabilisticSampler) String() string {
	return fmt.Sprintf("ProbabilisticSampler(samplingRate=%v)", s.samplingRate)
}

// NewProbabilisticSampler creates a ProbabilisticSampler
func NewProbabilisticSampler(samplingRate float64) *ProbabilisticSampler {
	return &ProbabilisticSampler{
		samplingRate: math.Max(0.0, math.Min(samplingRate, 1.0)),
	}
}
