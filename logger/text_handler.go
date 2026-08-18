package logger

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"sync"
)

const timeLayout = "2006-01-02T15:04:05.000Z07:00"

// textHandler formats log lines as "[service] time LEVEL : message key=value ...",
// pulling the "service" attr (set via slog.Logger.With) into the bracket prefix.
type textHandler struct {
	mu    *sync.Mutex
	w     io.Writer
	level slog.Leveler
	attrs []slog.Attr
}

func newTextHandler(w io.Writer, level slog.Leveler) *textHandler {
	return &textHandler{mu: &sync.Mutex{}, w: w, level: level}
}

func (h *textHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.level.Level()
}

func (h *textHandler) Handle(_ context.Context, r slog.Record) error {
	var service string
	extra := make([]string, 0, len(h.attrs)+r.NumAttrs())

	collect := func(a slog.Attr) bool {
		if a.Key == "service" {
			service = a.Value.String()
		} else {
			extra = append(extra, fmt.Sprintf("%s=%v", a.Key, a.Value.Any()))
		}
		return true
	}
	for _, a := range h.attrs {
		collect(a)
	}
	r.Attrs(collect)

	line := fmt.Sprintf("[%s] %s %s : %s", service, r.Time.Format(timeLayout), r.Level.String(), r.Message)
	if len(extra) > 0 {
		line += " " + strings.Join(extra, " ")
	}

	h.mu.Lock()
	defer h.mu.Unlock()
	_, err := fmt.Fprintln(h.w, line)
	return err
}

func (h *textHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	newAttrs := make([]slog.Attr, 0, len(h.attrs)+len(attrs))
	newAttrs = append(newAttrs, h.attrs...)
	newAttrs = append(newAttrs, attrs...)
	return &textHandler{mu: h.mu, w: h.w, level: h.level, attrs: newAttrs}
}

func (h *textHandler) WithGroup(_ string) slog.Handler {
	return h
}
