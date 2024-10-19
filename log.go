package openairt

import "log"

type Logger interface {
	Errorf(format string, v ...any)
	Warnf(format string, v ...any)
}

// NopLogger is a logger that does nothing.
type NopLogger struct{}

// Errorf does nothing.
func (l NopLogger) Errorf(format string, v ...any) {}

// Warnf does nothing.
func (l NopLogger) Warnf(format string, v ...any) {}

type StdLogger struct{}

func (l StdLogger) Errorf(format string, v ...any) {
	log.Printf("[ERROR] "+format, v...)
}

func (l StdLogger) Warnf(format string, v ...any) {
	log.Printf("[WARN] "+format, v...)
}
