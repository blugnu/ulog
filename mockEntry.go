package ulog

import (
	"fmt"
)

// MockEntry represents a log MockEntry.  It may represent an expected MockEntry or an
// actual one; for an expected entry, the actual field will reference the corresponding
// entry that was dispatched (if any).
//
// For an expected MockEntry, any field that has a nil expected value is deemed to
// match as long as that field is present in the target fields (regardless of value),
// otherwise the value of that field in the target must match the expected field value.
type MockEntry struct {
	level *Level
	*string
	fields map[string]*string
	actual *entry
	ok     bool
}

// matches returns true if the receiver matches the specified entry.
func (me *MockEntry) matches(target entry) bool {
	switch {
	//if set, the expected level must match the target level.
	case me.level != nil && *me.level != target.Level:
		return false

	// if set, the expected message must match the target message.
	case me.string != nil && *me.string != target.Message:
		return false

	// if set, there cannot be more fields expected than are present in the target
	case len(me.fields) > 0 && (target.fields == nil || len(me.fields) > len(target.fields.m)):
		return false

	// any field that is set must be present in the target, and if the value is also set
	// then the value of the corresponding field in the target must also match
	default:
		for k, ev := range me.fields {
			if tv, present := target.fields.m[k]; !present || (ev != nil && tv != *ev) {
				return false
			}
		}
	}
	return true
}

// String returns a string representation of the entry.
func (me *MockEntry) String() string {
	sl := "<any>"
	ss := "<any>"
	sf := "<any>"

	if me.level != nil {
		sl = me.level.String()
	}
	if me.string != nil {
		ss = fmt.Sprintf("%q", *me.string)
	}

	if me.fields != nil {
		sf = "[<none>]"
		if len(me.fields) > 0 {
			sf = ""
			for k, v := range me.fields {
				if v == nil {
					sf += fmt.Sprintf("\"%s\"=<any> ", k)
					continue
				}
				sf += fmt.Sprintf("\"%s\"=%q ", k, *v)
			}
			sf = "[" + sf[:len(sf)-1] + "]"
		}
	}

	return fmt.Sprintf("level=%s message=%s fields=%s", sl, ss, sf)
}
