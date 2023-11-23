package ulog

import "errors"

type JsonFormatterOption func(*jsonfmt) error // JsonFormatterOption is a function for configuring a json formatter

// NewJSONFormatter returns a function that configures a json formatter.
func NewJSONFormatter(opt ...JsonFormatterOption) FormatterFactory {
	return func() (Formatter, error) {
		mf := newJsonFormatter()

		errs := []error{}
		for _, cfg := range opt {
			errs = append(errs, cfg(mf))
		}
		if err := errors.Join(errs...); err != nil {
			return nil, err
		}

		return mf, nil
	}
}

// LogfmtLabels configures the labels used for the each of the core
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
func JsonLabels(keys map[FieldId]string) JsonFormatterOption {
	return func(lf *jsonfmt) error {
		if s, ok := keys[TimeField]; ok {
			lf.keys[TimeField] = s
		}
		if s, ok := keys[LevelField]; ok {
			lf.keys[LevelField] = s
		}
		if s, ok := keys[MessageField]; ok {
			lf.keys[MessageField] = s
		}
		return nil
	}
}

// JsonLevels configures the values used for the Level field
// in json formatted log entries.
func JsonLevels(levels map[Level]string) JsonFormatterOption {
	return func(jf *jsonfmt) error {
		for k, v := range levels {
			jf.levels[k] = v
		}
		return nil
	}
}
