package ulog

import (
	"bytes"
	"sync"
)

// FieldId is the type of a field identifier, providing an id for each standard
// field in a log entry.  "Standard" fields are those that are always present
// in a log entry (depending on logger configuration).
type FieldId int

const (
	TimeField             FieldId = iota // TimeField is the FieldId of the time field
	LevelField                           // LevelKey is the FieldId of the level field
	MessageField                         // MessageKey is the FieldId of the message field
	CallsiteFileField                    // CallsiteFileField is the FieldId of the file field; used only when call-site logging is enabled
	CallsiteFunctionField                // CallsiteFunctionField is the FieldId of the function field; used only when call-site logging is enabled

	numFields = 5 // the maximum number of standard fields in a log entry
)

// fields is a struct that holds custom fields associated with a *logcontext.
//
// the struct has a mutex to protect concurrent access to the fields and
// associated buffer maps.
//
// the buffer maps enable formatted fields to be cached for the lifetime
// of a *logcontext.  A Formatter obtains a buffer from the map using a
// key specific to that Formatter; if the buffer is empty, the Formatter
// will format the fields into the buffer, and then return the buffer to
// the  map with the corresponding Formatter key.
type fields struct {
	mutex                // protects concurrent access to the fields and buffer maps
	m     map[string]any // map of field values, keyed by field name
	b     map[int][]byte // map of formatted fields buffers, keyed by Formatter
}

// newFields returns a new fields struct initialised with an empty map of a
// specified capacity.
//
// If the specified capacity is 0 then no fields struct is created and nil is returned.
func newFields(cap int) *fields {
	if cap == 0 {
		return nil
	}

	return &fields{
		mutex: &sync.Mutex{},
		m:     make(map[string]any, cap),
		b:     map[int][]byte{},
	}
}

// merge returns a copy of the logger fields combined with the specified fields.
func (f *fields) merge(with map[string]any) *fields {
	switch {
	case len(with) == 0: // not merging anything, just return the original fields
		return f
	case f != nil: // merging into existing fields, acquire the mutex
		f.Lock()
		defer f.Unlock()
	}

	cpy := func() *fields {
		if f == nil {
			return newFields(len(with))
		}

		cpy := newFields(len(f.m) + len(with))
		for k, v := range f.m {
			cpy.m[k] = v
		}
		return cpy
	}()

	for k, v := range with {
		cpy.m[k] = v
	}
	return cpy
}

// getFormattedBytes returns the bytes.Buffer for the specified key or
// a new *bytes.Buffer if no buffer exists for the key.
//
// If the fields struct is nil then nil is returned, indicating that
// no buffer is available or required.
func (f *fields) getFormattedBytes(id int) *bytes.Buffer {
	if f == nil {
		return nil
	}

	f.Lock()
	defer f.Unlock()

	if b, ok := f.b[id]; ok {
		return bytes.NewBuffer(b)
	}
	return &bytes.Buffer{} //FUTURE: is it worth using a pool of buffers?
}

// setFormattedBytes sets the bytes.Buffer for the specified key.
func (f *fields) setFormattedBytes(id int, b []byte) {
	f.Lock()
	defer f.Unlock()

	f.b[id] = b
}
