package middleware

import (
	"bytes"
	"log/slog"
	"net/http"
	"time"

	"github.com/TechTeam-ZUS/zus-go-common/logger"
)

const (
	RequestStartTrace = "REQUEST_STARTED"
	RequestEndTrace   = "REQUEST_ENDED"
)

// Field is an extra request-scoped log field, e.g. userID, app version, route pattern.
// Consumers own how it's derived (chi route pattern, JWT claims, custom headers, etc.).
type Field struct {
	Key   string
	Value func(r *http.Request) any
}

// HttpRequestLogStruct configures HttpRequestLog. Set it once via Configure,
// before registering HttpRequestLog as middleware.
type HttpRequestLogStruct struct {
	// RequestID extracts the request ID (default: X-Request-Id header).
	RequestID func(r *http.Request) string
	// Fields are extra fields logged on both the start and end events, in order.
	Fields []Field
	// LogBody captures the response body on the end-of-request log.
	LogBody bool
}

var httpRequestLogCfg = HttpRequestLogStruct{RequestID: defaultRequestID}

// Configure sets the request-log configuration. Call once, before registering
// HttpRequestLog as middleware.
func ConfigureHttpRequestLog(c HttpRequestLogStruct) {
	if c.RequestID == nil {
		c.RequestID = defaultRequestID
	}

	httpRequestLogCfg = c
}

func defaultRequestID(r *http.Request) string {
	return r.Header.Get("X-Request-Id")
}

// HttpRequestLog is stdlib-compatible middleware (func(http.Handler) http.Handler)
// that logs a start and end event per request, and attaches a request-scoped
// logger to the request context (retrievable via logger.FromContext).
func HttpRequestLog(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		ww := &responseRecorder{ResponseWriter: w, status: http.StatusOK, captureBody: httpRequestLogCfg.LogBody}

		reqLogger := logger.GetLogger().With(
			slog.String("requestId", httpRequestLogCfg.RequestID(r)),
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
		)

		fields := make([]any, 0, len(httpRequestLogCfg.Fields)*2)
		for _, f := range httpRequestLogCfg.Fields {
			fields = append(fields, f.Key, f.Value(r))
		}

		reqLogger.Info(RequestStartTrace, slog.Group("context", fields...))

		ctx := logger.WithContext(r.Context(), reqLogger)
		next.ServeHTTP(ww, r.WithContext(ctx))

		endFields := append(fields,
			"http_status", ww.status,
			"duration_ms", float64(time.Since(start).Microseconds())/1000,
			"size_kb", float64(ww.size)/1024,
		)
		if httpRequestLogCfg.LogBody {
			endFields = append(endFields, "response_body", ww.body.String())
		}

		reqLogger.Info(RequestEndTrace, slog.Group("context", endFields...))
	})
}

// responseRecorder wraps http.ResponseWriter to capture status, size, and optionally the body.
type responseRecorder struct {
	http.ResponseWriter
	status      int
	size        int
	captureBody bool
	body        bytes.Buffer
}

func (rr *responseRecorder) WriteHeader(status int) {
	rr.status = status
	rr.ResponseWriter.WriteHeader(status)
}

func (rr *responseRecorder) Write(b []byte) (int, error) {
	if rr.captureBody {
		rr.body.Write(b)
	}
	n, err := rr.ResponseWriter.Write(b)
	rr.size += n
	return n, err
}
