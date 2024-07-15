package fal

import (
	"math"
	"math/rand"
	"time"
)

type Backoff interface {
	NextDelay(retries int) time.Duration
}

type ConstantBackOff struct {
	Base   time.Duration
	Jitter time.Duration
}

func (b *ConstantBackOff) NextDelay(_ int) time.Duration {
	jitter := time.Duration(rand.Float64() * float64(b.Jitter))
	return b.Base + jitter
}

type ExponentialBackOff struct {
	Base       time.Duration
	Jitter     time.Duration
	Multiplier float64
}

func (d *ExponentialBackOff) NextDelay(retries int) time.Duration {
	jitter := time.Duration(rand.Float64() * float64(d.Jitter))
	return time.Duration(float64(d.Base)*math.Pow(d.Multiplier, float64(retries))) + jitter
}
