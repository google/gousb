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
	"reflect"
	"testing"
)

func TestClaimAndRelease(t *testing.T) {
	_, done := newFakeLibusb()
	defer done()

	const (
		devIdx  = 1
		cfgNum  = 1
		if1Num  = 1
		alt1Num = 1
		ep1Num  = 6
		alt2Num = 0
		if2Num  = 0
	)
	c := NewContext()
	defer c.Close()
	dev, err := c.OpenDeviceWithVIDPID(0x8888, 0x0002)
	if dev == nil {
		t.Fatal("OpenDeviceWithVIDPID(0x8888, 0x0002): got nil device, need non-nil")
	}
	defer dev.Close()
	if err != nil {
		t.Fatalf("OpenDeviceWithVIDPID(0x8888, 0x0002): %v", err)
	}
	cfg, err := dev.Config(cfgNum)
	if err != nil {
		t.Fatalf("%s.Config(1): %v", dev, err)
	}
	defer cfg.Close()
	if got, err := dev.ActiveConfig(); err != nil {
		t.Errorf("%s.ActiveConfig(): got error %v, want nil", dev, err)
	} else if got != cfgNum {
		t.Errorf("%s.ActiveConfig(): got %d, want %d", dev, got, cfgNum)
	}

	intf, err := cfg.Interface(if1Num, alt1Num)
	if err != nil {
		t.Fatalf("%s.Interface(%d, %d): %v", cfg, if1Num, alt1Num, err)
	}
	defer intf.Close()
	got, err := intf.InEndpoint(ep1Num)
	if err != nil {
		t.Fatalf("%s.InEndpoint(%d): got error %v, want nil", intf, ep1Num, err)
	}
	if want := fakeDevices[devIdx].Configs[cfgNum].Interfaces[if1Num].AltSettings[alt1Num].Endpoints[ep1Num]; !reflect.DeepEqual(got.Info, want) {
		t.Errorf("%s.InEndpoint(%d): got %+v, want %+v", intf, ep1Num, got, want)
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

	if err := dev.Reset(); err == nil {
		t.Fatalf("%s.Reset(): got nil, want non nil, because Device is still has an active Config.", dev)
	}

	if err := cfg.Close(); err != nil {
		t.Fatalf("%s.Close(): got error %v, want nil", cfg, err)
	}

	if err := dev.Reset(); err != nil {
		t.Fatalf("%s.Reset(): got error %v, want nil", dev, err)
	}

	if err := dev.Close(); err != nil {
		t.Fatalf("%s.Close(): got error %v, want nil", dev, err)
	}
}
