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
	_, done := newFakeLibusb()
	defer done()

	c := NewContext()
	defer c.Close()
	dev, err := c.OpenDeviceWithVidPid(0x8888, 0x0002)
	if dev == nil {
		t.Fatal("OpenDeviceWithVidPid(0x8888, 0x0002): got nil device, need non-nil")
	}
	defer dev.Close()
	if err != nil {
		t.Fatalf("OpenDeviceWithVidPid(0x8888, 0x0002): %v", err)
	}
	cfg, err := dev.Config(1)
	if err != nil {
		t.Fatalf("%s.Config(1): %v", dev, err)
	}
	defer cfg.Close()
	intf, err := cfg.Interface(1, 1)
	if err != nil {
		t.Fatalf("%s.Interface(1, 1): %v", cfg, err)
	}
	defer intf.Close()
	got, err := intf.InEndpoint(6)
	if err != nil {
		t.Fatalf("%s.InEndpoint(6): got error %v, want nil", intf, err)
	}
	if want := fakeDevices[1].Configs[0].Interfaces[1].AltSettings[1].Endpoints[1]; !reflect.DeepEqual(got.Info, want) {
		t.Errorf("%s.InEndpoint(6): got %+v, want %+v", intf, got, want)
	}

	if _, err := cfg.Interface(1, 0); err == nil {
		t.Fatalf("%s.Interface(1, 0): got nil, want non nil, because Interface 1 is already claimed.", cfg)
	}

	// intf2 is interface #0, not claimed yet.
	intf2, err := cfg.Interface(0, 0)
	if err != nil {
		t.Fatalf("%s.Interface(0, 0): got %v, want nil", cfg, err)
	}

	if err := cfg.Close(); err == nil {
		t.Fatalf("%s.Close(): got nil, want non nil, because the Interface was not released.", cfg)
	}
	if err := dev.Close(); err == nil {
		t.Fatalf("%s.Close(): got nil, want non nil, because the Config was not released.", cfg)
	}

	intf.Close()
	if err := cfg.Close(); err == nil {
		t.Fatalf("%s.Close(): got nil, want non nil, because the Interface was not released.", cfg)
	}
	if err := dev.Close(); err == nil {
		t.Fatalf("%s.Close(): got nil, want non nil, because the Config was not released.", dev)
	}

	intf2.Close()
	if err := dev.Close(); err == nil {
		t.Fatalf("%s.Close(): got nil, want non nil, because the Config was not released.", dev)
	}

	if err := cfg.Close(); err != nil {
		t.Fatalf("%s.Close(): got error %v, want nil", cfg, err)
	}

	if err := dev.Close(); err != nil {
		t.Fatalf("%s.Close(): got error %v, want nil", dev, err)
	}
}
