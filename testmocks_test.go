package ulog

import "io"

type mockbackend struct {
	dispatchfn  func(entry)
	setformatfn func(Formatter) error
	setoutputfn func(io.Writer) error
	startfn     func() (func(), error)
}

func (b *mockbackend) dispatch(e entry) {
	b.dispatchfn(e)
}

func (b *mockbackend) SetFormatter(f Formatter) error {
	return b.setformatfn(f)
}

func (b *mockbackend) SetOutput(w io.Writer) error {
	return b.setoutputfn(w)
}

func (b *mockbackend) start() (func(), error) {
	return b.startfn()
}

type mockBatchHandler struct {
	sendCalls   int
	sentEntries int
	sentBytes   int
	sendfn      func(*Batch) error
}

func (m *mockBatchHandler) configure(key cfgkey, value any) error { return nil }
func (m *mockBatchHandler) send(batch *Batch) error {
	m.sendCalls++

	fn := func(*Batch) error { return nil }
	if m.sendfn != nil {
		fn = m.sendfn
	}

	if err := fn(batch); err != nil {
		return err
	}
	m.sentEntries += batch.len
	m.sentBytes += batch.size
	return nil
}
func (m *mockBatchHandler) reset() { m.sentBytes = 0; m.sentEntries = 0; m.sendCalls = 0 }

type mocktransport struct {
	logWasCalled  bool
	runWasCalled  bool
	stopWasCalled bool
}

func (m *mocktransport) run() {
	m.runWasCalled = true
}

func (m *mocktransport) stop() {
	m.stopWasCalled = true
}

func (m *mocktransport) log([]byte) {
	m.logWasCalled = true
}

type mockmutex struct {
	lockWasCalled   bool
	unlockWasCalled bool
}

func (m *mockmutex) Lock() {
	m.lockWasCalled = true
}

func (m *mockmutex) Unlock() {
	m.unlockWasCalled = true
}

func (m *mockmutex) Reset() {
	m.lockWasCalled = false
	m.unlockWasCalled = false
}

type mockformatter struct {
	formatWasCalled bool
	formatfn        func(int, entry, ByteWriter)
}

func (mock *mockformatter) Format(i int, e entry, w ByteWriter) {
	mock.formatWasCalled = true
	if mock.formatfn != nil {
		mock.formatfn(i, e, w)
		return
	}
	_, _ = w.Write([]byte(e.Message))
}

type mockdispatcher struct {
	entry entry
}

func (md *mockdispatcher) dispatch(e entry) {
	md.entry = e
}

func (md *mockdispatcher) Reset() {
	md.entry = noop.entry
}
