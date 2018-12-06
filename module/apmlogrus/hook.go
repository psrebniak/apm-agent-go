package apmlogrus

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"

	"go.elastic.co/apm"
	"go.elastic.co/apm/stacktrace"
)

var (
	// DefaultLogLevels is the log levels for which errors are reported by Hook, if Hook.LogLevels is not set.
	DefaultLogLevels = []logrus.Level{
		logrus.PanicLevel,
		logrus.FatalLevel,
		logrus.ErrorLevel,
	}
)

const (
	// DefaultFatalFlushTimeout is the default value for Hook.FatalFlushTimeout.
	DefaultFatalFlushTimeout = 5 * time.Second
)

func init() {
	stacktrace.RegisterLibraryPackage("github.com/sirupsen/logrus")
}

// Hook implements logrus.Hook, reporting log records as errors
// to the APM Server.
type Hook struct {
	// Tracer is the apm.Tracer to use for reporting errors.
	// If Tracer is nil, then apm.DefaultTracer will be used.
	Tracer *apm.Tracer

	// LogLevels holds the log levels to report as errors.
	// If LogLevels is nil, then the DefaultLogLevels will
	// be used.
	LogLevels []logrus.Level

	// FatalFlushTimeout is the amount of time to wait while
	// flushing a fatal log message to the APM Server before
	// the process is exited. If this is 0, then
	// DefaultFatalFlushTimeout will be used. If the timeout
	// is a negative value, then no flushing will be performed.
	FatalFlushTimeout time.Duration
}

func (h *Hook) tracer() *apm.Tracer {
	tracer := h.Tracer
	if tracer == nil {
		tracer = apm.DefaultTracer
	}
	return tracer
}

// Levels returns h.LogLevels, satisfying the logrus.Hook interface.
func (h *Hook) Levels() []logrus.Level {
	levels := h.LogLevels
	if levels == nil {
		levels = DefaultLogLevels
	}
	return levels
}

// Fire reports the log entry as an error to the APM Server.
func (h *Hook) Fire(entry *logrus.Entry) error {
	tracer := h.tracer()
	if !tracer.Active() {
		return nil
	}

	// TODO(axw) send an exception instead or as well if the entry contains an error?
	errlog := tracer.NewErrorLog(apm.ErrorLogRecord{
		Message: entry.Message,
		Level:   entry.Level.String(),
	})
	errlog.Timestamp = entry.Time
	// TODO(axw) look for trace context in entry.Data, associate with errlog.
	errlog.SetStacktrace(1)
	errlog.Send()

	if entry.Level == logrus.FatalLevel {
		// In its default configuration, logrus will exit the process
		// following a fatal log message, so we flush the tracer.
		flushTimeout := h.FatalFlushTimeout
		if flushTimeout == 0 {
			flushTimeout = DefaultFatalFlushTimeout
		}
		if flushTimeout >= 0 {
			ctx, cancel := context.WithTimeout(context.Background(), flushTimeout)
			defer cancel()
			tracer.Flush(ctx.Done())
		}
	}
	return nil
}
