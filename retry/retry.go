package retry

import (
	"time"

	"github.com/TechTeam-ZUS/zus-go-common/logger"
)

const RetryDelay = 2 * time.Second

// Do calls fn up to attempts times, waiting delay between tries. name to identifies what's being retried in the log line.
// Returns nil on first success, or the last error if every attempt fails.
func Do(attempts int, delay time.Duration, fn func() error, name string) error {
	if attempts < 1 {
		attempts = 1
	}

	var err error
	for i := 0; i < attempts; i++ {
		if err = fn(); err == nil {
			return nil
		}

		logger.Warnf("Retrying %s (attempt %d/%d): %s", name, i+1, attempts, err)
		if i < attempts-1 {
			time.Sleep(delay)
		}
	}
	return err
}
