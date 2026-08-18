package logger_test

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/TechTeam-ZUS/zus-go-common/logger"
)

// GetLogger has no varying input/output to table-drive; assert directly.
func TestGetLogger(t *testing.T) {
	l1 := logger.GetLogger()
	l2 := logger.GetLogger()

	assert.NotNil(t, l1)
	assert.Same(t, l1, l2)
}

func TestInit(t *testing.T) {
	tests := []struct {
		name           string
		handlerType    string
		expectedIsJSON bool
	}{
		{name: "text handler", handlerType: "text", expectedIsJSON: false},
		{name: "json handler", handlerType: "json", expectedIsJSON: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("LOG_HANDLER_TYPE", tt.handlerType)

			r, w, err := os.Pipe()
			require.NoError(t, err)
			origStdout := os.Stdout
			os.Stdout = w

			l := logger.Init()
			l.Info("test message")

			w.Close()
			os.Stdout = origStdout
			out, err := io.ReadAll(r)
			require.NoError(t, err)

			var js map[string]any
			isJSON := json.Unmarshal(out, &js) == nil

			assert.Equal(t, tt.expectedIsJSON, isJSON)
			assert.Same(t, l, logger.GetLogger())
		})
	}
}

func TestWithContextAndFromContext(t *testing.T) {
	tests := []struct {
		name         string
		setLogger    bool
		expectedSame bool
	}{
		{name: "logger stored in context is retrieved", setLogger: true, expectedSame: true},
		{name: "falls back to global logger when absent", setLogger: false, expectedSame: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			var scoped *slog.Logger
			if tt.setLogger {
				scoped = logger.GetLogger().With("scope", "request")
				ctx = logger.WithContext(ctx, scoped)
			}

			got := logger.FromContext(ctx)
			assert.NotNil(t, got)
			if tt.expectedSame {
				assert.Same(t, scoped, got)
			} else {
				assert.Same(t, logger.GetLogger(), got)
			}
		})
	}
}
