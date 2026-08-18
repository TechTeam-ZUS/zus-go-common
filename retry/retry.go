package retry

import (
	"time"

	"github.com/TechTeam-ZUS/zus-go-common/logger"
)

const RetryDelay = 2 * time.Second

// Do calls fn up to attempts times, waiting delay between tries.
// Returns nil on first success, or the last error if every attempt fails.
func Do(attempts int, delay time.Duration, fn func() error) error {
	if attempts < 1 {
		attempts = 1
	}

	var err error
	for i := 0; i < attempts; i++ {
		if err = fn(); err == nil {
			return nil
		}

		logger.Warn("Retrying, attempts: %d/%d", i+1, attempts)
		if i < attempts-1 {
			time.Sleep(delay)
		}
	}
	return err
}
