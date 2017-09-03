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
	"errors"
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
		ep1Addr = 0x86
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

	mfg, err := dev.Manufacturer()
	if err != nil {
		t.Errorf("%s.Manufacturer(): %v", dev, err)
	}
	if mfg != "ACME Industries" {
		t.Errorf("%s.Manufacturer(): %q", dev, mfg)
	}
	prod, err := dev.Product()
	if err != nil {
		t.Errorf("%s.Product(): %v", dev, err)
	}
	if prod != "Fidgety Gadget" {
		t.Errorf("%s.Product(): %q", dev, prod)
	}
	sn, err := dev.SerialNumber()
	if err != nil {
		t.Errorf("%s.SerialNumber(): %v", dev, err)
	}
	if sn != "01234567" {
		t.Errorf("%s.SerialNumber(): %q", dev, sn)
	}

	if err = dev.SetAutoDetach(true); err != nil {
		t.Fatalf("%s.SetAutoDetach(true): %v", dev, err)
	}
	cfg, err := dev.Config(cfgNum)
	if err != nil {
		t.Fatalf("%s.Config(1): %v", dev, err)
	}
	defer cfg.Close()
	if got, err := dev.ActiveConfigNum(); err != nil {
		t.Errorf("%s.ActiveConfigNum(): got error %v, want nil", dev, err)
	} else if got != cfgNum {
		t.Errorf("%s.ActiveConfigNum(): got %d, want %d", dev, got, cfgNum)
	}

	intf, err := cfg.Interface(if1Num, alt1Num)
	if err != nil {
		t.Fatalf("%s.Interface(%d, %d): %v", cfg, if1Num, alt1Num, err)
	}
	defer intf.Close()
	got, err := intf.InEndpoint(ep1Addr)
	if err != nil {
		t.Fatalf("%s.InEndpoint(%d): got error %v, want nil", intf, ep1Addr, err)
	}
	if want := fakeDevices[devIdx].devDesc.Configs[cfgNum].Interfaces[if1Num].AltSettings[alt1Num].Endpoints[ep1Addr]; !reflect.DeepEqual(got.Desc, want) {
		t.Errorf("%s.InEndpoint(%d): got %+v, want %+v", intf, ep1Addr, got, want)
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

	if _, err := dev.Manufacturer(); err == nil {
		t.Errorf("%s.Manufacturer(): expected an error after device is closed", dev)
	}

	if _, err := dev.Config(cfgNum); err == nil {
		t.Fatalf("%s.Config(1): got error nil, want no nil because it is closed", dev)
	}

	if err := dev.Reset(); err == nil {
		t.Fatalf("%s.Reset(): got error nil, want no nil because it is closed", dev)
	}

	if err := dev.SetAutoDetach(false); err == nil {
		t.Fatalf("%s.SetAutoDetach(false): got error nil, want no nil because it is closed", dev)
	}
}

type failDetachLib struct {
	*fakeLibusb
}

func (*failDetachLib) detachKernelDriver(h *libusbDevHandle, i uint8) error {
	if i == 1 {
		return errors.New("test detach fail")
	}
	return nil
}

func TestAutoDetachFailure(t *testing.T) {
	fake, done := newFakeLibusb()
	defer done()
	libusb = &failDetachLib{fake}

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
	dev.SetAutoDetach(true)
	_, err = dev.Config(1)
	if err == nil {
		t.Fatalf("%s.Config(1) got nil, but want no nil because interface fails to detach", dev)
	}
}
