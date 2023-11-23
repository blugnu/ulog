package ulog

import (
	"fmt"
)

// mockentry represents a log mockentry.  It may represent an expected mockentry or an
// actual one.  For an expected mockentry, any field that is not set is deemed to
// represent a wildcard match.
type mockentry struct {
	*Level
	*string
	fields map[string]*string
}

// matches returns true if the receiver matches the specified entry.
func (me *mockentry) matches(target *mockentry) bool {
	if (me.Level != nil && *me.Level != *target.Level) ||
		(me.string != nil && *me.string != *target.string) ||
		(me.fields != nil && len(me.fields) > len(target.fields)) {
		return false
	}

	for k, ev := range me.fields {
		if tv, ok := target.fields[k]; !ok || (ev != nil && *tv != *ev) {
			return false
		}
	}

	return true
}

// String returns a string representation of the entry.
func (me *mockentry) String() string {
	sl := "<any>"
	ss := "<any>"
	sf := "<any>"

	if me.Level != nil {
		sl = me.Level.String()
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
