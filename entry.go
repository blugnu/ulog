package ulog

import (
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
