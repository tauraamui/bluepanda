// Copyright (c) 2023 Adam Prakash Stringer
// All rights reserved.
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted (subject to the limitations in the disclaimer
// below) provided that the following conditions are met:
//
//     * Redistributions of source code must retain the above copyright notice,
//     this list of conditions and the following disclaimer.
//
//     * Redistributions in binary form must reproduce the above copyright
//     notice, this list of conditions and the following disclaimer in the
//     documentation and/or other materials provided with the distribution.
//
//     * Neither the name of the copyright holder nor the names of its
//     contributors may be used to endorse or promote products derived from this
//     software without specific prior written permission.
//
// NO EXPRESS OR IMPLIED LICENSES TO ANY PARTY'S PATENT RIGHTS ARE GRANTED BY
// THIS LICENSE. THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND
// CONTRIBUTORS "AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT
// LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A
// PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR
// CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL,
// EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO,
// PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR
// BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER
// IN CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE)
// ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE
// POSSIBILITY OF SUCH DAMAGE.

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
