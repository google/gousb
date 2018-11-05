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

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"reflect"
	"strconv"
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
		return errors.New("submit() called but fake result missing")
	}
	f.inFlight = true
	res := f.res[0]
	if res.submitErr != nil {
		f.res = nil
		return res.submitErr
	}
	return nil
}

func (f *fakeStreamTransfer) wait(ctx context.Context) (int, error) {
	if f.released {
		return 0, errors.New("wait() called on a free()d transfer")
	}
	if !f.inFlight {
		return 0, nil
	}
	if len(f.res) == 0 {
		return 0, errors.New("wait() called but fake result missing")
	}
	f.inFlight = false
	res := f.res[0]
	if res.waitErr == nil {
		f.res = f.res[1:]
	} else {
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

var errSentinel = errors.New("sentinel error")

type readRes struct {
	n   int
	err error
}

func (r readRes) String() string {
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "<%d bytes", r.n)
	if r.err != nil {
		fmt.Fprintf(&buf, ", error: %s", r.err.Error())
	}
	buf.WriteString(">")
	return buf.String()
}

func TestTransferReadStream(t *testing.T) {
	t.Parallel()
	for tcNum, tc := range []struct {
		desc        string
		closeBefore int
		// transfers is a list of allocated transfers, each transfers
		// carries a list of results for subsequent submits/waits.
		transfers [][]fakeStreamResult
		want      []readRes
	}{
		{
			desc:        "two transfers submitted, close, read returns both and EOF",
			closeBefore: 1,
			transfers: [][]fakeStreamResult{
				{{n: 400}},
				{{n: 400}},
			},
			want: []readRes{
				{n: 400},
				{n: 400},
				{err: io.EOF},
				{err: io.ErrClosedPipe},
			},
		},
		{
			desc:        "two transfers, two and a half cycles through transfer queue",
			closeBefore: 4,
			transfers: [][]fakeStreamResult{
				{{n: 400}, {n: 400}, {n: 400}, {waitErr: errors.New("fake wait error")}},
				{{n: 400}, {n: 400}, {waitErr: errors.New("fake wait error")}},
			},
			want: []readRes{
				{n: 400},
				{n: 400},
				{n: 400},
				{n: 400},
				{n: 400},
				{err: io.EOF},
				{err: io.ErrClosedPipe},
			},
		},
		{
			desc: "4 transfers submitted, two return, third fails on wait",
			transfers: [][]fakeStreamResult{
				{{n: 500}},
				{{n: 500}},
				{{n: 123, waitErr: errSentinel}},
				{{n: 500}},
			},
			want: []readRes{
				{n: 400},
				{n: 100},
				{n: 400},
				{n: 100},
				{n: 123, err: errSentinel},
				{err: io.ErrClosedPipe},
			},
		},
		{
			desc: "2 transfers, second submit fails initialization but error overshadowed by wait error",
			transfers: [][]fakeStreamResult{
				{{n: 123, waitErr: errSentinel}},
				{{submitErr: errors.New("fake submit error")}},
			},
			want: []readRes{
				{n: 123, err: errSentinel},
				{err: io.ErrClosedPipe},
			},
		},
		{
			desc: "2 transfers, second submit fails during initialization",
			transfers: [][]fakeStreamResult{
				{{n: 400}},
				{{submitErr: errSentinel}},
			},
			want: []readRes{
				{n: 400},
				{err: errSentinel},
				{err: io.ErrClosedPipe},
			},
		},
		{
			desc: "2 transfers, 3rd submit fails during second round",
			transfers: [][]fakeStreamResult{
				{{n: 400}, {submitErr: errSentinel}},
				{{n: 400}},
			},
			want: []readRes{
				{n: 400},
				{n: 400},
				{err: errSentinel},
				{err: io.ErrClosedPipe},
			},
		},
		{
			desc: "fail quickly",
			transfers: [][]fakeStreamResult{
				{{waitErr: errSentinel}},
				{{n: 500}},
				{{n: 500}},
			},
			want: []readRes{
				{err: errSentinel},
				{err: io.ErrClosedPipe},
			},
		},
	} {
		tcNum, tc := tcNum, tc // t.Parallel will delay the execution of the test, save the iteration values.
		t.Run(strconv.Itoa(tcNum), func(t *testing.T) {
			t.Parallel()
			t.Logf("Case %d: %s", tcNum, tc.desc)
			ftt := make([]*fakeStreamTransfer, len(tc.transfers))
			tt := make([]transferIntf, len(tc.transfers))
			for i := range tc.transfers {
				ftt[i] = &fakeStreamTransfer{
					res: tc.transfers[i],
				}
				tt[i] = ftt[i]
			}
			s := ReadStream{s: newStream(tt)}
			s.s.submitAll()
			buf := make([]byte, 400)
			got := make([]readRes, len(tc.want))
			for i := range tc.want {
				if i == tc.closeBefore-1 {
					t.Log("Close()")
					s.Close()
				}
				n, err := s.Read(buf)
				t.Logf("Read(): got %d, %v", n, err)
				got[i] = readRes{
					n:   n,
					err: err,
				}
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("Got Read() results:\n%v\nwant Read() results:\n%v\n", got, tc.want)
			}
			for i := range ftt {
				if !ftt[i].released {
					t.Errorf("Transfer #%d was not freed after stream completed", i)
				}
			}
		})
	}
}

func TestTransferWriteStream(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		desc      string
		transfers [][]fakeStreamResult
		writes    []int
		want      []int
		total     int
		err       error
	}{
		{
			desc: "successful two transfers",
			transfers: [][]fakeStreamResult{
				{{n: 1500}},
				{{n: 1500}},
				{{n: 1500}},
			},
			writes: []int{3000},
			want:   []int{3000},
			total:  3000,
		},
		{
			desc: "submit failed on second transfer",
			transfers: [][]fakeStreamResult{
				{{n: 1500}},
				{{submitErr: errSentinel}},
				{{n: 1500}},
			},
			writes: []int{3000},
			want:   []int{1500},
			total:  1500,
			err:    errSentinel,
		},
		{
			desc: "wait failed on second transfer",
			transfers: [][]fakeStreamResult{
				{{n: 1500}},
				{{waitErr: errSentinel}},
				{{n: 1500}},
			},
			writes: []int{3000, 1500},
			want:   []int{3000, 1500},
			total:  1500,
			err:    errSentinel,
		},
		{
			desc: "reused transfer",
			transfers: [][]fakeStreamResult{
				{{n: 1500}, {n: 1500}},
				{{n: 1500}, {n: 1500}},
				{{n: 1500}, {n: 500}},
			},
			writes: []int{3000, 3000, 2000},
			want:   []int{3000, 3000, 2000},
			total:  8000,
		},
		{
			desc: "wait failed on reused transfer",
			transfers: [][]fakeStreamResult{
				{{n: 1500}, {n: 1500}},
				{{waitErr: errSentinel}, {n: 1500}},
				{{n: 1500}, {n: 1500}},
			},
			writes: []int{1500, 1500, 1500, 1500, 1500},
			want:   []int{1500, 1500, 1500, 1500, 0},
			total:  1500,
			err:    errSentinel,
		},
	} {
		tc := tc
		t.Run(tc.desc, func(t *testing.T) {
			t.Parallel()
			ftt := make([]*fakeStreamTransfer, len(tc.transfers))
			tt := make([]transferIntf, len(tc.transfers))
			for i := range tc.transfers {
				ftt[i] = &fakeStreamTransfer{
					res: tc.transfers[i],
				}
				tt[i] = ftt[i]
			}
			s := WriteStream{s: newStream(tt)}
			for i, w := range tc.writes {
				got, err := s.Write(make([]byte, w))
				if want := tc.want[i]; got != want {
					t.Errorf("WriteStream.Write #%d: got %d, want %d", i, got, want)
				}
				if err != nil && err != tc.err {
					t.Errorf("WriteStream.Write: got error %v, want %v", err, tc.err)
				}
			}
			if err := s.Close(); err != tc.err {
				t.Fatalf("WriteStream.Close: got %v, want %v", err, tc.err)
			}
			if err := s.Close(); err != io.ErrClosedPipe {
				t.Fatalf("second WriteStream.Close: got %v, want %v", err, io.ErrClosedPipe)
			}
			if got := s.Written(); got != tc.total {
				t.Fatalf("WriteStream.Written: got %d, want %d", got, tc.total)
			}
		})
	}
}
