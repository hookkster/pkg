package sl

import (
	"log/slog"
	"os"

	"github.com/hookkster/pkg/logger/handlers/slogpretty"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

var defaultLogger *slog.Logger

func DefaultLogger() *slog.Logger {
	if defaultLogger == nil {
		env := os.Getenv("APP_ENV")
		defaultLogger = NewLogger(env)
	}
	return defaultLogger
}

func NewLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case envProd:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	case envLocal, envDev:
		log = setupPrettySlog()
	default:
		log = setupPrettySlog()
	}

	return log
}

func Err(err error) slog.Attr {
	return slog.Attr{
		Key:   "error",
		Value: slog.StringValue(err.Error()),
	}
}

func setupPrettySlog() *slog.Logger {
	opts := slogpretty.PrettyHandlerOptions{
		SlogOpts: &slog.HandlerOptions{
			Level: slog.LevelDebug,
		},
	}

	handler := opts.NewPrettyHandler(os.Stdout)

	return slog.New(handler)
}