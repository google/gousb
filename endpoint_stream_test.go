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

import "testing"

func TestEndpointReadStream(t *testing.T) {
	t.Parallel()
	lib := newFakeLibusb()
	ctx := newContextWithImpl(lib)
	defer func() {
		if err := ctx.Close(); err != nil {
			t.Errorf("Context.Close: %v", err)
		}
	}()

	goodTransfers := 7
	done := make(chan struct{})
	go func() {
		var num int
		for {
			xfr := lib.waitForSubmitted(done)
			if xfr == nil {
				return
			}
			if num < goodTransfers {
				xfr.setData(make([]byte, len(xfr.buf)))
				xfr.setStatus(TransferCompleted)
			} else {
				xfr.setStatus(TransferError)
			}
			num++
		}
	}()

	dev, err := ctx.OpenDeviceWithVIDPID(0x9999, 0x0001)
	if err != nil {
		t.Fatalf("OpenDeviceWithVIDPID(9999, 0001): %v", err)
	}
	defer dev.Close()
	cfg, err := dev.Config(1)
	if err != nil {
		t.Fatalf("%s.Config(1): %v", dev, err)
	}
	defer cfg.Close()
	intf, err := cfg.Interface(0, 0)
	if err != nil {
		t.Fatalf("%s.Interface(0, 0): %v", cfg, err)
	}
	defer intf.Close()
	ep, err := intf.InEndpoint(2)
	if err != nil {
		t.Fatalf("%s.Endpoint(2): %v", intf, err)
	}
	stream, err := ep.NewStream(1024, 5)
	if err != nil {
		t.Fatalf("%s.NewStream(1024, 5): %v", ep, err)
	}
	defer stream.Close()
	var got int
	want := goodTransfers * 512 // EP max packet size is 512
	buf := make([]byte, 1024)
	for got <= want {
		num, err := stream.Read(buf)
		if err != nil {
			break
		}
		got += num
	}
	if got != want {
		t.Errorf("stream.Read(): read %d bytes, want %d", got, want)
	}
	close(done)
}

func TestEndpointWriteStream(t *testing.T) {
	t.Parallel()
	lib := newFakeLibusb()
	ctx := newContextWithImpl(lib)
	defer func() {
		if err := ctx.Close(); err != nil {
			t.Errorf("Context.Close: %v", err)
		}
	}()

	done := make(chan struct{})
	total := 0
	num := 0
	go func() {
		for {
			xfr := lib.waitForSubmitted(done)
			if xfr == nil {
				return
			}
			xfr.setData(make([]byte, len(xfr.buf)))
			xfr.setStatus(TransferCompleted)
			num++
			total += xfr.length
		}
	}()

	dev, err := ctx.OpenDeviceWithVIDPID(0x9999, 0x0001)
	if err != nil {
		t.Fatalf("OpenDeviceWithVIDPID(9999, 0001): %v", err)
	}
	defer dev.Close()
	cfg, err := dev.Config(1)
	if err != nil {
		t.Fatalf("%s.Config(1): %v", dev, err)
	}
	defer cfg.Close()
	intf, err := cfg.Interface(0, 0)
	if err != nil {
		t.Fatalf("%s.Interface(0, 0): %v", cfg, err)
	}
	defer intf.Close()
	ep, err := intf.OutEndpoint(1)
	if err != nil {
		t.Fatalf("%s.Endpoint(1): %v", intf, err)
	}
	bufSize := 1024
	stream, err := ep.NewStream(bufSize, 5)
	if err != nil {
		t.Fatalf("%s.NewStream(%d, 5): %v", ep, bufSize, err)
	}
	defer stream.Close()
	for i := 0; i < 5; i++ {
		if n, err := stream.Write(make([]byte, bufSize*2)); err != nil {
			t.Fatalf("stream.Write: got error %v", err)
		} else if n != bufSize*2 {
			t.Fatalf("stream.Write: %d, want %d", n, bufSize*2)
		}
	}
	if err := stream.Close(); err != nil {
		t.Fatalf("stream.Close: got error %v", err)
	}
	if got, want := stream.Written(), 10240; got != want { // 5 stream.Writes, each with 2048 bytes
		t.Errorf("stream.Written: got %d, want %d", got, want)
	}
	done <- struct{}{}
	if w := stream.Written(); total != w {
		t.Errorf("received data: got %d, but stream.Written returned %d, results should be identical", total, w)
	}
	if wantXfers := 20; num != wantXfers { // transferred 10240 bytes, device max packet size is 512
		t.Errorf("received transfers: got %d, want %d", num, wantXfers)
	}
}
