package ulog

// levelLogger provides a logger which is limited to emitting
// log messages using Log() or Logf() function which emit logs
// at the specified level.
type levelLogger struct {
	*logcontext
	Level
}

// Log emits a log message at the level of the logger.
func (lv *levelLogger) Log(s string) {
	lv.log(lv.logcontext.makeEntry(lv.Level, s))
}

// Logf emits a log message at the level of the logger.  The message is
// formatted using the specified format string and args.
func (lv *levelLogger) Logf(format string, args ...any) {
	lv.log(lv.logcontext.makeEntryf(lv.Level, format, args...))
}

// WithField returns a new LevelLogger with an additional field.
func (lv *levelLogger) WithField(name string, value any) LevelLogger {
	logger := lv.logcontext.new(lv.ctx, lv.dispatcher, map[string]any{name: value}, lv.exitCode)
	return &levelLogger{logger, lv.Level}
}

// WithFields returns a new LevelLogger with additional fields.
func (lv *levelLogger) WithFields(fields map[string]any) LevelLogger {
	logger := lv.logcontext.new(lv.ctx, lv.dispatcher, fields, lv.exitCode)
	return &levelLogger{logger, lv.Level}
}
