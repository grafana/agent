package slogadapter

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
)

// GoKit returns a [log.Logger] that writes to the provided slog.Handler.
func GoKit(h slog.Handler) log.Logger {
	return slogAdapter{h}
}

type slogAdapter struct {
	h slog.Handler
}

func (sa slogAdapter) Log(kvps ...interface{}) error {
	// We don't know what the log level or message are yet, so we set them to
	// defaults until we iterate through kvps.
	//
	// Since there's a pattern of wrapping loggers at different depths in go-kit,
	// we can't consistently know who the caller is, so we set no value for pc
	// here.
	rec := slog.NewRecord(time.Now(), slog.LevelInfo, "", 0)

	for i := 0; i < len(kvps); i += 2 {
		var key, value any

		if i+1 < len(kvps) {
			key = kvps[i+0]
			value = kvps[i+1]
		} else {
			// Mismatched pair
			key = "!BADKEY"
			value = kvps[i+0]
		}

		if key == "msg" || key == "message" {
			rec.Message = fmt.Sprint(value)
			continue
		}

		if key == level.Key() {
			levelValue, _ := value.(level.Value)

			switch levelValue {
			case level.DebugValue():
				rec.Level = slog.LevelDebug
			case level.InfoValue():
				rec.Level = slog.LevelInfo
			case level.WarnValue():
				rec.Level = slog.LevelWarn
			case level.ErrorValue():
				rec.Level = slog.LevelError
			}

			continue
		}

		rec.AddAttrs(slog.Attr{
			Key:   fmt.Sprint(key),
			Value: slog.AnyValue(value),
		})
	}

	if !sa.h.Enabled(context.Background(), rec.Level) {
		return nil
	}
	return sa.h.Handle(context.Background(), rec)
}
