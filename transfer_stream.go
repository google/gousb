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
	// delayedErr is the delayed error, returned to the user after all
	// remaining data was read.
	delayedErr error
}

func (s *stream) setDelayedErr(err error) {
	if s.delayedErr == nil {
		s.delayedErr = err
		close(s.transfers)
	}
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
			return 0, s.delayedErr
		}
		n, err := t.wait()
		if err != nil {
			// wait error aborts immediately, all remaining data is invalid.
			t.free()
			if s.delayedErr == nil {
				close(s.transfers)
			}
			for t := range s.transfers {
				t.cancel()
				t.wait()
				t.free()
			}
			s.transfers = nil
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
		if s.delayedErr == nil {
			if err := s.current.submit(); err == nil {
				// guaranteed to not block, len(transfers) == number of allocated transfers
				s.transfers <- s.current
			} else {
				s.setDelayedErr(err)
			}
		}
		if s.delayedErr != nil {
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
	r.s.setDelayedErr(io.EOF)
	return nil
}

// WriteStream is a buffer that will send data asynchronously, reducing
// the latency between subsequent Write()s.
/*
type WriteStream struct {
	s *stream
}
*/

// Write sends the data to the endpoint. Write returning a nil error doesn't
// mean that data was written to the device, only that it was written to the
// buffer. Only a call to Flush() that returns nil error guarantees that
// all transfers have succeeded.
// TODO(sebek): not implemented and tested yet
/*
func (w WriteStream) Write(p []byte) (int, error) {
	s := w.s
	written := 0
	all := len(p)
	for written < all {
		if s.current == nil {
			s.current = <-s.transfers
			s.total = len(s.current.data())
			s.used = 0
		}
		use := all - written
		if use > s.total {
			use = s.total
		}
		copy(s.current.data()[s.used:], p[written:written+use])
	}
	return 0, nil
}

func (w WriteStream) Flush() error {
	return nil
}
*/

func newStream(tt []transferIntf, submit bool) *stream {
	s := &stream{
		transfers: make(chan transferIntf, len(tt)),
	}
	for _, t := range tt {
		if submit {
			if err := t.submit(); err != nil {
				t.free()
				s.setDelayedErr(err)
				break
			}
		}
		s.transfers <- t
	}
	return s
}
