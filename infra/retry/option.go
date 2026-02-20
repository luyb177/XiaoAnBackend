package retry

import (
	"time"

	"github.com/zeromicro/go-zero/core/logx"
)

// Option 定义

func defaultConfig() *Config {
	return &Config{
		maxAttempts: 3,
		baseDelay:   100 * time.Millisecond,
		maxDelay:    2 * time.Second,
		jitter:      false,
		retryIf:     func(error) bool { return true },
		onRetry: func(attempt int, err error, nextDelay time.Duration) {
			logx.Errorf("retry %d failed: %v, next retry in %v", attempt, err, nextDelay)
		},
	}
}

func WithMaxAttempts(n int) Option {
	return func(c *Config) {
		if n > 0 {
			c.maxAttempts = n
		}
	}
}
func WithBaseDelay(d time.Duration) Option {
	return func(c *Config) {
		if d > 0 {
			c.baseDelay = d
		}
	}
}
func WithMaxDelay(d time.Duration) Option {
	return func(c *Config) {
		if d > 0 {
			c.maxDelay = d
		}
	}
}
func WithRetryIf(fn func(error) bool) Option {
	return func(c *Config) {
		if fn != nil {
			c.retryIf = fn
		}
	}
}
func WithJitter() Option {
	return func(c *Config) {
		c.jitter = true
	}
}

func WithOnRetry(fn func(attempt int, err error, nextDelay time.Duration)) Option {
	return func(c *Config) {
		c.onRetry = fn
	}
}
