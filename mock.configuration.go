package ulog

type MockConfiguration = func(*mockentry)

// ExpectEntry adds an expectation for a log entry and applies functions
// to refine the expectation.
//
// An expectation with no refinement will match any log entry.
func (mk *mock) ExpectEntry(fn ...MockConfiguration) {
	me := &mockentry{fields: map[string]*string{}}
	for _, f := range fn {
		f(me)
	}
	mk.cfg.expectations = append(mk.cfg.expectations, &expectation{mockentry: me, met: false})

	if mk.cfg.expected == nil {
		mk.cfg.expected = mk.cfg.expectations[0]
	}
}

// ExpectTrace adds an expectation for a log entry at TraceLevel, applying
// additional configuration as specified.
//
// ExpectTrace(fn...) is equivalent to ExpectEntry(AtLevel(TraceLevel), fn...)
func (mk *mock) ExpectTrace(fn ...MockConfiguration) {
	mk.ExpectEntry(append([]MockConfiguration{ExpectLevel(TraceLevel)}, fn...)...)
}

// ExpectDebug adds an expectation for a log entry at DebugLevel, applying
// additional configuration as specified.
//
// ExpectDebug(fn...) is equivalent to ExpectEntry(AtLevel(DebugLevel), fn...)
func (mk *mock) ExpectDebug(fn ...MockConfiguration) {
	mk.ExpectEntry(append([]MockConfiguration{ExpectLevel(DebugLevel)}, fn...)...)
}

// ExpectInfo adds an expectation for a log entry at InfoLevel, applying
// additional configuration as specified.
//
// ExpectInfo(fn...) is equivalent to ExpectEntry(AtLevel(InfoLevel), fn...)
func (mk *mock) ExpectInfo(fn ...MockConfiguration) {
	mk.ExpectEntry(append([]MockConfiguration{ExpectLevel(InfoLevel)}, fn...)...)
}

// ExpectWarn adds an expectation for a log entry at WarnLevel, applying
// additional configuration as specified.
//
// ExpectWarnfn...) is equivalent to ExpectEntry(AtLevel(WarnLevel), fn...)
func (mk *mock) ExpectWarn(fn ...MockConfiguration) {
	mk.ExpectEntry(append([]MockConfiguration{ExpectLevel(WarnLevel)}, fn...)...)
}

// ExpectError adds an expectation for a log entry at ErrorLevel, applying
// additional configuration as specified.
//
// ExpectError(fn...) is equivalent to ExpectEntry(AtLevel(ErrorLevel), fn...)
func (mk *mock) ExpectError(fn ...MockConfiguration) {
	mk.ExpectEntry(append([]MockConfiguration{ExpectLevel(ErrorLevel)}, fn...)...)
}

// ExpectFatal adds an expectation for a log entry at FatalLevel, applying
// additional configuration as specified.
//
// ExpectFatal(fn...) is equivalent to ExpectEntry(AtLevel(FatalLevel), fn...)
func (mk *mock) ExpectFatal(fn ...MockConfiguration) {
	mk.ExpectEntry(append([]MockConfiguration{ExpectLevel(FatalLevel)}, fn...)...)
}

// ExpectLevel returns a function that sets the expected log level of an entry.
func ExpectLevel(l Level) func(*mockentry) {
	return func(me *mockentry) {
		me.Level = &l
	}
}

// ExpectField returns a function that sets an expected field in an entry.
//
// The expectation will match an entry that has the specified field among its
// fields, regardless of the value of the field.
func ExpectField(s string) func(*mockentry) {
	return func(me *mockentry) {
		me.fields[s] = nil
	}
}

// ExpectFieldValue returns a function that sets an expected field and value.
//
// The expectation will match an entry if it has, among its fields, the
// specified field with matching value.
func ExpectFieldValue(s, v string) func(*mockentry) {
	return func(me *mockentry) {
		me.fields[s] = &v
	}
}

// ExpectFieldValues returns a function that sets expected fields in an entry.
//
// The expectation will match if it has fields that, at a minimum, contain
// all of the specified fields with values matching those expected.
func ExpectFieldValues(fields map[string]string) func(*mockentry) {
	return func(me *mockentry) {
		for k, v := range fields {
			v := v // shadow loop variable as we need a unique address
			me.fields[k] = &v
		}
	}
}

// ExpectMessage returns a function that sets an expected message.
//
// The expectation will only match an entry if it has the specified message.
func ExpectMessage(s string) func(*mockentry) {
	return func(me *mockentry) {
		me.string = &s
	}
}
