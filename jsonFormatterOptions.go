package ulog

// LogfmtLabels configures the labels used for the each of the core
// fields in a logfmt log: time, level, message, file and function.
//
// A map[FieldId]string is used to override the default label for each
// field that is required; if a field is not included in the map the
// default label will continue to be used for that field.
//
// The default labels for each field are:
//
//	TimeField:     time
//	LevelField:    level
//	MessageField:  message
//	FileField:     file
//	FunctionField: function
//
// Although the label for each field may be configured, the inclusion
// of these fields in a log entry is fixed and cannot be changed, as is
// the order of the fields in the output.
func JSONFieldNames(keys map[FieldId]string) JsonFormatterOption {
	return func(lf *jsonfmt) error {
		for k, v := range keys {
			lf.keys[k] = v
		}
		return nil
	}
}

// JSONLevelLabels configures the values used for the Level field
// in json formatted log entries.
func JSONLevelLabels(levels map[Level]string) JsonFormatterOption {
	return func(jf *jsonfmt) error {
		for k, v := range levels {
			jf.levels[k] = v
		}
		return nil
	}
}
