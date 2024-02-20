package ulog

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/blugnu/test"
)

func TestLogtailBatchHandler(t *testing.T) {
	// ARRANGE
	sut := newLogtailBatchHandler()

	testcases := []struct {
		scenario string
		exec     func(t *testing.T)
	}{
		// configure tests
		{scenario: "configure/endpoint",
			exec: func(t *testing.T) {
				// ARRANGE

				// ACT
				err := sut.configure(logtailEndpoint, "endpoint")

				// ASSERT
				test.That(t, sut.endpoint).Equals("endpoint")
				test.Error(t, err).IsNil()
			},
		},
		{scenario: "configure/token",
			exec: func(t *testing.T) {
				// ARRANGE

				// ACT
				err := sut.configure(logtailSourceToken, "token")

				// ASSERT
				test.That(t, sut.token).Equals("token")
				test.Error(t, err).IsNil()
			},
		},
		{scenario: "configure/unknown key",
			exec: func(t *testing.T) {
				// ARRANGE

				// ACT
				err := sut.configure("unknown", "not used")

				// ASSERT
				test.Error(t, err).Is(ErrLogtailConfiguration)
				test.Error(t, err).Is(ErrKeyNotSupported)
			},
		},

		// send tests
		{scenario: "send",
			exec: func(t *testing.T) {
				// ARRANGE
				var (
					authheader  string
					contenttype string
					body        []byte
				)
				srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					authheader = r.Header.Get("Authorization")
					contenttype = r.Header.Get("Content-Type")
					body, _ = io.ReadAll(r.Body)
					defer r.Body.Close()
				}))
				sut.endpoint = srv.URL
				sut.token = "token"

				b := &Batch{
					entries: [][]byte{packedBytes(0x81, 0xa3, "key", 0xa5, "value")},
					size:    4,
					len:     1,
					max:     1,
				}

				// ACT
				_ = sut.send(b)

				// ASSERT
				test.That(t, authheader).Equals("Bearer token")
				test.That(t, contenttype).Equals("application/msgpack")
				test.That(t, body).Equals(packedBytes(0x91, 0x81, 0xa3, "key", 0xa5, "value"))
			},
		},
		{scenario: "send/encoding error",
			exec: func(t *testing.T) {
				// ARRANGE
				encerr := errors.New("encoding error")
				sut.encodeBatch = func(*bytes.Buffer, *Batch) error { return encerr }

				// ACT
				err := sut.send(&Batch{})

				// ASSERT
				test.Error(t, err).Is(encerr)
			},
		},
		{scenario: "send/invalid endpoint",
			exec: func(t *testing.T) {
				// ARRANGE
				sut.endpoint = "\n"
				b := &Batch{}

				// ACT
				err := sut.send(b)

				// ASSERT
				_, _ = test.IsType[*url.Error](t, err)
			},
		},
		{scenario: "send/invalid url scheme",
			exec: func(t *testing.T) {
				// ARRANGE
				b := &Batch{}

				// ACT
				err := sut.send(b)

				// ASSERT
				_, _ = test.IsType[*url.Error](t, err)
			},
		},
	}
	for _, tc := range testcases {
		t.Run(tc.scenario, func(t *testing.T) {
			// ARRANGE
			sut.init()

			// ACT
			tc.exec(t)
		})
	}
}
