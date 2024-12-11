// internal/logging/logging.go

package logging

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/mattn/go-isatty"
	"gopkg.in/natefinch/lumberjack.v2"
)

// ColoredTextHandler wraps a TextHandler and adds color based on log level.
type ColoredTextHandler struct {
	handler slog.Handler
}

// NewColoredTextHandler creates a new ColoredTextHandler.
func NewColoredTextHandler(out *os.File, opts *slog.HandlerOptions) *ColoredTextHandler {
	textHandler := slog.NewTextHandler(out, opts)
	return &ColoredTextHandler{handler: textHandler}
}

// Enabled checks if the handler is enabled for the given level.
func (h *ColoredTextHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.handler.Enabled(ctx, level)
}

// Handle processes the log record with appropriate coloring.
func (h *ColoredTextHandler) Handle(ctx context.Context, r slog.Record) error {
	// Add color based on log level
	switch r.Level {
	case slog.LevelError:
		fmt.Fprint(os.Stdout, "\033[31m") // Red
	case slog.LevelWarn:
		fmt.Fprint(os.Stdout, "\033[33m") // Yellow
	case slog.LevelInfo:
		fmt.Fprint(os.Stdout, "\033[32m") // Green
	default:
		fmt.Fprint(os.Stdout, "\033[0m") // Reset
	}

	// Handle the log
	err := h.handler.Handle(ctx, r)

	// Reset color
	fmt.Fprint(os.Stdout, "\033[0m")
	return err
}

// WithAttrs returns a new handler with the given attributes.
func (h *ColoredTextHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &ColoredTextHandler{handler: h.handler.WithAttrs(attrs)}
}

// WithGroup returns a new handler with the given group name.
func (h *ColoredTextHandler) WithGroup(name string) slog.Handler {
	return &ColoredTextHandler{handler: h.handler.WithGroup(name)}
}

// ParseLogLevel converts a string log level to slog.Level.
func ParseLogLevel(levelStr string) slog.Level {
	switch strings.ToUpper(levelStr) {
	case "DEBUG":
		return slog.LevelDebug
	case "INFO":
		return slog.LevelInfo
	case "WARN":
		return slog.LevelWarn
	case "ERROR":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// InitializeLogger sets up the logger based on the pretty flag and terminal detection.
// If pretty is true and the output is a terminal, it uses a colored text handler.
// Otherwise, it defaults to a JSON handler suitable for Datadog.
func InitializeLogger(pretty bool) *slog.Logger {
	var handler slog.Handler

	isTerminal := isatty.IsTerminal(os.Stdout.Fd()) || isatty.IsCygwinTerminal(os.Stdout.Fd())

	logLevelStr := os.Getenv("LOG_LEVEL")
	if logLevelStr == "" {
		logLevelStr = "INFO"
	}
	logLevel := ParseLogLevel(logLevelStr)

	if pretty && isTerminal {
		// Initialize Colored Text Handler for CLI
		handler = NewColoredTextHandler(os.Stdout, &slog.HandlerOptions{
			Level:     logLevel,
			AddSource: true,
		})
	} else {
		// Initialize JSON Handler for Datadog with Log Rotation
		logFilePath := os.Getenv("LOG_FILE")
		if logFilePath == "" {
			logFilePath = "app_datadog.log"
		}

		jsonFile := &lumberjack.Logger{
			Filename:   logFilePath,
			MaxSize:    1000, // megabytes
			MaxBackups: 3,
			MaxAge:     28,   // days
			Compress:   true,
		}

		handler = slog.NewJSONHandler(jsonFile, &slog.HandlerOptions{
			Level:     logLevel,
			AddSource: true,
		})
	}

	return slog.New(handler)
}
