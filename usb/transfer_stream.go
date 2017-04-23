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

package usb

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
	// err is the first error encountered, returned to the user as soon
	// as all remaining data was read.
	err error
}

func (s *stream) cleanup() {
	close(s.transfers)
	for t := range s.transfers {
		t.cancel()
		t.wait()
		t.free()
	}
}

type ReadStream struct {
	s *stream
}

// Read reads data from the transfer stream.
// The data will come from at most a single transfer, so the returned number
// might be smaller than the length of p.
// After a non-nil error is returned, all subsequent attempts to read will
// return io.ErrClosedPipe.
func (r ReadStream) Read(p []byte) (int, error) {
	s := r.s
	if s.current == nil {
		t, ok := <-s.transfers
		if !ok {
			// no more transfers in flight
			retErr := io.ErrClosedPipe
			if s.err != nil {
				retErr = s.err
				s.err = nil
			}
			return 0, retErr
		}
		n, err := t.wait()
		if err != nil {
			s.err = err
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
				s.err = err
			}
		}
		if s.err != nil {
			s.current.free()
		}
		s.current = nil
	}
	var retErr error
	if s.current == nil && s.err != nil {
		s.cleanup()
		retErr = s.err
		s.err = nil
	}
	return use, retErr
}

// Close signals that the transfer should stop. After Close is called,
// subsequent Read()s will return data from all transfers that were already
// in progress before returning an io.EOF error, unless another error
// was encountered earlier.
func (r ReadStream) Close() {
	s := r.s
	if s.err != nil {
		s.err = io.EOF
	}
}

func newStream(tt []transferIntf, submit bool) *stream {
	s := &stream{
		transfers: make(chan transferIntf, len(tt)),
	}
	for _, t := range tt {
		s.transfers <- t
	}
	if submit {
		for _, t := range tt {
			if err := t.submit(); err != nil {
				s.err = err
				break
			}
		}
	}
	return s
}
