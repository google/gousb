// Copyright 2017 the gousb Authors.  All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package gousb

import "io"

type transferIntf interface {
	submit() error
	cancel() error
	wait() (int, error)
	free() error
	data() []byte
}

type stream struct {
	// a fifo of USB transfers.
	transfers chan transferIntf
	// current holds the last transfer to return.
	current transferIntf
	// total/used are the number of all/used bytes in the current transfer.
	total, used int
	// err is the first encountered error, returned to the user.
	err error
}

func (s *stream) setErr(err error) {
	if s.err == nil {
		s.err = err
		close(s.transfers)
	}
}

func (s *stream) done() {
	if s.err == nil {
		close(s.transfers)
	}
}

func (s *stream) flush() {
	for t := range s.transfers {
		t.cancel()
		t.wait()
		t.free()
	}
	s.transfers = nil
}

// ReadStream is a buffer that tries to prefetch data from the IN endpoint,
// reducing the latency between subsequent Read()s.
// ReadStream keeps prefetching data until Close() is called or until
// an error is encountered. After Close(), the buffer might still have
// data left from transfers that were initiated before Close. Read()ing
// from the ReadStream will keep returning available data. When no more
// data is left, io.EOF is returned.
type ReadStream struct {
	s *stream
}

// Read reads data from the transfer stream.
// The data will come from at most a single transfer, so the returned number
// might be smaller than the length of p.
// After a non-nil error is returned, all subsequent attempts to read will
// return io.ErrClosedPipe.
// Read cannot be called concurrently with other Read or Close.
func (r ReadStream) Read(p []byte) (int, error) {
	s := r.s
	if s.transfers == nil {
		return 0, io.ErrClosedPipe
	}
	if s.current == nil {
		t, ok := <-s.transfers
		if !ok {
			// no more transfers in flight
			s.transfers = nil
			return 0, s.err
		}
		n, err := t.wait()
		if err != nil {
			// wait error aborts immediately, all remaining data is invalid.
			t.free()
			s.done()
			s.flush()
			return n, err
		}
		s.current = t
		s.total = n
		s.used = 0
	}
	use := s.total - s.used
	if use > len(p) {
		use = len(p)
	}
	copy(p, s.current.data()[s.used:s.used+use])
	s.used += use
	if s.used == s.total {
		if s.err == nil {
			if err := s.current.submit(); err == nil {
				// guaranteed to not block, len(transfers) == number of allocated transfers
				s.transfers <- s.current
			} else {
				s.setErr(err)
			}
		}
		if s.err != nil {
			s.current.free()
		}
		s.current = nil
	}
	return use, nil
}

// Close signals that the transfer should stop. After Close is called,
// subsequent Read()s will return data from all transfers that were already
// in progress before returning an io.EOF error, unless another error
// was encountered earlier.
// Close cannot be called concurrently with Read.
func (r ReadStream) Close() error {
	if r.s.transfers == nil {
		return nil
	}
	r.s.setErr(io.EOF)
	return nil
}

// WriteStream is a buffer that will send data asynchronously, reducing
// the latency between subsequent Write()s.
type WriteStream struct {
	s     *stream
	total int
}

// Write sends the data to the endpoint. Write returning a nil error doesn't
// mean that data was written to the device, only that it was written to the
// buffer. Only a call to Close() that returns nil error guarantees that
// all transfers have succeeded.
// If the slice passed to Write does not align exactly with the transfer
// buffer size (as declared in a call to NewStream), the last USB transfer
// of this Write will be sent with less data than the full buffer.
// After a non-nil error is returned, all subsequent attempts to write will
// return io.ErrClosedPipe.
// If Write encounters an error when preparing the transfer, the stream
// will still try to complete any pending transfers. The total number
// of bytes successfully written can be retrieved through a Written()
// call after Close() has returned.
// Write cannot be called concurrently with another Write, Written or Close.
func (w *WriteStream) Write(p []byte) (int, error) {
	if w.s.transfers == nil || w.s.err != nil {
		return 0, io.ErrClosedPipe
	}
	written := 0
	all := len(p)
	for written < all {
		t := <-w.s.transfers
		n, err := t.wait() // unsubmitted transfers will return 0 bytes and no error
		w.total += n
		if err != nil {
			t.free()
			w.s.setErr(err)
			return written, err
		}
		use := all - written
		if max := len(t.data()); use > max {
			use = max
		}
		copy(t.data(), p[written:written+use])
		if err := t.submit(); err != nil {
			t.free()
			w.s.setErr(err)
			return written, err
		}
		written += use
		w.s.transfers <- t // guaranteed non blocking
	}
	return written, nil
}

// Close signals end of data to write. Close blocks until all transfers
// that were sent are finished. The error returned by Close is the first
// error encountered during writing the entire stream (if any).
// Close returning nil indicates all transfers completed successfuly.
// After Close, the total number of bytes successfuly written can be
// retrieved using Written().
func (w *WriteStream) Close() error {
	if w.s.transfers == nil {
		return w.s.err
	}
	w.s.done()
	for t := range w.s.transfers {
		n, err := t.wait()
		w.total += n
		if w.s.err == nil && err != nil {
			w.s.err = err
		}
		t.free()
	}
	w.s.transfers = nil
	return w.s.err
}

// Written returns the number of bytes successfuly written by the stream.
// Written may be called only after Close() has been called and returned.
func (w *WriteStream) Written() int {
	return w.total
}

func newStream(tt []transferIntf, submit bool) *stream {
	s := &stream{
		transfers: make(chan transferIntf, len(tt)),
	}
	for _, t := range tt {
		if submit {
			if err := t.submit(); err != nil {
				t.free()
				s.setErr(err)
				break
			}
		}
		s.transfers <- t
	}
	return s
}
