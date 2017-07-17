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
	lib, done := newFakeLibusb()
	defer done()

	goodTransfers := 7
	go func() {
		var num int
		for {
			xfr := lib.waitForSubmitted()
			if xfr == nil {
				return
			}
			if num < goodTransfers {
				xfr.status = TransferCompleted
				xfr.length = len(xfr.buf)
			} else {
				xfr.status = TransferError
				xfr.length = 0
			}
			xfr.done <- struct{}{}
			num++
		}
	}()

	ctx := NewContext()
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
	want := goodTransfers * 1024
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
}
