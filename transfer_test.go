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
	"testing"
	"time"
)

func TestNewTransfer(t *testing.T) {
	t.Parallel()
	ctx := newContextWithImpl(newFakeLibusb())
	defer func() {
		if err := ctx.Close(); err != nil {
			t.Errorf("Context.Close(): %v", err)
		}
	}()

	for _, tc := range []struct {
		desc        string
		dir         EndpointDirection
		tt          TransferType
		maxPkt      int
		buf         int
		timeout     time.Duration
		wantIso     int
		wantLength  int
		wantTimeout int
	}{
		{
			desc:       "bulk in transfer, 512B packets",
			dir:        EndpointDirectionIn,
			tt:         TransferTypeBulk,
			maxPkt:     512,
			buf:        1024,
			timeout:    time.Second,
			wantLength: 1024,
		},
		{
			desc:       "iso out transfer, 3 * 1024B packets",
			dir:        EndpointDirectionOut,
			tt:         TransferTypeIsochronous,
			maxPkt:     3 * 1024,
			buf:        10000,
			wantLength: 10000,
		},
	} {
		xfer, err := newUSBTransfer(ctx, nil, &EndpointDesc{
			Number:        2,
			Direction:     tc.dir,
			TransferType:  tc.tt,
			MaxPacketSize: tc.maxPkt,
		}, tc.buf, tc.timeout)

		if err != nil {
			t.Fatalf("newUSBTransfer(): %v", err)
		}
		defer xfer.free()
		if got, want := len(xfer.data()), tc.wantLength; got != want {
			t.Errorf("xfer.buf: got %d bytes, want %d", got, want)
		}
	}
}

func TestTransferProtocol(t *testing.T) {
	t.Parallel()
	f := newFakeLibusb()
	ctx := newContextWithImpl(f)
	defer func() {
		if err := ctx.Close(); err != nil {
			t.Errorf("Context.Close(): %v", err)
		}
	}()

	xfers := make([]*usbTransfer, 2)
	var err error
	for i := 0; i < 2; i++ {
		xfers[i], err = newUSBTransfer(ctx, nil, &EndpointDesc{
			Number:        6,
			Direction:     EndpointDirectionIn,
			TransferType:  TransferTypeBulk,
			MaxPacketSize: 512,
		}, 10240, time.Second)
		if err != nil {
			t.Fatalf("newUSBTransfer: %v", err)
		}
	}

	go func() {
		ft := f.waitForSubmitted()
		ft.length = 5
		ft.status = TransferCompleted
		copy(ft.buf, []byte{1, 2, 3, 4, 5})
		ft.done <- struct{}{}

		ft = f.waitForSubmitted()
		ft.length = 99
		ft.status = TransferCompleted
		copy(ft.buf, []byte{12, 12, 12, 12, 12})
		ft.done <- struct{}{}

		ft = f.waitForSubmitted()
		ft.length = 123
		ft.status = TransferCancelled
		ft.done <- struct{}{}
	}()

	xfers[0].submit()
	xfers[1].submit()
	got, err := xfers[0].wait()
	if err != nil {
		t.Errorf("xfer#0.wait returned error %v, want nil", err)
	}
	if want := 5; got != want {
		t.Errorf("xfer#0.wait returned %d bytes, want %d", got, want)
	}
	got, err = xfers[1].wait()
	if err != nil {
		t.Errorf("xfer#0.wait returned error %v, want nil", err)
	}
	if want := 99; got != want {
		t.Errorf("xfer#0.wait returned %d bytes, want %d", got, want)
	}

	xfers[1].submit()
	xfers[1].cancel()
	got, err = xfers[1].wait()
	if err == nil {
		t.Error("xfer#1(resubmitted).wait returned error nil, want non-nil")
	}
	if want := 123; got != want {
		t.Errorf("xfer#1(resubmitted).wait returned %d bytes, want %d", got, want)
	}

	for _, x := range xfers {
		x.cancel()
		x.wait()
		x.free()
	}
}

func BenchmarkSubSlice(b *testing.B) {
	x := make([]byte, 512)
	start, len := 50, 50
	b.Run("start:start+len", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			y := x
			y = y[start : start+len]
		}
	})
	b.Run("[start:][:len]", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			y := x
			y = y[start:][:len]
		}
	})
}
