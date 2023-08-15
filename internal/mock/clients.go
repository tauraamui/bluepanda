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

package mock

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/rs/zerolog"
)

type LogWriter struct {
	mu           sync.Mutex
	logsPerLevel map[zerolog.Level][]string
}

func (m *LogWriter) Write(p []byte) (n int, err error) { return 0, nil }

func (m *LogWriter) WriteLevel(level zerolog.Level, p []byte) (n int, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.logsPerLevel == nil {
		m.logsPerLevel = map[zerolog.Level][]string{}
	}
	m.logsPerLevel[level] = append(m.logsPerLevel[level], string(p))
	return 0, nil
}

func (m *LogWriter) InfoLogs() []string {
	return m.logsPerLevel[zerolog.InfoLevel]
}

func (m *LogWriter) DebugLogs() []string {
	return m.logsPerLevel[zerolog.DebugLevel]
}

func (m *LogWriter) WarnLogs() []string {
	return m.logsPerLevel[zerolog.WarnLevel]
}

func (m *LogWriter) ErrorLogs() []string {
	return m.logsPerLevel[zerolog.ErrorLevel]
}

func (m *LogWriter) ErrorLogsErrorFieldsOnly() []string {
	errorLogErrs := []string{}
	dst := struct {
		Error string `json:"error"`
	}{}

	for _, el := range m.ErrorLogs() {
		if err := json.Unmarshal([]byte(el), &dst); err != nil {
			continue
		}
		errorLogErrs = append(errorLogErrs, dst.Error)
	}

	return errorLogErrs
}

func (m *LogWriter) ErrorLogsErrorFieldNLDelimStr() string {
	buf := strings.Builder{}
	for _, el := range m.ErrorLogsErrorFieldsOnly() {
		buf.WriteString(el)
		buf.WriteString("\n")
	}
	return buf.String()
}

func (m *LogWriter) MatchInfoLogs(expInfos []string) error {
	return resolveLogs(zerolog.InfoLevel, expInfos, m.InfoLogs())
}

func (m *LogWriter) MatchErrorLogs(expErrs []string) error {
	return resolveLogs(zerolog.ErrorLevel, expErrs, m.ErrorLogs())
}

func (m *LogWriter) MatchDebugLogs(expDebugs []string) error {
	return resolveLogs(zerolog.DebugLevel, expDebugs, m.DebugLogs())
}

func resolveLogs(level zerolog.Level, exp, act []string) error {
	sb := strings.Builder{}

	renderNonExistentLogs(level, &sb, exp, act)
	renderUnexpectedLogs(level, &sb, exp, act)

	failureMsg := sb.String()
	if len(failureMsg) > 0 {
		return errors.New(failureMsg)
	}
	return nil
}

func renderNonExistentLogs(level zerolog.Level, sb *strings.Builder, exp, act []string) {
	if str := renderLogs(fmt.Sprintf("non-existent expected %s(s)", level), exp, act); len(str) > 0 {
		sb.WriteString(str)
	}
}

func renderUnexpectedLogs(level zerolog.Level, sb *strings.Builder, exp, act []string) {
	if str := renderLogs(fmt.Sprintf("unexpected %s(s)", level), act, exp); len(str) > 0 {
		sb.WriteString(str)
	}
}

func renderLogs(header string, exp, act []string) string {
	sb := strings.Builder{}
	errs := missingLogs(exp, act)
	if c := len(errs); c > 0 {
		sb.WriteString(fmt.Sprintf("%d %s:\n", c, header))

		for _, ee := range errs {
			sb.WriteString(fmt.Sprintf("\t%s\n", strings.TrimSuffix(ee, "\n")))
		}
	}
	return sb.String()
}

func missingLogs(exp, act []string) []string {
	missing := []string{}
	for _, ee := range exp {
		foundMatch := false
		for _, ae := range act {
			if strings.TrimSuffix(ee, "\n") == strings.TrimSuffix(ae, "\n") {
				foundMatch = true
				break
			}
		}

		if !foundMatch {
			missing = append(missing, ee)
		}
	}
	return missing
}
