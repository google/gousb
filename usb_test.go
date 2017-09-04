// Copyright 2013 Google Inc.  All rights reserved.
// Copyright 2016 the gousb Authors.  All rights reserved.
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

func TestOPenDevices(t *testing.T) {
	t.Parallel()
	c := newContextWithImpl(newFakeLibusb())
	defer func() {
		if err := c.Close(); err != nil {
			t.Errorf("Context.Close(): %v", err)
		}
	}()
	c.Debug(0)

	descs := []*DeviceDesc{}
	devs, err := c.OpenDevices(func(desc *DeviceDesc) bool {
		descs = append(descs, desc)
		return true
	})
	defer func() {
		for _, d := range devs {
			d.Close()
		}
	}()
	if err != nil {
		t.Fatalf("OpenDevices(): %s", err)
	}

	if got, want := len(devs), len(fakeDevices); got != want {
		t.Fatalf("len(devs) = %d, want %d (based on num fake devs)", got, want)
	}
	if got, want := len(devs), len(descs); got != want {
		t.Fatalf("len(devs) = %d, want %d (based on num opened devices)", got, want)
	}

	for i := range devs {
		if got, want := devs[i].Desc, descs[i]; got != want {
			t.Errorf("dev[%d].Desc = %p, want %p", i, got, want)
		}
	}
}

func TestOpenDeviceWithVIDPID(t *testing.T) {
	t.Parallel()
	ctx := newContextWithImpl(newFakeLibusb())
	defer func() {
		if err := ctx.Close(); err != nil {
			t.Errorf("Context.Close(): %v", err)
		}
	}()

	for _, d := range []struct {
		vid, pid ID
		exists   bool
	}{
		{0x7777, 0x0003, false},
		{0x8888, 0x0001, false},
		{0x8888, 0x0002, true},
		{0x9999, 0x0001, true},
		{0x9999, 0x0002, false},
	} {
		dev, err := ctx.OpenDeviceWithVIDPID(d.vid, d.pid)
		if (dev != nil) != d.exists {
			t.Errorf("OpenDeviceWithVIDPID(%s/%s): device != nil is %v, want %v", ID(d.vid), ID(d.pid), dev != nil, d.exists)
		}
		if err != nil {
			t.Errorf("OpenDeviceWithVIDPID(%s/%s): got error %v, want nil", ID(d.vid), ID(d.pid), err)
		}
		if dev != nil {
			if dev.Desc.Vendor != ID(d.vid) || dev.Desc.Product != ID(d.pid) {
				t.Errorf("OpenDeviceWithVIDPID(%s/%s): the device returned has VID/PID %s/%s, different from specified in the arguments", ID(d.vid), ID(d.pid), dev.Desc.Vendor, dev.Desc.Product)
			}
			dev.Close()
		}
	}
}
