package retry

import (
	"context"
	"time"
)

// Do + config

type Config struct {
	maxAttempts int
	baseDelay   time.Duration
	maxDelay    time.Duration

	jitter  bool
	retryIf func(err error) bool                                  // 用于判断是否应该在特定错误发生时进行重试
	onRetry func(attempt int, err error, nextDelay time.Duration) // 重试回调函数，在每次重试前调用，提供当前尝试次数、错误信息和下一次重试的延迟时间
}

type Option func(config *Config)

func Do(ctx context.Context, fn func() error, opts ...Option) error {
	cfg := defaultConfig()
	for _, opt := range opts {
		opt(cfg)
	}

	delay := cfg.baseDelay

	for attempt := 1; attempt <= cfg.maxAttempts; attempt++ {
		err := fn()
		if err == nil {
			return nil
		}

		if attempt == cfg.maxAttempts || !cfg.retryIf(err) {
			// 最后一次失败或不满足重试条件，返回错误
			return err
		}

		sleep := delay
		if cfg.jitter {
			sleep = withJitter(delay)
		}
		if sleep > cfg.maxDelay {
			sleep = cfg.maxDelay
		}

		if cfg.onRetry != nil {
			cfg.onRetry(attempt, err, sleep)
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(sleep):
			// 继续下一次重试
		}

		// 增加延迟时间
		delay *= 2
		if delay > cfg.maxDelay {
			delay = cfg.maxDelay
		}
	}
	return nil
}
