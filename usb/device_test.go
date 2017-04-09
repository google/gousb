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
)

func TestOpenEndpoint(t *testing.T) {
	origLib := libusb
	defer func() { libusb = origLib }()
	libusb = newFakeLibusb()

	c := NewContext()
	dev, err := c.OpenDeviceWithVidPid(0x8888, 0x0002)
	if dev == nil {
		t.Fatal("OpenDeviceWithVidPid(0x8888, 0x0002): got nil device, need non-nil")
	}
	defer dev.Close()
	if err != nil {
		t.Fatalf("OpenDeviceWithVidPid(0x8888, 0x0002): got error %v, want nil", err)
	}
	got, err := dev.OpenEndpoint(0x86, 1, 1, 2)
	if err != nil {
		t.Fatalf("OpenEndpoint(cfg=1, if=1, alt=2, ep=0x86): got error %v, want nil", err)
	}
	if want := fakeDevices[1].Configs[0].Interfaces[1].AltSettings[2].Endpoints[1]; !reflect.DeepEqual(got.Info, want) {
		t.Errorf("OpenEndpoint(cfg=1, if=1, alt=2, ep=0x86): got %+v, want %+v", got, want)
	}
}