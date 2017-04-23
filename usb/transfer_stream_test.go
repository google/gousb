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

import (
	"errors"
	"io"
	"testing"
)

var fakeTransferBuf = make([]byte, 1500)

type fakeStreamResult struct {
	n         int
	waitErr   error
	submitErr error
}

type fakeStreamTransfer struct {
	res      []fakeStreamResult
	inFlight bool
	released bool
}

func (f *fakeStreamTransfer) submit() error {
	if f.released {
		return errors.New("submit() called on a free()d transfer")
	}
	if f.inFlight {
		return errors.New("submit() called twice")
	}
	if len(f.res) == 0 {
		return io.ErrUnexpectedEOF
	}
	f.inFlight = true
	res := f.res[0]
	if res.submitErr != nil {
		f.res = nil
		return res.submitErr
	}
	return nil
}

func (f *fakeStreamTransfer) wait() (int, error) {
	if f.released {
		return 0, errors.New("wait() called on a free()d transfer")
	}
	if !f.inFlight {
		return 0, errors.New("wait() called without submit()")
	}
	if len(f.res) == 0 {
		return 0, io.ErrUnexpectedEOF
	}
	f.inFlight = false
	res := f.res[0]
	if res.waitErr != nil {
		f.res = nil
	}
	return res.n, res.waitErr
}

func (f *fakeStreamTransfer) free() error {
	if f.released {
		return errors.New("free() called twice")
	}
	f.released = true
	return nil
}

func (f *fakeStreamTransfer) cancel() error { return nil }
func (f *fakeStreamTransfer) data() []byte  { return fakeTransferBuf }

var sentinelError = errors.New("sentinel error")

func TestReadStream(t *testing.T) {
	transfers := []*fakeStreamTransfer{
		{res: []fakeStreamResult{
			{n: 500},
		}},
		{res: []fakeStreamResult{
			{n: 500},
		}},
		{res: []fakeStreamResult{
			{n: 123, waitErr: sentinelError},
		}},
		{res: []fakeStreamResult{
			{n: 500},
		}},
	}
	intfs := make([]transferIntf, len(transfers))
	for i := range transfers {
		intfs[i] = transfers[i]
	}
	s := ReadStream{newStream(intfs, true)}
	buf := make([]byte, 400)
	for _, rs := range []struct {
		want int
		err  error
	}{
		{400, nil},
		{100, nil},
		{400, nil},
		{100, nil},
		{123, sentinelError},
		{0, io.ErrClosedPipe},
	} {
		n, err := s.Read(buf)
		if n != rs.want {
			t.Errorf("Read(): got %d bytes, want %d", n, rs.want)
		}
		if err != rs.err {
			t.Errorf("Read(): got error %v, want %v", err, rs.err)
		}
	}
	for i := range transfers {
		if !transfers[i].released {
			t.Errorf("Transfer #%d was not freed after stream completed", i)
		}
	}
}
