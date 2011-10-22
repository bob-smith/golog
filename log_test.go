package logging

import (
	"os"
	"strings"
	"testing"
)

// Note: the below is deliberately PLACED AT THE TOP OF THIS FILE because
// it is fragile. It ensures the right file:line is logged. Sorry!
func TestLogCorrectLineNumbers(t *testing.T) {
	l, m := SetUp()
	l.Log(Error, "Error!")
	if s := string(m[Error].written); s[20:] != "log_test.go:13: ERROR Error!\n" {
		t.Errorf("Error incorrectly logged (check line numbers!)")
	}
}

type writerMap map[LogLevel]*mockWriter

type mockWriter struct {
	written []byte
}

func (w *mockWriter) Write(p []byte) (n int, err os.Error) {
	w.written = append(w.written, p...)
	return len(p), nil
}

func (w *mockWriter) reset() {
	w.written = w.written[:0]
}

// Make these accessible for other packages doing logging testing
func SetUp() (*logger, writerMap) {
	wMap := writerMap{
		Debug: &mockWriter{make([]byte, 0)},
		Info:  &mockWriter{make([]byte, 0)},
		Warn:  &mockWriter{make([]byte, 0)},
		Error: &mockWriter{make([]byte, 0)},
		Fatal: &mockWriter{make([]byte, 0)},
	}
	logMap := make(LogMap)
	for lv, w := range wMap {
		logMap[lv] = makeLogger(w)
	}
	return New(logMap, Error, false), wMap
}

func (m writerMap) CheckNothingWritten(t *testing.T) {
	for lv, w := range m {
		if len(w.written) > 0 {
			t.Errorf("%d bytes logged at level %s, expected none:",
				len(w.written), logString[lv])
			t.Errorf("\t%s", w.written)
			w.reset()
		}
	}
}

func (m writerMap) CheckWrittenAtLevel(t *testing.T, lv LogLevel, exp string) {
	var w *mockWriter
	if _, ok := m[lv]; !ok {
		w = m[Debug]
	} else {
		w = m[lv]
	}
	if len(w.written) == 0 {
		t.Errorf("No bytes logged at level %s, expected:", LogString(lv))
		t.Errorf("\t%s", exp)
	}
	// 32 bytes covers the date, time and filename up to the colon in
	// 2011/10/22 10:22:57 log_test.go:<line no>: <level> <log message>
	s := string(w.written[32:])
	// 2 covers the : itself and the extra space
	idx := strings.Index(s, ":") + 2
	// s will end in "\n", so -1 to chop that off too
	s = s[idx:len(s)-1]
	// expected won't contain the log level prefix, so prepend that
	exp = LogString(lv) + " " + exp
	if s != exp {
		t.Errorf("Log message at level %s differed.", LogString(lv))
		t.Errorf("\texp: %s\n\tgot: %s", exp, s)
	}
	w.reset()
	// Calling checkNothingWritten here both tests that w.reset() works
	// and verifies that nothing was written at any other levels.
	m.CheckNothingWritten(t)
}

func TestStandardLogging(t *testing.T) {
	l, m := SetUp()

	l.Log(4, "Nothing should be logged yet")
	m.CheckNothingWritten(t)

	l.Log(Debug, "or yet...")
	m.CheckNothingWritten(t)

	l.Log(Info, "or yet...")
	m.CheckNothingWritten(t)

	l.Log(Warn, "or yet!")
	m.CheckNothingWritten(t)

	l.Log(Error, "Error!")
	m.CheckWrittenAtLevel(t, Error, "Error!")
}

func TestAllLoggingLevels(t *testing.T) {
	l, m := SetUp()
	l.SetLogLevel(5)

	l.Log(4, "Log to level 4.")
	m.CheckWrittenAtLevel(t, 4, "Log to level 4.")

	l.Debug("Log to debug.")
	m.CheckWrittenAtLevel(t, Debug, "Log to debug.")

	l.Info("Log to info.")
	m.CheckWrittenAtLevel(t, Info, "Log to info.")

	l.Warn("Log to warning.")
	m.CheckWrittenAtLevel(t, Warn, "Log to warning.")

	l.Error("Log to error.")
	m.CheckWrittenAtLevel(t, Error, "Log to error.")

	// recover to track the panic caused by Fatal.
	defer func() { recover() }()
	l.Fatal("Log to fatal.")
	m.CheckWrittenAtLevel(t, Fatal, "Log to fatal.")
}
