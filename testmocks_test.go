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

func (b *mockbackend) setFormatter(f Formatter) error {
	return b.setformatfn(f)
}

func (b *mockbackend) setOutput(w io.Writer) error {
	return b.setoutputfn(w)
}

func (b *mockbackend) start() (func(), error) {
	return b.startfn()
}

// type mockBatchHandler struct {
// 	sent   int
// 	sendfn func(*Batch) (int, error)
// }

// func (m *mockBatchHandler) configure(key cfgkey, value any) error { return nil }
// func (m *mockBatchHandler) send(batch *Batch) (int, error) {
// 	if m.sendfn != nil {
// 		return m.sendfn(batch)
// 	}
// 	m.sent += batch.len
// 	return 0, nil
// }
// func (m *mockBatchHandler) reset() { m.sent = 0 }

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

type mockpool[T any] struct {
	getWasCalled bool
	putWasCalled bool
}

func (mock *mockpool[T]) Get() any {
	mock.getWasCalled = true
	return new(T)
}

func (mock *mockpool[T]) Put(any) {
	mock.putWasCalled = true
}

func (mock *mockpool[T]) GetWasCalled() bool {
	return mock.getWasCalled
}

func (mock *mockpool[T]) PutWasCalled() bool {
	return mock.putWasCalled
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
