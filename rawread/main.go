// Copyright 2013 Google Inc.  All rights reserved.
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

// rawread attempts to read from the specified USB device.
package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/kylelemons/gousb/usb"
	"github.com/kylelemons/gousb/usbid"
)

var (
	device   = flag.String("device", "vend:prod", "Device to which to connect")
	config   = flag.Int("config", 1, "Endpoint to which to connect")
	iface    = flag.Int("interface", 0, "Endpoint to which to connect")
	setup    = flag.Int("setup", 0, "Endpoint to which to connect")
	endpoint = flag.Int("endpoint", 1, "Endpoint to which to connect")
	debug    = flag.Int("debug", 3, "Debug level for libusb")
)

func main() {
	flag.Parse()

	// Only one context should be needed for an application.  It should always be closed.
	ctx := usb.NewContext()
	defer ctx.Close()

	ctx.Debug(*debug)

	log.Printf("Scanning for device %q...", *device)

	// ListDevices is used to find the devices to open.
	devs, err := ctx.ListDevices(func(desc *usb.Descriptor) bool {
		if fmt.Sprintf("%s:%s", desc.Vendor, desc.Product) != *device {
			return false
		}

		// The usbid package can be used to print out human readable information.
		fmt.Printf("  Protocol: %s\n", usbid.Classify(desc))

		// The configurations can be examined from the Descriptor, though they can only
		// be set once the device is opened.  All configuration references must be closed,
		// to free up the memory in libusb.
		for _, cfg := range desc.Configs {
			// This loop just uses more of the built-in and usbid pretty printing to list
			// the USB devices.
			fmt.Printf("  %s:\n", cfg)
			for _, alt := range cfg.Interfaces {
				fmt.Printf("    --------------\n")
				for _, iface := range alt.Setups {
					fmt.Printf("    %s\n", iface)
					fmt.Printf("      %s\n", usbid.Classify(iface))
					for _, end := range iface.Endpoints {
						fmt.Printf("      %s\n", end)
					}
				}
			}
			fmt.Printf("    --------------\n")
		}

		return true
	})

	// All Devices returned from ListDevices must be closed.
	defer func() {
		for _, d := range devs {
			d.Close()
		}
	}()

	// ListDevices can occaionally fail, so be sure to check its return value.
	if err != nil {
		log.Fatalf("list: %s", err)
	}

	if len(devs) == 0 {
		log.Fatalf("no devices found")
	}

	dev := devs[0]

	log.Printf("Connecting to endpoint...")
	log.Printf("- %#v", dev.Descriptor)
	ep, err := dev.OpenEndpoint(uint8(*config), uint8(*iface), uint8(*setup), uint8(*endpoint)|uint8(usb.ENDPOINT_DIR_IN))
	if err != nil {
		log.Fatalf("open: %s", err)
	}
	_ = ep
}
