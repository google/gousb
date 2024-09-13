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
	t.Parallel()
	const (
		devIdx  = 1
		cfgNum  = 1
		if1Num  = 1
		alt1Num = 1
		ep1Addr = 0x86
		alt2Num = 0
		if2Num  = 0
	)

	c := newContextWithImpl(newFakeLibusb())
	defer func() {
		if err := c.Close(); err != nil {
			t.Errorf("Context.Close: %v", err)
		}
	}()

	dev, err := c.OpenDeviceWithVIDPID(0x8888, 0x0002)
	if dev == nil {
		t.Fatal("OpenDeviceWithVIDPID(0x8888, 0x0002): got nil device, need non-nil")
	}
	defer dev.Close()
	if err != nil {
		t.Fatalf("OpenDeviceWithVIDPID(0x8888, 0x0002): %v", err)
	}

	if mfg, err := dev.Manufacturer(); err != nil {
		t.Errorf("%s.Manufacturer(): error %v", dev, err)
	} else if want := "ACME Industries"; mfg != want {
		t.Errorf("%s.Manufacturer(): %q, want %q", dev, mfg, want)
	}
	if prod, err := dev.Product(); err != nil {
		t.Errorf("%s.Product(): error %v", dev, err)
	} else if want := "Fidgety Gadget"; prod != want {
		t.Errorf("%s.Product(): %q, want %q", dev, prod, want)
	}
	if sn, err := dev.SerialNumber(); err != nil {
		t.Errorf("%s.SerialNumber(): error %v", dev, err)
	} else if want := "01234567"; sn != want {
		t.Errorf("%s.SerialNumber(): %q, want %q", dev, sn, want)
	}

	if got, err := dev.ConfigDescription(1); err != nil {
		t.Errorf("%s.ConfigDescription(1): %v", dev, err)
	} else if want := "Weird configuration"; got != want {
		t.Errorf("%s.ConfigDescription(1): %q, want %q", dev, got, want)
	}
	if got, err := dev.ConfigDescription(2); err == nil {
		t.Errorf("%s.ConfigDescription(2): %q, want error", dev, got)
	}

	for _, tc := range []struct {
		intf, alt int
		want      string
	}{
		{0, 0, "Boring setting"},
		{1, 0, "Fast streaming"},
		{1, 1, "Slower streaming"},
		{1, 2, ""},
		{3, 2, "Interface for https://github.com/google/gousb/issues/65"},
	} {
		if got, err := dev.InterfaceDescription(1, tc.intf, tc.alt); err != nil {
			t.Errorf("%s.InterfaceDescription(1, %d, %d): %v", dev, tc.intf, tc.alt, err)
		} else if got != tc.want {
			t.Errorf("%s.InterfaceDescription(1, %d, %d): %q, want %q", dev, tc.intf, tc.alt, got, tc.want)
		}
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

	if _, err := cfg.Interface(if1Num, 0); err == nil {
		t.Fatalf("%s.Interface(1, 0): got nil, want non nil, because Interface 1 is already claimed.", cfg)
	}

	// intf2 is interface #0, not claimed yet.
	intf2, err := cfg.Interface(if2Num, alt2Num)
	if err != nil {
		t.Fatalf("%s.Interface(0, 0): got %v, want nil", cfg, err)
	}

	if err := cfg.Close(); err == nil {
		t.Fatalf("%s.Close(): got nil, want non nil, because the Interfaces #1/2 was not released.", cfg)
	}
	if err := dev.Close(); err == nil {
		t.Fatalf("%s.Close(): got nil, want non nil, because the Config was not released.", cfg)
	}

	intf.Close()
	if err := cfg.Close(); err == nil {
		t.Fatalf("%s.Close(): got nil, want non nil, because the Interface #2 was not released.", cfg)
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

func TestInterfaceDescriptionError(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name           string
		cfg, intf, alt int
	}{
		{"no config", 2, 1, 1},
		{"no interface", 1, 3, 1},
		{"no alt setting", 1, 1, 5},
	} {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			c := newContextWithImpl(newFakeLibusb())
			defer func() {
				if err := c.Close(); err != nil {
					t.Errorf("Context.Close(): %v", err)
				}
			}()
			dev, err := c.OpenDeviceWithVIDPID(0x8888, 0x0002)
			if dev == nil {
				t.Fatal("OpenDeviceWithVIDPID(0x8888, 0x0002): got nil device, need non-nil")
			}
			defer dev.Close()
			if err != nil {
				t.Fatalf("OpenDeviceWithVIDPID(0x8888, 0x0002): %v", err)
			}
			if desc, err := dev.InterfaceDescription(tc.cfg, tc.intf, tc.alt); err == nil {
				t.Errorf("%s.InterfaceDescriptor(%d, %d, %d): %q, want error", dev, tc.cfg, tc.intf, tc.alt, desc)
			}
		})
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
	t.Parallel()
	fake := newFakeLibusb()
	c := newContextWithImpl(&failDetachLib{fake})
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

func TestInterface(t *testing.T) {
	t.Parallel()
	c := newContextWithImpl(newFakeLibusb())
	defer func() {
		if err := c.Close(); err != nil {
			t.Errorf("Context.Close: %v", err)
		}
	}()

	for _, tc := range []struct {
		name                  string
		vid, pid              ID
		cfgNum, ifNum, altNum int
	}{{
		name:   "simple",
		vid:    0x8888,
		pid:    0x0002,
		cfgNum: 1,
		ifNum:  0,
		altNum: 0,
	}, {
		name:   "alt_setting",
		vid:    0x8888,
		pid:    0x0002,
		cfgNum: 1,
		ifNum:  1,
		altNum: 1,
	}, {
		name:   "noncontiguous_interfaces",
		vid:    0x8888,
		pid:    0x0002,
		cfgNum: 1,
		ifNum:  3,
		altNum: 2,
	}} {
		t.Run(tc.name, func(t *testing.T) {
			dev, err := c.OpenDeviceWithVIDPID(0x8888, 0x0002)
			if err != nil {
				t.Fatalf("OpenDeviceWithVIDPID(0x8888, 0x0002): %v", err)
			}
			if dev == nil {
				t.Fatal("OpenDeviceWithVIDPID(0x8888, 0x0002): got nil device, need non-nil")
			}
			defer dev.Close()
			cfg, err := dev.Config(tc.cfgNum)
			if err != nil {
				t.Fatalf("%s.Config(%d): %v", dev, tc.cfgNum, err)
			}
			defer cfg.Close()
			intf, err := cfg.Interface(tc.ifNum, tc.altNum)
			if err != nil {
				t.Fatalf("%s.Interface(%d, %d): %v", cfg, tc.ifNum, tc.altNum, err)
			}
			intf.Close()
		})
	}
}
