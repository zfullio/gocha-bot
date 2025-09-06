package logger

import (
	"io"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/pkgerrors"
)

const (
	DefaultTimestampFieldName   = "time"
	DefaultLevelFieldName       = "level"
	DefaultMessageFieldName     = "message"
	DefaultErrorStackFieldName  = "stacktrace"
	DefaultCallerSkipFrameCount = 2
	componentField              = "component"
	Debug                       = "debug"
	debugMessage                = "debug log is enabled"
)

func init() {
	// UNIX Time is faster and smaller than most timestamps
	// If you set zerolog.TimeFieldFormat to an empty string,
	// logs will write with UNIX time
	// zerolog.TimeFieldFormat = zerolog.TimeFormatUnixMicro
	// zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	zerolog.TimeFieldFormat = time.RFC3339Nano
	zerolog.TimestampFieldName = DefaultTimestampFieldName
	zerolog.LevelFieldName = DefaultLevelFieldName
	zerolog.MessageFieldName = DefaultMessageFieldName
	zerolog.ErrorStackFieldName = DefaultErrorStackFieldName
	zerolog.CallerSkipFrameCount = DefaultCallerSkipFrameCount
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack
}

type myWriteCloser struct {
	io.Writer
}

func (mwc *myWriteCloser) Close() error {
	// Noop
	return nil
}

func NewLogger(w io.Writer, logLevel string) (zerolog.Logger, io.WriteCloser, error) {
	lvl, err := zerolog.ParseLevel(logLevel)
	if err != nil {
		return zerolog.Logger{}, nil, err
	}

	wc, ok := w.(io.WriteCloser)
	if !ok {
		wc = &myWriteCloser{w}
	}

	logger := zerolog.New(wc).Level(lvl).With().Timestamp().Logger()
	if lvl == zerolog.DebugLevel {
		logger.Debug().Bool(Debug, true).Msg(debugMessage)
	}

	return logger, wc, nil
}

func NewComponentLogger(logger zerolog.Logger, component string, skipFrameCount int) zerolog.Logger {
	return logger.With().Str(componentField, component).CallerWithSkipFrameCount(skipFrameCount).Logger()
}
