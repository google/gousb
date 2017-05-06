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

import (
	"fmt"
	"log"
	"testing"
)

func TestListDevices(t *testing.T) {
	_, done := newFakeLibusb()
	defer done()

	c := NewContext()
	defer c.Close()
	c.Debug(0)

	descs := []*Descriptor{}
	devs, err := c.ListDevices(func(desc *Descriptor) bool {
		descs = append(descs, desc)
		return true
	})
	defer func() {
		for _, d := range devs {
			d.Close()
		}
	}()
	if err != nil {
		t.Fatalf("ListDevices(): %s", err)
	}

	if got, want := len(devs), len(fakeDevices); got != want {
		t.Fatalf("len(devs) = %d, want %d (based on num fake devs)", got, want)
	}
	if got, want := len(devs), len(descs); got != want {
		t.Fatalf("len(devs) = %d, want %d (based on num opened devices)", got, want)
	}

	for i := range devs {
		if got, want := devs[i].Descriptor, descs[i]; got != want {
			t.Errorf("dev[%d].Descriptor = %p, want %p", i, got, want)
		}
	}
}

func TestOpenDeviceWithVIDPID(t *testing.T) {
	_, done := newFakeLibusb()
	defer done()

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
		c := NewContext()
		defer c.Close()
		dev, err := c.OpenDeviceWithVIDPID(d.vid, d.pid)
		if (dev != nil) != d.exists {
			t.Errorf("OpenDeviceWithVIDPID(%s/%s): device != nil is %v, want %v", ID(d.vid), ID(d.pid), dev != nil, d.exists)
		}
		if err != nil {
			t.Errorf("OpenDeviceWithVIDPID(%s/%s): got error %v, want nil", ID(d.vid), ID(d.pid), err)
		}
		if dev != nil {
			if dev.Descriptor.Vendor != ID(d.vid) || dev.Descriptor.Product != ID(d.pid) {
				t.Errorf("OpenDeviceWithVIDPID(%s/%s): the device returned has VID/PID %s/%s, different from specified in the arguments", ID(d.vid), ID(d.pid), dev.Descriptor.Vendor, dev.Descriptor.Product)
			}
			dev.Close()
		}
	}
}

// This examples demonstrates the use of a few convenience functions that
// can be used in simple situations and with simple devices.
// It opens a device with a given VID/PID,
// claims the default interface (use the same config as currently active,
// interface 0, alternate setting 0) and tries to write 5 bytes of data
// to endpoint number 7.
func Example_simple() {
	// Initialize a new Context.
	ctx := NewContext()
	defer ctx.Close()

	// Open any device with a given VID/PID using a convenience function.
	dev, err := ctx.OpenDeviceWithVIDPID(0x046d, 0xc526)
	if err != nil {
		log.Fatalf("Could not open a device: %v", err)
	}
	defer dev.Close()

	// Claim the default interface using a convenience function.
	// The default interface is always #0 alt #0 in the currently active
	// config.
	intf, err := dev.DefaultInterface()
	if err != nil {
		log.Fatalf("%s.DefaultInterface(): %v", dev, err)
	}
	defer intf.Close()

	// Open an OUT endpoint.
	ep, err := intf.OutEndpoint(7)
	if err != nil {
		log.Fatalf("%s.OutEndpoint(7): %v", intf, err)
	}

	// Generate some data to write.
	data := make([]byte, 5)
	for i := range data {
		data[i] = byte(i)
	}

	// Write data to the USB device.
	numBytes, err := ep.Write(data)
	if numBytes != 5 {
		log.Fatalf("%s.Write([5]): only %d bytes written, returned error is %v", numBytes, err)
	}
	fmt.Println("5 bytes successfuly sent to the endpoint")
}

// This example demostrates the full API for accessing endpoints.
// It opens a device with a known VID/PID, switches the device to
// configuration #2, in that configuration it opens (claims) interface #3 with alternate setting #0.
// Within that interface setting it opens an IN endpoint number 6 and an OUT endpoint number 5, then starts copying
// data between them,
func Example_complex() {
	// Initialize a new Context.
	ctx := NewContext()
	defer ctx.Close()

	// Iterate through available Devices, finding all that match a known VID/PID.
	vid, pid := ID(0x04f2), ID(0xb531)
	devs, err := ctx.ListDevices(func(desc *Descriptor) bool {
		// this function is called for every device present.
		// Returning true means the device should be opened.
		return desc.Vendor == vid && desc.Product == pid
	})
	// All returned devices are now open and will need to be closed.
	for _, d := range devs {
		defer d.Close()
	}
	if err != nil {
		log.Fatalf("ListDevices(): %v", err)
	}
	if len(devs) == 0 {
		log.Fatalf("no devices found matching VID %s and PID %s", vid, pid)
	}

	// Pick the first device found.
	dev := devs[0]

	// Switch the configuration to #2.
	cfg, err := dev.Config(2)
	if err != nil {
		log.Fatalf("%s.Config(2): %v", dev, err)
	}
	defer cfg.Close()

	// In the config #2, claim interface #3 with alt setting #0.
	intf, err := cfg.Interface(3, 0)
	if err != nil {
		log.Fatalf("%s.Interface(3, 0): %v", cfg, err)
	}
	defer intf.Close()

	// In this interface open endpoint #6 for reading.
	epIn, err := intf.InEndpoint(6)
	if err != nil {
		log.Fatalf("%s.InEndpoint(6): %v", intf, err)
	}

	// And in the same interface open endpoint #5 for writing.
	epOut, err := intf.OutEndpoint(5)
	if err != nil {
		log.Fatalf("%s.OutEndpoint(5): %v", intf, err)
	}

	// Buffer large enough for 10 USB packets from endpoint 6.
	buf := make([]byte, 10*epIn.Info.MaxPacketSize)
	total := 0
	// Repeat the read/write cycle 10 times.
	for i := 0; i < 10; i++ {
		// readBytes might be smaller than the buffer size. readBytes might be greater than zero even if err is not nil.
		readBytes, err := epIn.Read(buf)
		if err != nil {
			fmt.Println("Read returned an error:", err)
		}
		if readBytes == 0 {
			log.Fatalf("IN endpoint 6 returned 0 bytes of data.")
		}
		// writeBytes might be smaller than the buffer size if an error occured. writeBytes might be greater than zero even if err is not nil.
		writeBytes, err := epOut.Write(buf[:readBytes])
		if err != nil {
			fmt.Println("Write returned an error:", err)
		}
		if writeBytes != readBytes {
			log.Fatalf("IN endpoint 5 received only %d bytes of data out of %d sent", writeBytes, readBytes)
		}
		total += writeBytes
	}
	fmt.Printf("Total number of bytes copied: %d\n", total)
}
