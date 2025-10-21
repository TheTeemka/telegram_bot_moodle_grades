package logging

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
)

func SetSlog(debug bool) {
	l := slog.LevelInfo
	if debug {
		l = slog.LevelDebug
	}

	h := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level:     l,
		AddSource: true,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key != slog.SourceKey {
				return a
			}

			switch v := a.Value.Any().(type) {
			case *slog.Source:
				if v != nil {
					short := filepath.Base(v.File)
					a.Value = slog.StringValue(fmt.Sprintf("%s:%d", short, v.Line))
				}
			case slog.Source:
				short := filepath.Base(v.File)
				a.Value = slog.StringValue(fmt.Sprintf("%s:%d", short, v.Line))
			default:
				// Fallback: shorten the string representation
				s := a.Value.String()
				if s != "" {
					a.Value = slog.StringValue(filepath.Base(s))
				}
			}
			return a
		},
	})

	slog.SetDefault(slog.New(h))
}
