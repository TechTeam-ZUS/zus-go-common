package middleware_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/TechTeam-ZUS/zus-go-common/logger"
	"github.com/TechTeam-ZUS/zus-go-common/middleware"
)

// captureLogOutput redirects the shared logger to a JSON handler writing to
// a pipe, runs fn, and returns each emitted log line parsed as JSON.
func captureLogOutput(t *testing.T, fn func()) []map[string]any {
	t.Helper()
	t.Setenv("LOG_HANDLER_TYPE", "json")

	r, w, err := os.Pipe()
	require.NoError(t, err)
	origStdout := os.Stdout
	os.Stdout = w
	logger.Init()

	fn()

	w.Close()
	os.Stdout = origStdout
	out, err := io.ReadAll(r)
	require.NoError(t, err)

	var lines []map[string]any
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if line == "" {
			continue
		}
		var m map[string]any
		require.NoError(t, json.Unmarshal([]byte(line), &m))
		lines = append(lines, m)
	}
	return lines
}

func TestHttpRequestLog(t *testing.T) {
	tests := []struct {
		name            string
		cfg             middleware.HttpRequestLogStruct
		reqIDHeader     string
		expectedStatus  int
		expectedBody    string
		expectedReqID   string
		expectedContext map[string]any
	}{
		{
			name:           "default config uses X-Request-Id header",
			cfg:            middleware.HttpRequestLogStruct{},
			reqIDHeader:    "req-1",
			expectedStatus: http.StatusOK,
			expectedBody:   "ok",
			expectedReqID:  "req-1",
		},
		{
			name: "custom RequestID and Fields appear in both log lines",
			cfg: middleware.HttpRequestLogStruct{
				RequestID: func(r *http.Request) string { return "custom-id" },
				Fields: []middleware.Field{
					{Key: "user", Value: func(r *http.Request) any { return "u-1" }},
					{Key: "app_version", Value: func(r *http.Request) any { return "2.0" }},
				},
			},
			expectedStatus:  http.StatusOK,
			expectedBody:    "ok",
			expectedReqID:   "custom-id",
			expectedContext: map[string]any{"user": "u-1", "app_version": "2.0"},
		},
		{
			name:           "LogBody true includes response_body in end log",
			cfg:            middleware.HttpRequestLogStruct{LogBody: true},
			expectedStatus: http.StatusOK,
			expectedBody:   "ok",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.expectedStatus)
				w.Write([]byte(tt.expectedBody))
			})

			req := httptest.NewRequest(http.MethodGet, "/path", nil)
			if tt.reqIDHeader != "" {
				req.Header.Set("X-Request-Id", tt.reqIDHeader)
			}
			rec := httptest.NewRecorder()

			lines := captureLogOutput(t, func() {
				middleware.HttpRequestLog(tt.cfg)(next).ServeHTTP(rec, req)
			})

			require.Len(t, lines, 2)
			startLog, endLog := lines[0], lines[1]

			assert.Equal(t, middleware.RequestStartTrace, startLog["msg"])
			assert.Equal(t, middleware.RequestEndTrace, endLog["msg"])

			if tt.expectedReqID != "" {
				assert.Equal(t, tt.expectedReqID, startLog["requestId"])
				assert.Equal(t, tt.expectedReqID, endLog["requestId"])
			}

			startCtx, _ := startLog["context"].(map[string]any)
			endCtx, _ := endLog["context"].(map[string]any)
			for k, v := range tt.expectedContext {
				assert.Equal(t, v, startCtx[k], "start log context[%s]", k)
				assert.Equal(t, v, endCtx[k], "end log context[%s]", k)
			}

			assert.Equal(t, float64(tt.expectedStatus), endCtx["http_status"])
			if tt.cfg.LogBody {
				assert.Equal(t, tt.expectedBody, endCtx["response_body"])
			} else {
				assert.NotContains(t, endCtx, "response_body")
			}

			assert.Equal(t, tt.expectedStatus, rec.Code)
			assert.Equal(t, tt.expectedBody, rec.Body.String())
		})
	}
}

func TestHttpRequestLog_InjectsRequestScopedLogger(t *testing.T) {
	var gotLogger any
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotLogger = logger.FromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/path", nil)
	rec := httptest.NewRecorder()

	middleware.HttpRequestLog(middleware.HttpRequestLogStruct{})(next).ServeHTTP(rec, req)

	assert.NotNil(t, gotLogger)
}

func TestJSONResponseMiddleware(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	middleware.JSONResponseMiddleware(next).ServeHTTP(rec, req)

	assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))
}
