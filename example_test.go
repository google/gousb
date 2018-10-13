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

package gousb_test

import (
	"fmt"
	"log"

	"github.com/google/gousb"
)

// This examples demonstrates the use of a few convenience functions that
// can be used in simple situations and with simple devices.
// It opens a device with a given VID/PID,
// claims the default interface (use the same config as currently active,
// interface 0, alternate setting 0) and tries to write 5 bytes of data
// to endpoint number 7.
func Example_simple() {
	// Initialize a new Context.
	ctx := gousb.NewContext()
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
	intf, done, err := dev.DefaultInterface()
	if err != nil {
		log.Fatalf("%s.DefaultInterface(): %v", dev, err)
	}
	defer done()

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
		log.Fatalf("%s.Write([5]): only %d bytes written, returned error is %v", ep, numBytes, err)
	}
	fmt.Println("5 bytes successfully sent to the endpoint")
}

// This example demostrates the full API for accessing endpoints.
// It opens a device with a known VID/PID, switches the device to
// configuration #2, in that configuration it opens (claims) interface #3 with alternate setting #0.
// Within that interface setting it opens an IN endpoint number 6 and an OUT endpoint number 5, then starts copying
// data between them,
func Example_complex() {
	// Initialize a new Context.
	ctx := gousb.NewContext()
	defer ctx.Close()

	// Iterate through available Devices, finding all that match a known VID/PID.
	vid, pid := gousb.ID(0x04f2), gousb.ID(0xb531)
	devs, err := ctx.OpenDevices(func(desc *gousb.DeviceDesc) bool {
		// this function is called for every device present.
		// Returning true means the device should be opened.
		return desc.Vendor == vid && desc.Product == pid
	})
	// All returned devices are now open and will need to be closed.
	for _, d := range devs {
		defer d.Close()
	}
	if err != nil {
		log.Fatalf("OpenDevices(): %v", err)
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
	buf := make([]byte, 10*epIn.Desc.MaxPacketSize)
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
		// writeBytes might be smaller than the buffer size if an error occurred. writeBytes might be greater than zero even if err is not nil.
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
