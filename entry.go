package ulog

import (
	"fmt"
	"time"
)

// entry is a log entry.
type entry struct {
	*logcontext        // context of the log entry, including fields
	*callsite          // if call-site logging is enabled, callsite is the first non-ulog runtime Frame in the call stack that created the context
	noop        bool   // true if the entry is a noop
	time.Time          // time of the log entry
	Level              // level of the log entry
	Message     string // message of the log entry
}

// String returns a string representation of the entry.
func (e entry) String() string {
	sf := "[<none>]"
	if e.fields != nil {
		if len(e.fields.m) > 0 {
			sf = ""
			for k, v := range e.fields.m {
				sf += fmt.Sprintf("\"%s\"=%q ", k, v)
			}
			sf = "[" + sf[:len(sf)-1] + "]"
		}
	}

	return fmt.Sprintf("level=%s message=%q fields=%s", e.Level, e.Message, sf)
}
