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
	"reflect"
	"testing"
	"time"
)

type fakeTransfer struct {
	buf []byte
	ret int
	err error
}

func (t *fakeTransfer) submit() error      { return nil }
func (t *fakeTransfer) wait() (int, error) { return t.ret, t.err }
func (t *fakeTransfer) free() error        { return nil }

func TestEndpoint(t *testing.T) {
	for _, epCfg := range []struct {
		method string
		InterfaceSetup
		EndpointInfo
	}{
		{"Read", testBulkInSetup, testBulkInEP},
		{"Write", testIsoOutSetup, testIsoOutEP},
	} {
		t.Run(epCfg.method, func(t *testing.T) {
			for _, tc := range []struct {
				buf     []byte
				ret     int
				err     error
				want    int
				wantErr bool
			}{
				{buf: nil, ret: 10, want: 0},
				{buf: make([]byte, 128), ret: 60, want: 60},
				{buf: make([]byte, 128), ret: 10, err: errors.New("some error"), want: 10, wantErr: true},
			} {
				ep := &endpoint{
					InterfaceSetup: epCfg.InterfaceSetup,
					EndpointInfo:   epCfg.EndpointInfo,
					newUSBTransfer: func(buf []byte, timeout time.Duration) (transferIntf, error) {
						return &fakeTransfer{buf: buf, ret: tc.ret, err: tc.err}, nil
					},
				}
				op, ok := reflect.TypeOf(ep).MethodByName(epCfg.method)
				if !ok {
					t.Fatalf("method %s not found in endpoint struct", epCfg.method)
				}
				opv := op.Func.Interface().(func(*endpoint, []byte) (int, error))
				got, err := opv(ep, tc.buf)
				if (err != nil) != tc.wantErr {
					t.Errorf("bulkInEP.Read(): got err: %v, err != nil is %v, want %v", err, err != nil, tc.wantErr)
					continue
				}
				if got != tc.want {
					t.Errorf("bulkInEP.Read(): got %d bytes, want %d", got, tc.want)
				}
			}
		})
	}
}

func TestEndpointWrongDirection(t *testing.T) {
	ep := &endpoint{
		InterfaceSetup: testBulkInSetup,
		EndpointInfo:   testBulkInEP,
	}
	_, err := ep.Write([]byte{1, 2, 3})
	if err == nil {
		t.Error("bulkInEP.Write(): got nil error, want non-nil")
	}
	ep = &endpoint{
		InterfaceSetup: testIsoOutSetup,
		EndpointInfo:   testIsoOutEP,
	}
	_, err = ep.Read(make([]byte, 64))
	if err == nil {
		t.Error("isoOutEP.Read(): got nil error, want non-nil")
	}
}
