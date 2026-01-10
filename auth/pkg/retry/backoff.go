package retry

import "time"

// ExponentialBackoffRetry 指数退避
func ExponentialBackoffRetry(maxAttempts int, baseDelay time.Duration, maxDelay time.Duration, fn func() error) error {
	delay := baseDelay
	for i := 0; i < maxAttempts; i++ {
		err := fn()
		if err == nil {
			return nil
		}

		// 最后一次失败
		if i == maxAttempts-1 {
			return err
		}

		time.Sleep(delay)

		delay = delay * 2
		if delay > maxDelay {
			delay = maxDelay
		}
	}
	return nil
}
