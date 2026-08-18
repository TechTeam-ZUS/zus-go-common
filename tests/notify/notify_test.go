package notify_test

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/TechTeam-ZUS/zus-go-common/notify"
)

// timeLayout mirrors notify's internal wrap() timestamp format.
const timeLayout = "2006-01-02T15:04:05.000Z07:00"

// assertWrappedContent checks content is "Time: <timeLayout>\n<want>" —
// a real newline, not the literal two-character "\n".
func assertWrappedContent(t *testing.T, content, want string) {
	t.Helper()
	lines := strings.SplitN(content, "\n", 2)
	require.Len(t, lines, 2, "content must contain a real newline")
	assert.True(t, strings.HasPrefix(lines[0], "Time: "))
	_, err := time.Parse(timeLayout, strings.TrimPrefix(lines[0], "Time: "))
	assert.NoError(t, err, "timestamp must match logger's text handler layout")
	assert.Equal(t, want, lines[1])
}

// newCapturingServer returns a test server that records the last request
// body and lets the test control the response status.
func newCapturingServer(t *testing.T, status int) (*httptest.Server, *map[string]any) {
	t.Helper()
	var captured map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.NoError(t, json.NewDecoder(r.Body).Decode(&captured))
		w.WriteHeader(status)
	}))
	t.Cleanup(srv.Close)
	return srv, &captured
}

func TestBot_Notify(t *testing.T) {
	tests := []struct {
		name             string
		level            slog.Level
		expectedTemplate string
	}{
		{name: "debug uses plain card", level: slog.LevelDebug, expectedTemplate: notify.TemplateBlue},
		{name: "info uses plain card", level: slog.LevelInfo, expectedTemplate: notify.TemplateBlue},
		{name: "warn uses alert card, orange", level: slog.LevelWarn, expectedTemplate: notify.TemplateOrange},
		{name: "error uses alert card, red", level: slog.LevelError, expectedTemplate: notify.TemplateRed},
		{name: "above error is red", level: slog.Level(12), expectedTemplate: notify.TemplateRed},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv, captured := newCapturingServer(t, http.StatusOK)
			notify.NewBot(srv.URL, "svc")

			notify.GetBot().Notify(tt.level, "title", "content")

			require.NotNil(t, *captured)
			card := (*captured)["card"].(map[string]any)
			header := card["header"].(map[string]any)
			assert.Equal(t, tt.expectedTemplate, header["template"])

			titleContent := header["title"].(map[string]any)["content"].(string)
			assert.Contains(t, titleContent, "svc")
			assert.Contains(t, titleContent, "title")

			content := card["elements"].([]any)[0].(map[string]any)["text"].(map[string]any)["content"].(string)
			assertWrappedContent(t, content, "content")
		})
	}
}

func TestBot_Notify_ReplacesSingleton(t *testing.T) {
	srv1, captured1 := newCapturingServer(t, http.StatusOK)
	srv2, captured2 := newCapturingServer(t, http.StatusOK)

	notify.NewBot(srv1.URL, "svc")
	notify.GetBot().Notify(slog.LevelInfo, "t", "m")
	assert.NotNil(t, *captured1)
	assert.Nil(t, *captured2)

	notify.NewBot(srv2.URL, "svc")
	notify.GetBot().Notify(slog.LevelInfo, "t", "m")
	assert.NotNil(t, *captured2)
}

func TestBot_Notify_DoesNotPanicOnErrorResponse(t *testing.T) {
	srv, _ := newCapturingServer(t, http.StatusInternalServerError)
	notify.NewBot(srv.URL, "svc")

	assert.NotPanics(t, func() { notify.GetBot().Notify(slog.LevelWarn, "t", "m") })
}

func TestNewCard(t *testing.T) {
	c := notify.NewCard("My Title", notify.TemplateRed, "**bold** content")

	data, err := json.Marshal(c)
	require.NoError(t, err)

	var got map[string]any
	require.NoError(t, json.Unmarshal(data, &got))

	assert.Equal(t, "interactive", got["msg_type"])
	card := got["card"].(map[string]any)
	assert.Equal(t, true, card["config"].(map[string]any)["wide_screen_mode"])
	assert.Equal(t, "red", card["header"].(map[string]any)["template"])
	assert.Equal(t, "plain_text", card["header"].(map[string]any)["title"].(map[string]any)["tag"])
	elements := card["elements"].([]any)
	require.Len(t, elements, 1)
	el := elements[0].(map[string]any)
	assert.Equal(t, "div", el["tag"])
	assert.Equal(t, "lark_md", el["text"].(map[string]any)["tag"])
	assertWrappedContent(t, el["text"].(map[string]any)["content"].(string), "**bold** content")
}
