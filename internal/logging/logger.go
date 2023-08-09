package logging

import (
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
)

type Logger interface {
	I() zerolog.Logger
	Debug() Event
	Info() Event
	Warn() Event
	Error() Event
	Fatal() Event
	Panic() Event
	With() Context
	WithCopy() Context
}

type logger struct {
	levelWriter io.Writer
	rootZero    zerolog.Logger
}

func (l *logger) I() zerolog.Logger {
	return l.rootZero
}

func (l *logger) Debug() Event {
	return &event{l.rootZero.Debug()}
}

func (l *logger) Info() Event {
	return &event{l.rootZero.Info()}
}

func (l *logger) Warn() Event {
	return &event{l.rootZero.Warn()}
}

func (l *logger) Error() Event {
	return &event{l.rootZero.Error()}
}

func (l *logger) Fatal() Event {
	return &event{l.rootZero.Fatal()}
}

func (l *logger) Panic() Event {
	return &event{l.rootZero.Panic()}
}

func (l *logger) With() Context {
	return &context{l}
}

func (l *logger) WithCopy() Context {
	return &context{
		&logger{
			levelWriter: l.levelWriter,
			rootZero:    l.rootZero.With().Logger(),
		},
	}
}

type Context interface {
	Timestamp() Context
	Str(string, string) Context
	Logger() Logger
}

type context struct {
	rootLogger *logger
}

func (c *context) Timestamp() Context {
	c.rootLogger.rootZero = c.rootLogger.rootZero.With().Timestamp().Logger()
	return c
}

func (c *context) Str(key, val string) Context {
	c.rootLogger.rootZero = c.rootLogger.rootZero.With().Str(key, val).Logger()
	return c
}

func (c *context) Logger() Logger {
	return c.rootLogger
}

type Event interface {
	Msg(string)
	Msgf(string, ...interface{})
	Err(error) Event
	Fields(interface{}) Event
	Dict(string, Event) Event
	Str(string, string) Event
	Strs(string, []string) Event
	Int(string, int) Event
	Bool(string, bool) Event
	Timestamp() Event
	TimeDiff(key string, t time.Time, start time.Time) Event
	root() *zerolog.Event
}

type event struct {
	rootEvent *zerolog.Event
}

func (e *event) Err(err error) Event {
	e.rootEvent.Err(err)
	return e
}

func (e *event) Msg(msg string) {
	e.rootEvent.Msg(msg)
}

func (e *event) Msgf(format string, v ...interface{}) {
	e.rootEvent.Msgf(format, v...)
}

func (e *event) Fields(fields interface{}) Event {
	e.rootEvent.Fields(fields)
	return e
}

func (e *event) Dict(key string, dict Event) Event {
	e.rootEvent.Dict(key, dict.root())
	return e
}

func (e *event) Str(key, val string) Event {
	e.rootEvent.Str(key, val)
	return e
}

func (e *event) Strs(key string, vals []string) Event {
	e.rootEvent.Strs(key, vals)
	return e
}

func (e *event) Int(key string, i int) Event {
	e.rootEvent.Int(key, i)
	return e
}

func (e *event) Bool(key string, b bool) Event {
	e.rootEvent.Bool(key, b)
	return e
}

func (e *event) Timestamp() Event {
	e.rootEvent.Timestamp()
	return e
}

func (e *event) TimeDiff(key string, t time.Time, start time.Time) Event {
	e.rootEvent.TimeDiff(key, t, start)
	return e
}

func (e *event) root() *zerolog.Event {
	return e.rootEvent
}

func New(r ...zerolog.LevelWriter) Logger {
	var logWriter io.Writer = os.Stdout
	if len(r) > 0 {
		logWriter = r[0]
	}

	return &logger{
		levelWriter: logWriter,
		rootZero:    zerolog.New(logWriter),
	}
}
