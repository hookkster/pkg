package sl

import (
	"log/slog"
	"os"

	"github.com/hookkster/pkg/logger/handlers/slogpretty"
)

const (
	EnvLocal = "local"
	EnvDev   = "dev"
	EnvProd  = "prod"
)

// Config describes how a service wants its logger built.
type Config struct {
	// Env selects the format. EnvLocal is the colourised handler; every other
	// value — including an unrecognised one — is JSON, because a service that
	// quietly writes ANSI escapes into a log shipper is far harder to notice
	// than one that writes JSON on your laptop.
	Env string
 
	// Service is attached to every record. Once several services write into
	// one index, this is the first thing you filter on.
	Service string
 
	// Version is attached to every record. Pass the git tag or short SHA the
	// binary was built from, so you can tell which build produced a line.
	Version string
}

func NewLogger(cfg Config) *slog.Logger {
	var handler slog.Handler
 
	if cfg.Env == EnvLocal {
		opts := slogpretty.PrettyHandlerOptions{
			SlogOpts: &slog.HandlerOptions{Level: slog.LevelDebug},
		}
		handler = opts.NewPrettyHandler(os.Stdout)
	} else {
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		})
	}
 
	log := slog.New(&contextHandler{Handler: handler})
 
	attrs := make([]any, 0, 3)
	if cfg.Service != "" {
		attrs = append(attrs, slog.String("service", cfg.Service))
	}
	if cfg.Env != "" {
		attrs = append(attrs, slog.String("env", cfg.Env))
	}
	if cfg.Version != "" {
		attrs = append(attrs, slog.String("version", cfg.Version))
	}
	if len(attrs) > 0 {
		log = log.With(attrs...)
	}
 
	return log
}

// Err renders an error as a log attribute. The nil check matters because the
// old version dereferenced a nil error and panicked inside the logger — the
// one place where a panic is hardest to diagnose.
func Err(err error) slog.Attr {
	if err == nil {
		return slog.String("error", "")
	}
 
	return slog.String("error", err.Error())
}