package ulog

import (
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/blugnu/test"
)

func TestLogtailTransport(t *testing.T) {
	// ARRANGE
	testcases := []struct {
		scenario string
		exec     func(t *testing.T)
	}{
		// LogtailTransport tests
		{scenario: "LogtailTransport/option error",
			exec: func(t *testing.T) {
				// ARRANGE
				opterr := errors.New("option error")
				opt := func(*logtail) error { return opterr }

				// ACT
				result, err := LogtailTransport(opt)()

				// ASSERT
				test.That(t, result).IsNil()
				test.Error(t, err).Is(opterr)
			},
		},
		{scenario: "LogtailTransport/with no options",
			exec: func(t *testing.T) {
				// ACT
				result, err := LogtailTransport()()

				// ASSERT
				test.That(t, result).IsNotNil()
				test.That(t, err).IsNil()

				if result, ok := test.IsType[*logtail](t, result); ok {
					batch := result.batch

					test.That(t, batch, "batch").IsNotNil()
					test.That(t, batch.max, "batch capacity").Equals(16)

					if handler, ok := test.IsType[*logtailBatchHandler](t, batch.batchHandler); ok {
						test.That(t, handler.endpoint, "endpoint").Equals("https://in.logs.betterstack.com")
					}

					test.That(t, result.maxLatency, "max latency").Equals(10 * time.Second)
				}
			},
		},

		{scenario: "log",
			exec: func(t *testing.T) {
				// ARRANGE
				sut := &logtail{ch: make(chan []byte)}

				// setup a coroutine to listen to the channel used by the transport
				// when sending log entries, copying the sent bytes
				sent := []byte{}
				wg := &sync.WaitGroup{}
				wg.Add(1)
				go func() {
					defer wg.Done()
					sent = append(sent, <-sut.ch...)
				}()

				// ACT
				sut.log([]byte("bytes sent"))

				// CLEANUP
				close(sut.ch)
				wg.Wait()

				// ASSERT
				test.That(t, sent).Equals([]byte("bytes sent"))
			},
		},

		// run scenarios
		{scenario: "run",
			exec: func(t *testing.T) {
				// ARRANGE
				mh := &mockBatchHandler{}
				sut := &logtail{
					ch:         make(chan []byte),
					batch:      &Batch{},
					maxLatency: 50 * time.Millisecond,
				}

				testcases := []struct {
					scenario string
					exec     func(t *testing.T)
				}{
					{scenario: "adds entries to batch and flushes when batch full or channel closed",
						exec: func(t *testing.T) {
							// ARRANGE
							wg := &sync.WaitGroup{}
							wg.Add(1)
							go func() {
								defer wg.Done()
								sut.run()
							}()

							// ACT
							sut.ch <- []byte("entry 1")
							sut.ch <- []byte("entry 2")
							sut.ch <- []byte("entry 3") // <- fills the first batch
							sut.ch <- []byte("entry 4") // <- incomplete batch will be sent when the channel is closed

							// CLEANUP
							close(sut.ch)
							wg.Wait()

							// ASSERT
							test.That(t, mh.sendCalls).Equals(2)
							test.That(t, mh.sentEntries).Equals(4)
						},
					},
					{scenario: "sends partial batch when max latency exceeded",
						exec: func(t *testing.T) {
							// ARRANGE
							wg := &sync.WaitGroup{}
							wg.Add(1)
							go func() {
								defer wg.Done()
								sut.run()
							}()

							// ACT
							sut.ch <- []byte("entry 1")
							sut.ch <- []byte("entry 2")
							time.Sleep(sut.maxLatency * 2)

							// CLEANUP
							close(sut.ch)
							wg.Wait()

							// ASSERT
							test.That(t, mh.sendCalls).Equals(1)
							test.That(t, mh.sentEntries).Equals(2)
						},
					},
				}
				for _, tc := range testcases {
					t.Run(tc.scenario, func(t *testing.T) {
						// ARRANGE
						mh.reset()
						sut.batch.init(mh, 3)
						sut.ch = make(chan []byte)

						// ACT
						tc.exec(t)
					})
				}
			},
		},

		// stop
		{scenario: "stop",
			exec: func(t *testing.T) {
				// ARRANGE
				sut := &logtail{ch: make(chan []byte)}
				wg := &sync.WaitGroup{}
				wg.Add(1)
				go func() {
					defer wg.Done()
					<-sut.ch // will not receive anything; will yield nil when the channel is closed
				}()

				// ACT
				sut.stop()

				// ACT / ASSERT
				// (nothing to assert; the test will timeout if the channel is not closed)
				wg.Wait()
			},
		},
	}
	for _, tc := range testcases {
		t.Run(tc.scenario, func(t *testing.T) {
			tc.exec(t)
		})
	}
}
