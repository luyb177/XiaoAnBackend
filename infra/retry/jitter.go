package retry

import (
	"math/rand/v2"
	"time"
)

// 抖动策略

func withJitter(d time.Duration) time.Duration {
	// 0.5x ~ 1.5x
	return d/2 + time.Duration(rand.Float64()*float64(d))
}
