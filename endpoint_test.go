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

func TestEndpoint(t *testing.T) {
	// Can't be parallelized, newFakeLibusb modifies a shared global state.
	lib, done := newFakeLibusb()
	defer done()

	for _, epData := range []struct {
		ei   EndpointDesc
		intf InterfaceSetting
	}{
		{
			ei: EndpointDesc{
				Address:       0x82,
				Number:        2,
				Direction:     EndpointDirectionIn,
				MaxPacketSize: 512,
				TransferType:  TransferTypeBulk,
			},
			intf: InterfaceSetting{
				Number:    0,
				Alternate: 0,
				Class:     ClassVendorSpec,
			},
		},
		{
			ei: EndpointDesc{
				Address:       0x06,
				Number:        6,
				MaxPacketSize: 3 * 1024,
				TransferType:  TransferTypeIsochronous,
				PollInterval:  125 * time.Microsecond,
				UsageType:     IsoUsageTypeData,
			},
			intf: InterfaceSetting{
				Number:    0,
				Alternate: 0,
				Class:     ClassVendorSpec,
			},
		},
	} {
		epData.intf.Endpoints = map[EndpointAddress]EndpointDesc{epData.ei.Address: epData.ei}
		for _, tc := range []struct {
			desc       string
			buf        []byte
			ret        int
			wantSubmit bool
			status     TransferStatus
			want       int
			wantErr    bool
		}{
			{
				desc:       "empty buffer",
				buf:        nil,
				ret:        10,
				wantSubmit: false,
				want:       0,
			},
			{
				desc:       "128B buffer, 60 transferred",
				buf:        make([]byte, 128),
				ret:        60,
				wantSubmit: true,
				want:       60,
			},
			{
				desc:       "128B buffer, 10 transferred and then error",
				buf:        make([]byte, 128),
				ret:        10,
				wantSubmit: true,
				status:     TransferError,
				want:       10,
				wantErr:    true,
			},
		} {
			ep := &endpoint{h: nil, InterfaceSetting: epData.intf, Desc: epData.ei}
			if tc.wantSubmit {
				go func() {
					fakeT := lib.waitForSubmitted()
					fakeT.length = tc.ret
					fakeT.status = tc.status
					close(fakeT.done)
				}()
			}
			got, err := ep.transfer(tc.buf)
			if (err != nil) != tc.wantErr {
				t.Errorf("%s, %s: ep.transfer(...): got err: %v, err != nil is %v, want %v", epData.ei, tc.desc, err, err != nil, tc.wantErr)
				continue
			}
			if got != tc.want {
				t.Errorf("%s, %s: ep.transfer(...): got %d bytes, want %d", epData.ei, tc.desc, got, tc.want)
			}
			if !lib.empty() {
				t.Fatalf("%s, %s: transfers still pending when none were expected", epData.ei, tc.desc)
			}
		}
	}
}

func TestEndpointInfo(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		ep   EndpointDesc
		want string
	}{
		{
			ep: EndpointDesc{
				Address:       0x86,
				Number:        6,
				Direction:     EndpointDirectionIn,
				TransferType:  TransferTypeBulk,
				MaxPacketSize: 512,
			},
			want: "ep #6 IN (address 0x86) bulk [512 bytes]",
		},
		{
			ep: EndpointDesc{
				Address:       0x02,
				Number:        2,
				Direction:     EndpointDirectionOut,
				TransferType:  TransferTypeIsochronous,
				MaxPacketSize: 512,
				IsoSyncType:   IsoSyncTypeAsync,
				UsageType:     IsoUsageTypeData,
			},
			want: "ep #2 OUT (address 0x02) isochronous - asynchronous data [512 bytes]",
		},
		{
			ep: EndpointDesc{
				Address:       0x83,
				Number:        3,
				Direction:     EndpointDirectionIn,
				TransferType:  TransferTypeInterrupt,
				MaxPacketSize: 16,
				UsageType:     InterruptUsageTypePeriodic,
			},
			want: "ep #3 IN (address 0x83) interrupt - periodic [16 bytes]",
		},
	} {
		if got := tc.ep.String(); got != tc.want {
			t.Errorf("%#v.String(): got %q, want %q", tc.ep, got, tc.want)
		}
	}
}

func TestEndpointInOut(t *testing.T) {
	// Can't be parallelized, newFakeLibusb modifies a shared global state.
	lib, done := newFakeLibusb()
	defer done()

	ctx := NewContext()
	defer ctx.Close()
	d, err := ctx.OpenDeviceWithVIDPID(0x9999, 0x0001)
	if err != nil {
		t.Fatalf("OpenDeviceWithVIDPID(0x9999, 0x0001): got error %v, want nil", err)
	}
	defer func() {
		if err := d.Close(); err != nil {
			t.Errorf("%s.Close(): %v", d, err)
		}
	}()
	cfg, err := d.Config(1)
	if err != nil {
		t.Fatalf("%s.Config(1): %v", d, err)
	}
	defer func() {
		if err := cfg.Close(); err != nil {
			t.Errorf("%s.Close(): %v", cfg, err)
		}
	}()
	intf, err := cfg.Interface(0, 0)
	if err != nil {
		t.Fatalf("%s.Interface(0, 0): %v", cfg, err)
	}
	defer intf.Close()

	// IN endpoint 2
	iep, err := intf.InEndpoint(2)
	if err != nil {
		t.Fatalf("%s.InEndpoint(2): got error %v, want nil", intf, err)
	}
	dataTransferred := 100
	go func() {
		fakeT := lib.waitForSubmitted()
		fakeT.length = dataTransferred
		fakeT.status = TransferCompleted
		close(fakeT.done)
	}()
	buf := make([]byte, 512)
	got, err := iep.Read(buf)
	if err != nil {
		t.Errorf("%s.Read: got error %v, want nil", iep, err)
	} else if got != dataTransferred {
		t.Errorf("%s.Read: got %d, want %d", iep, got, dataTransferred)
	}

	_, err = intf.InEndpoint(1)
	if err == nil {
		t.Errorf("%s.InEndpoint(1): got nil, want error", intf)
	}

	// OUT endpoint 1
	oep, err := intf.OutEndpoint(1)
	if err != nil {
		t.Fatalf("%s.OutEndpoint(1): got error %v, want nil", intf, err)
	}
	go func() {
		fakeT := lib.waitForSubmitted()
		fakeT.length = dataTransferred
		fakeT.status = TransferCompleted
		close(fakeT.done)
	}()
	got, err = oep.Write(buf)
	if err != nil {
		t.Errorf("%s.Write: got error %v, want nil", oep, err)
	} else if got != dataTransferred {
		t.Errorf("%s.Write: got %d, want %d", oep, got, dataTransferred)
	}

	_, err = intf.OutEndpoint(2)
	if err == nil {
		t.Errorf("%s.OutEndpoint(2): got nil, want error", intf)
	}
}

func TestSameEndpointNumberInOut(t *testing.T) {
	// Can't be parallelized, newFakeLibusb modifies a shared global state.
	_, done := newFakeLibusb()
	defer done()

	ctx := NewContext()
	defer ctx.Close()
	d, err := ctx.OpenDeviceWithVIDPID(0x1111, 0x1111)
	if err != nil {
		t.Fatalf("OpenDeviceWithVIDPID(0x1111, 0x1111): got error %v, want nil", err)
	}
	defer func() {
		if err := d.Close(); err != nil {
			t.Errorf("%s.Close(): %v", d, err)
		}
	}()
	cfg, err := d.Config(1)
	if err != nil {
		t.Fatalf("%s.Config(1): %v", d, err)
	}
	defer func() {
		if err := cfg.Close(); err != nil {
			t.Errorf("%s.Close(): %v", cfg, err)
		}
	}()
	intf, err := cfg.Interface(0, 0)
	if err != nil {
		t.Fatalf("%s.Interface(0, 0): %v", cfg, err)
	}
	defer intf.Close()

	if _, err := intf.InEndpoint(1); err != nil {
		t.Errorf("%s.InEndpoint(1): got error %v, want nil", intf, err)
	}
	if _, err := intf.OutEndpoint(1); err != nil {
		t.Errorf("%s.OutEndpoint(1): got error %v, want nil", intf, err)
	}
}
