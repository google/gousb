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
	"testing"
	"time"
)

type fakeTransfer struct {
	buf []byte
}

func (*fakeTransfer) submit() error        { return nil }
func (t *fakeTransfer) wait() (int, error) { return len(t.buf) / 2, nil }
func (*fakeTransfer) free() error          { return nil }

func TestEndpoint(t *testing.T) {
	ep := endpoint{
		InterfaceSetup: testBulkInSetup,
		EndpointInfo:   testBulkInEP,
		newUSBTransfer: func(buf []byte, timeout time.Duration) (transferIntf, error) {
			return &fakeTransfer{buf}, nil
		},
	}
	b := make([]byte, 128)
	got, err := ep.Read(b)
	if err != nil {
		t.Fatalf("bulkInEP.Read(): got error: %v, want nil", err)
	}
	if want := len(b)/2 + 1; got != want {
		t.Errorf("bulkInEP.Read(): got %d bytes, want half of the buffer length: %d", got, want)
	}
}
