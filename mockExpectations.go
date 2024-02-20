package ulog

// EntryExpectation is a function that refines the properties of an expected log entry
type EntryExpectation = func(*MockEntry)

// ExpectEntry adds an expectation for a log entry and applies functions
// to refine the expectation.
//
// An expectation with no refinement will match any log entry.
func (mk *mock) ExpectEntry(refs ...EntryExpectation) {
	me := &MockEntry{fields: map[string]*string{}}
	for _, f := range refs {
		f(me)
	}
	mk.expecting = append(mk.expecting, me)

	if mk.expected == nil {
		mk.expected = me
	}
}

// ExpectTrace adds an expectation for a log entry at TraceLevel, applying
// additional configuration as specified.
//
// ExpectTrace(refs...) is equivalent to ExpectEntry(AtLevel(TraceLevel), refs...)
func (mk *mock) ExpectTrace(refs ...EntryExpectation) {
	mk.ExpectEntry(append([]EntryExpectation{AtLevel(TraceLevel)}, refs...)...)
}

// ExpectDebug adds an expectation for a log entry at DebugLevel, applying
// additional configuration as specified.
//
// ExpectDebug(refs...) is equivalent to ExpectEntry(AtLevel(DebugLevel), refs...)
func (mk *mock) ExpectDebug(refs ...EntryExpectation) {
	mk.ExpectEntry(append([]EntryExpectation{AtLevel(DebugLevel)}, refs...)...)
}

// ExpectInfo adds an expectation for a log entry at InfoLevel, applying
// additional configuration as specified.
//
// ExpectInfo(refs...) is equivalent to ExpectEntry(AtLevel(InfoLevel), refs...)
func (mk *mock) ExpectInfo(refs ...EntryExpectation) {
	mk.ExpectEntry(append([]EntryExpectation{AtLevel(InfoLevel)}, refs...)...)
}

// ExpectWarn adds an expectation for a log entry at WarnLevel, applying
// additional configuration as specified.
//
// ExpectWarnrefs...) is equivalent to ExpectEntry(AtLevel(WarnLevel), refs...)
func (mk *mock) ExpectWarn(refs ...EntryExpectation) {
	mk.ExpectEntry(append([]EntryExpectation{AtLevel(WarnLevel)}, refs...)...)
}

// ExpectError adds an expectation for a log entry at ErrorLevel, applying
// additional configuration as specified.
//
// ExpectError(refs...) is equivalent to ExpectEntry(AtLevel(ErrorLevel), refs...)
func (mk *mock) ExpectError(refs ...EntryExpectation) {
	mk.ExpectEntry(append([]EntryExpectation{AtLevel(ErrorLevel)}, refs...)...)
}

// ExpectFatal adds an expectation for a log entry at FatalLevel, applying
// additional configuration as specified.
//
// ExpectFatal(refs...) is equivalent to ExpectEntry(AtLevel(FatalLevel), refs...)
func (mk *mock) ExpectFatal(refs ...EntryExpectation) {
	mk.ExpectEntry(append([]EntryExpectation{AtLevel(FatalLevel)}, refs...)...)
}

// AtLevel returns a function that sets the expected log level of an entry.
func AtLevel(l Level) func(*MockEntry) {
	return func(me *MockEntry) {
		me.level = &l
	}
}

// WithField returns a function that sets an expected field in an entry.
//
// The expectation will match an entry that has the specified field among its
// fields, regardless of the value of the field.
func WithField(key string) func(*MockEntry) {
	return WithFields(key)
}

// WithFields returns a function that sets one or more expected fields in an entry.
//
// The expectation will match an entry that has the specified fields among its
// fields, regardless of the value of those fields.
func WithFields(key string, moreKeys ...string) func(*MockEntry) {
	return func(me *MockEntry) {
		me.fields[key] = nil
		for _, k := range moreKeys {
			me.fields[k] = nil
		}
	}
}

// WithFieldValue returns a function that sets an expected field and value.
//
// The expectation will match an entry if it has, among its fields, the
// specified field with matching value.
func WithFieldValue(k, v string) func(*MockEntry) {
	return func(me *MockEntry) {
		me.fields[k] = &v
	}
}

// WithFieldValues returns a function that sets expected fields in an entry.
//
// The expectation will match if it has fields that, at a minimum, contain
// all of the specified fields with values matching those expected.
func WithFieldValues(fields map[string]string) func(*MockEntry) {
	return func(me *MockEntry) {
		for k, v := range fields {
			v := v // shadow loop variable as we need a unique address
			me.fields[k] = &v
		}
	}
}

// WithMessage returns a function that sets an expected message.
//
// The expectation will only match an entry if it has the specified message.
func WithMessage(s string) func(*MockEntry) {
	return func(me *MockEntry) {
		me.string = &s
	}
}
