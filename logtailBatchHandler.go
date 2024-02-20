package ulog

import (
	"bytes"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/blugnu/msgpack"
)

const (
	logtailEndpoint    = cfgkey("logtail.endpoint")
	logtailMaxLatency  = cfgkey("logtail.maxLatency")
	logtailSourceToken = cfgkey("logtail.sourceToken")
)

type logtailBatchHandler struct {
	endpoint    string
	token       string
	buf         *sync.Pool
	encodeBatch func(*bytes.Buffer, *Batch) error
}

// newLogtailBatchHandler initialises a logtail batch handler with a
// buffer pool and batch encoding function.
func (h *logtailBatchHandler) init() {
	*h = logtailBatchHandler{
		buf: &sync.Pool{New: func() any { return &bytes.Buffer{} }},
		encodeBatch: func(buf *bytes.Buffer, batch *Batch) error {
			enc, _ := msgpack.NewEncoder(buf)
			return msgpack.EncodeArray(*enc, batch.entries, func(msg msgpack.Encoder, e []byte) error {
				return msg.Write(e)
			})
		},
	}
}

// newLogtailBatchHandler creates a new, initialised logtail batch handler.
func newLogtailBatchHandler() *logtailBatchHandler {
	h := &logtailBatchHandler{}
	h.init()
	return h
}

// configure applies configuration to the logtail batch handler.
func (h *logtailBatchHandler) configure(key cfgkey, value any) error {
	switch key {
	case logtailEndpoint:
		h.endpoint = value.(string)
	case logtailSourceToken:
		h.token = value.(string)
	default:
		return fmt.Errorf("%w: %w: %s", ErrLogtailConfiguration, ErrKeyNotSupported, key)
	}
	return nil
}

// send sends a batch of entries to the BetterStack Logs service.
func (h *logtailBatchHandler) send(batch *Batch) error {
	tracef("logtail.send: sending %d entries", batch.len)

	buf := h.buf.Get().(*bytes.Buffer)
	buf.Reset()

	// encodeBatch is a function ref to enable testing of an encoding
	// error by replacing the func with a mock
	if err := h.encodeBatch(buf, batch); err != nil {
		tracef("logtail.send: batch encoding failed: %s", err)
		return err
	}

	trace("sending: ", buf.String())

	body := bytes.NewReader(buf.Bytes())

	client := http.Client{
		Timeout: 5 * time.Second,
	}
	rq, err := http.NewRequest(http.MethodPost, h.endpoint, body)
	if err != nil {
		tracef("logtail.send: error initialising request: %s", err)
		return err
	}

	rq.Header.Add("Authorization", fmt.Sprintf("Bearer %s", h.token))
	rq.Header.Add("Content-Type", "application/msgpack")
	rw, err := client.Do(rq)
	if err != nil {
		tracef("logtail.send: error sending request: %s", err)
		return err
	}

	tracef("logtail.send: result: %d", rw.Status)

	return nil
}
