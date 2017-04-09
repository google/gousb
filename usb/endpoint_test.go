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
	"reflect"
	"testing"
	"time"
)

// IN bulk endpoint
var testBulkInEP = EndpointInfo{
	Number:        2,
	Direction:     EndpointDirectionIn,
	MaxPacketSize: 512,
	TransferType:  TransferTypeBulk,
}

var testBulkInSetting = InterfaceSetting{
	Number:    0,
	Alternate: 0,
	Class:     ClassVendorSpec,
	Endpoints: []EndpointInfo{testBulkInEP},
}

// OUT iso endpoint
var testIsoOutEP = EndpointInfo{
	Number:        6,
	MaxPacketSize: 3 * 1024,
	TransferType:  TransferTypeIsochronous,
	PollInterval:  125 * time.Microsecond,
	UsageType:     IsoUsageTypeData,
}

var testIsoOutSetting = InterfaceSetting{
	Number:    0,
	Alternate: 0,
	Class:     ClassVendorSpec,
	Endpoints: []EndpointInfo{testIsoOutEP},
}

func TestEndpoint(t *testing.T) {
	defer func(i libusbIntf) { libusb = i }(libusb)

	for _, epCfg := range []struct {
		method string
		InterfaceSetting
		EndpointInfo
	}{
		{"Read", testBulkInSetting, testBulkInEP},
		{"Write", testIsoOutSetting, testIsoOutEP},
	} {
		t.Run(epCfg.method, func(t *testing.T) {
			for _, tc := range []struct {
				desc    string
				buf     []byte
				ret     int
				status  TransferStatus
				want    int
				wantErr bool
			}{
				{
					desc: "empty buffer",
					buf:  nil,
					ret:  10,
					want: 0,
				},
				{
					desc: "128B buffer, 60 transferred",
					buf:  make([]byte, 128),
					ret:  60,
					want: 60,
				},
				{
					desc:    "128B buffer, 10 transferred and then error",
					buf:     make([]byte, 128),
					ret:     10,
					status:  TransferError,
					want:    10,
					wantErr: true,
				},
			} {
				lib := newFakeLibusb()
				libusb = lib
				ep := newEndpoint(nil, epCfg.InterfaceSetting, epCfg.EndpointInfo, time.Second, time.Second)
				op, ok := reflect.TypeOf(ep).MethodByName(epCfg.method)
				if !ok {
					t.Fatalf("method %s not found in endpoint struct", epCfg.method)
				}
				go func() {
					fakeT := lib.waitForSubmitted()
					fakeT.length = tc.ret
					fakeT.status = tc.status
					close(fakeT.done)
				}()
				opv := op.Func.Interface().(func(*Endpoint, []byte) (int, error))
				got, err := opv(ep, tc.buf)
				if (err != nil) != tc.wantErr {
					t.Errorf("%s: bulkInEP.Read(): got err: %v, err != nil is %v, want %v", tc.desc, err, err != nil, tc.wantErr)
					continue
				}
				if got != tc.want {
					t.Errorf("%s: bulkInEP.Read(): got %d bytes, want %d", tc.desc, got, tc.want)
				}
			}
		})
	}
}

func TestEndpointWrongDirection(t *testing.T) {
	ep := &Endpoint{
		InterfaceSetting: testBulkInSetting,
		Info:             testBulkInEP,
	}
	_, err := ep.Write([]byte{1, 2, 3})
	if err == nil {
		t.Error("bulkInEP.Write(): got nil error, want non-nil")
	}
	ep = &Endpoint{
		InterfaceSetting: testIsoOutSetting,
		Info:             testIsoOutEP,
	}
	_, err = ep.Read(make([]byte, 64))
	if err == nil {
		t.Error("isoOutEP.Read(): got nil error, want non-nil")
	}
}

func TestEndpointInfo(t *testing.T) {
	for _, tc := range []struct {
		ep   EndpointInfo
		want string
	}{
		{
			ep: EndpointInfo{
				Number:        6,
				Direction:     EndpointDirectionIn,
				TransferType:  TransferTypeBulk,
				MaxPacketSize: 512,
			},
			want: "Endpoint #6 IN (address 0x86) bulk [512 bytes]",
		},
		{
			ep: EndpointInfo{
				Number:        2,
				Direction:     EndpointDirectionOut,
				TransferType:  TransferTypeIsochronous,
				MaxPacketSize: 512,
				IsoSyncType:   IsoSyncTypeAsync,
				UsageType:     IsoUsageTypeData,
			},
			want: "Endpoint #2 OUT (address 0x02) isochronous - asynchronous data [512 bytes]",
		},
		{
			ep: EndpointInfo{
				Number:        3,
				Direction:     EndpointDirectionIn,
				TransferType:  TransferTypeInterrupt,
				MaxPacketSize: 16,
				UsageType:     InterruptUsageTypePeriodic,
			},
			want: "Endpoint #3 IN (address 0x83) interrupt - periodic [16 bytes]",
		},
	} {
		if got := tc.ep.String(); got != tc.want {
			t.Errorf("%#v.String(): got %q, want %q", tc.ep, got, tc.want)
		}
	}
}
