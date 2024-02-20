package ulog

type cfgkey string

// batchHandler is the interface for batch handlers.  A transport
// will provide a batch handler when initialising a batch, which
// will use the handler to send full or flushed batches.
type batchHandler interface {
	configure(key cfgkey, value any) error
	send(*Batch) error
}

// Batch is a collection of log entries that can be written by
// a Transport in a single operation.
type Batch struct {
	entries      [][]byte // the batched entries
	size         int      // size of the batch in bytes
	len          int      // number of entries in the batch
	max          int      // maximum number of entries in the batch
	batchHandler          // the handler for the batch
}

// init initialises the batch with a handler and a maximum number
// of entries.
func (b *Batch) init(handler batchHandler, max int) {
	b.len = 0
	b.size = 0
	b.max = max
	b.entries = make([][]byte, 0, max)
	b.batchHandler = handler
}

// clear resets the batch to an empty state.
func (b *Batch) clear() {
	b.entries = make([][]byte, 0, b.max)
	b.size = 0
	b.len = 0
}

// add adds an entry to the batch.  If the batch is full it is
// sent to the handler and the batch is reset.
func (b *Batch) add(entry []byte) {
	b.entries = append(b.entries, entry)
	b.size += len(entry)
	b.len += 1

	if b.len >= b.max {
		b.flush()
	}
}

// flush sends a non-empty batch to the handler and resets.
func (b *Batch) flush() {
	if b.len > 0 {
		if err := b.send(b); err == nil {
			b.clear()
		}
	}
}
