package ulog

import (
	"fmt"
	"strings"
)

// LogfmtFieldNames configures the labels used for the each of the core
// fields in a logfmt log: time, level and message.
//
// A map[FieldId]string is used to override the default label for each
// field that is required; if a field is not included in the map, the
// default label will continue to be used for that field.
//
// The default labels for each field are:
//
//	TimeField:     time
//	LevelField:    level
//	MessageField:  message
//
// Although the label for each field may be configured, the inclusion
// of these fields and their order is fixed, and cannot be changed.
func LogfmtFieldNames(keys map[FieldId]string) LogfmtOption {
	return func(lf *logfmt) error {
		if s, ok := keys[TimeField]; ok {
			lf.keys[TimeField] = []byte(fmt.Sprintf("%s=", s))
		}
		if s, ok := keys[LevelField]; ok {
			lf.keys[LevelField] = []byte(fmt.Sprintf(" %s=", s))
		}
		if s, ok := keys[MessageField]; ok {
			lf.keys[MessageField] = []byte(fmt.Sprintf(" %s=\"", s))
		}
		return nil
	}
}

// LogfmtLevelLabels may be used to override the values used for the Level
// field in logfmt log entries.
//
// A map[Level]string is used to override the default value for each level
// that is required; for any Level not included in the map, the currently
// configured value will be left as-is.
//
// The default labels are:
//
//	TraceLevel: TRACE
//	DebugLevel: DEBUG
//	InfoLevel:  INFO
//	WarnLevel:  WARN
//	ErrorLevel: ERROR
//	FatalLevel: FATAL
//
// Values are automatically right-padded with spaces to be of equal length
// to make it easier to visually parse log entries when reading a log,
// ensuring that the message part of each entry starts at the same position.
func LogfmtLevelLabels(levels map[Level]string) LogfmtOption {
	return func(lf *logfmt) error {
		w := 0
		for k, v := range levels {
			lf.levels[k] = []byte(v)
			if len(v) > w {
				w = len(v)
			}
		}

		// rpad with spaces to ensure all labels are the same length
		for i := 1; i < numLevels; i++ {
			v := lf.levels[i]
			if len(v) < w {
				lf.levels[i] = append(v, []byte(strings.Repeat(" ", w-len(v)))...)
			}
		}

		return nil
	}
}
