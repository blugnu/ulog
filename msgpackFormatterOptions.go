package ulog

import (
	"github.com/blugnu/msgpack"
)

// MsgpackKeys configures the keys used for the each of the core
// fields in a log entry: time, level, message, file and function.
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
func MsgpackKeys(keys map[FieldId]string) MsgpackOption {
	return func(mp *msgpackfmt) error {
		for k, v := range keys {
			mp.keys[k] = msgpack.EncodeString(v)
		}
		return nil
	}
}

// MsgpackLevels configures the values used for the Level field
// in msgpack formatted log entries.
func MsgpackLevels(levels map[Level]string) MsgpackOption {
	return func(mf *msgpackfmt) error {
		for k, v := range levels {
			mf.levels[k] = msgpack.EncodeString(v)
		}
		return nil
	}
}
