package ulog

import "fmt"

// Level identifies the logging level for a particular log entry. Possible values,
// in increasing order of severity (decreasing ordinal value), are:
//   - TraceLevel (6)
//   - DebugLevel
//   - InfoLevel
//   - WarnLevel
//   - ErrorLevel
//   - FatalLevel (1)
type Level int

var Levels = []Level{TraceLevel, DebugLevel, InfoLevel, WarnLevel, ErrorLevel, FatalLevel}

const (
	levelNotSet Level = iota
	FatalLevel        // FatalLevel indicates a condition that requires the process to terminate; after writing the log ulog.ExitFn is called
	ErrorLevel        // ErrorLevel indicates the log entry relates to an error
	WarnLevel         // WarnLevel indicates the log entry is a warning about some unexpected but recoverable condition
	InfoLevel         // InfoLevel indicates the log entry is an informational message
	DebugLevel        // DebugLevel indicates the log entry is a debug message, typically used during development
	TraceLevel        // TraceLevel indicates the log entry is a trace message, usually to aid in diagnostics

	numLevels = 7 // the number of defined log levels: 7 => 6 levels (trace, debug, info, warn, error, fatal) + 1 not set
)

// String implements the Stringer interface for Level
func (lv Level) String() string {
	s := map[Level]string{
		levelNotSet: "<not set>",
		FatalLevel:  "FATAL",
		ErrorLevel:  "ERROR",
		WarnLevel:   "warn",
		InfoLevel:   "info",
		DebugLevel:  "debug",
		TraceLevel:  "trace",
	}
	if str, ok := s[lv]; ok {
		return str
	}
	return fmt.Sprintf("<invalid level (%d)>", lv)
}
