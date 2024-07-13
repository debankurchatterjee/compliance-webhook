package log

import (
	"context"
	"io"
	"os"
	"sync"

	"github.com/bombsimon/logrusr/v3"
	"github.com/go-logr/logr"
	"github.com/sirupsen/logrus"
)

type ctxValueType string

var timeformatdefault string
var file *os.File
var once sync.Once
var l *logrus.Logger

const (
	logfilePath              = "complaince-webhook.log"
	logKey      ctxValueType = "_logger"
)

// Sync access to logger by different go-routines
var logMutex sync.Mutex

// From extracts a logger from a context. If one does not exist, a new logger is created.
func From(ctx context.Context) logr.Logger {
	logMutex.Lock()
	defer logMutex.Unlock()
	timeformatdefault = "02-01-2006 15:04:05.0000 UTC"
	if ctx == nil {
		return newLogger(&timeformatdefault)
	}

	if logger, ok := ctx.Value(logKey).(logr.Logger); ok {
		return logger
	}

	return newLogger(&timeformatdefault)
}

// LoggerOption is an option for configuring the logger.
type LoggerOption func(config *LoggerConfig)

// LoggerOutput sets the output location for the logger.
func LoggerOutput(w io.Writer) LoggerOption {
	if w == nil {
		panic("logger output cannot be nil")
	}

	return func(config *LoggerConfig) {
		config.out = w
	}
}

// WithExistingLogger creates a new context with an existing logger.
func WithExistingLogger(ctx context.Context, logger logr.Logger) context.Context {
	return context.WithValue(ctx, logKey, logger)
}

// WithLogger creates a new context with an embedded logger.
func WithLogger(ctx context.Context, timeformat *string, options ...LoggerOption) context.Context {
	return context.WithValue(ctx, logKey, newLogger(timeformat, options...))
}

// LoggerConfig is logger configuration.NewLogger
type LoggerConfig struct {
	out io.Writer
}

func newLoggerConfig() *LoggerConfig {
	config := &LoggerConfig{
		out: os.Stderr,
	}

	return config
}

type UTCFormatter struct {
	Formatter logrus.Formatter
}

func (u UTCFormatter) Format(e *logrus.Entry) ([]byte, error) {
	e.Time = e.Time.UTC()
	return u.Formatter.Format(e)
}

func (config *LoggerConfig) update(logger *logrus.Logger, timeformat *string) {
	logger.Out = config.out
	logger.SetFormatter(UTCFormatter{&logrus.TextFormatter{TimestampFormat: *timeformat,
		FullTimestamp: true}})
}

func newLogger(timeformat *string, options ...LoggerOption) logr.Logger {
	once.Do(func() {
		currentLogFilePath := logfilePath
		config := newLoggerConfig()
		for _, option := range options {
			option(config)
		}
		l = logrus.New()
		config.update(l, timeformat)
		file, _ = os.OpenFile(currentLogFilePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0o755)
		l.SetOutput(io.MultiWriter(os.Stderr, file))
		logrus.RegisterExitHandler(func() {
			err := file.Close()
			if err != nil {
				return
			}
		})
	})
	return logrusr.New(l)
}
