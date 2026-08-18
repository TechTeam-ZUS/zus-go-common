package retry_test

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/TechTeam-ZUS/zus-go-common/retry"
)

func TestDo(t *testing.T) {
	tests := []struct {
		name           string
		attempts       int
		failuresBefore int
		expectedErr    bool
		expectedCalls  int
	}{
		{name: "succeeds first try", attempts: 3, failuresBefore: 0, expectedErr: false, expectedCalls: 1},
		{name: "succeeds after 2 failures", attempts: 3, failuresBefore: 2, expectedErr: false, expectedCalls: 3},
		{name: "exhausts all attempts", attempts: 3, failuresBefore: 5, expectedErr: true, expectedCalls: 3},
		{name: "attempts < 1 treated as 1", attempts: 0, failuresBefore: 5, expectedErr: true, expectedCalls: 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			calls := 0
			err := retry.Do(tt.attempts, time.Millisecond, func() error {
				calls++
				if calls <= tt.failuresBefore {
					return errors.New("boom")
				}
				return nil
			}, "svc")

			if tt.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.expectedCalls, calls)
		})
	}
}
