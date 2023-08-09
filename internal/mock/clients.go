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
