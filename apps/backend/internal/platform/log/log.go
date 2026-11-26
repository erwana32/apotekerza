package log

import (
	"fmt"
	"log/slog"
	"os"
)

// New returns a JSON logger writing to stdout at the given level
// (debug, info, warn, error).
func New(level string) (*slog.Logger, error) {
	var parsed slog.Level
	if err := parsed.UnmarshalText([]byte(level)); err != nil {
		return nil, fmt.Errorf("log: invalid level %q: %w", level, err)
	}
	return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: parsed})), nil
}
