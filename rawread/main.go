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

// rawread attempts to read from the specified USB device.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/kylelemons/gousb/usb"
	"github.com/kylelemons/gousb/usbid"
)

var (
	vid      = flag.Uint("vid", 0, "VID of the device to which to connect. Exclusive with bus/addr flags.")
	pid      = flag.Uint("pid", 0, "PID of the device to which to connect. Exclusive with bus/addr flags.")
	bus      = flag.Uint("bus", 0, "Bus number for the device to which to connect. Exclusive with vid/pid flags.")
	addr     = flag.Uint("addr", 0, "Address of the device to which to connect. Exclusive with vid/pid flags.")
	config   = flag.Uint("config", 1, "Endpoint to which to connect")
	iface    = flag.Uint("interface", 0, "Endpoint to which to connect")
	setup    = flag.Uint("setup", 0, "Endpoint to which to connect")
	endpoint = flag.Uint("endpoint", 1, "Endpoint to which to connect")
	debug    = flag.Int("debug", 3, "Debug level for libusb")
	size     = flag.Uint("read_size", 1024, "Maximum number of bytes of data to read. Collected will be printed to STDOUT.")
)

func main() {
	flag.Parse()

	// Only one context should be needed for an application.  It should always be closed.
	ctx := usb.NewContext()
	defer ctx.Close()

	ctx.Debug(*debug)

	var devName string
	switch {
	case *vid == 0 && *pid == 0 && *bus == 0 && *addr == 0:
		log.Fatal("You need to specify the device, either through --vid/--pid flags or through --bus/--addr flags.")
	case (*vid > 0 || *pid > 0) && (*bus > 0 || *addr > 0):
		log.Fatal("You can't use --vid/--pid flags at the same time as --bus/--addr.")
	case *vid > 0 || *pid > 0:
		devName = fmt.Sprintf("VID:PID %04x:%04x", *vid, *pid)
	default:
		devName = fmt.Sprintf("bus:addr %d:%d", *bus, *addr)
	}

	log.Printf("Scanning for device %q...", devName)
	// ListDevices is used to find the devices to open.
	devs, err := ctx.ListDevices(func(desc *usb.Descriptor) bool {
		switch {
		case usb.ID(*vid) == desc.Vendor && usb.ID(*pid) == desc.Product:
			return true
		case uint8(*bus) == desc.Bus && uint8(*addr) == desc.Address:
			return true
		}
		return false
	})
	// All Devices returned from ListDevices must be closed.
	defer func() {
		for _, d := range devs {
			d.Close()
		}
	}()

	// ListDevices can occasionally fail, so be sure to check its return value.
	if err != nil {
		log.Printf("Warning: ListDevices: %s.", err)
	}
	switch {
	case len(devs) == 0:
		log.Fatal("No matching devices found.")
	case len(devs) > 1:
		log.Printf("Warning: multiple devices found. Using bus %d, addr %d.", devs[0].Bus, devs[0].Address)
		for _, d := range devs[1:] {
			d.Close()
		}
		devs = devs[:1]
	}
	dev := devs[0]

	// The usbid package can be used to print out human readable information.
	log.Printf("  Protocol: %s\n", usbid.Classify(dev.Descriptor))

	// The configurations can be examined from the Descriptor, though they can only
	// be set once the device is opened.  All configuration references must be closed,
	// to free up the memory in libusb.
	for _, cfg := range dev.Configs {
		// This loop just uses more of the built-in and usbid pretty printing to list
		// the USB devices.
		log.Printf("  %s:\n", cfg)
		for _, alt := range cfg.Interfaces {
			log.Printf("    --------------\n")
			for _, iface := range alt.Setups {
				log.Printf("    %s\n", iface)
				log.Printf("      %s\n", usbid.Classify(iface))
				for _, end := range iface.Endpoints {
					log.Printf("      %s\n", end)
				}
			}
		}
		log.Printf("    --------------\n")
	}

	log.Printf("Connecting to endpoint...")
	ep, err := dev.OpenEndpoint(uint8(*config), uint8(*iface), uint8(*setup), uint8(*endpoint)|uint8(usb.ENDPOINT_DIR_IN))
	if err != nil {
		log.Fatalf("open: %s", err)
	}

	buf := make([]byte, *size)
	num, err := ep.Read(buf)
	if err != nil {
		log.Fatalf("Reading from device failed: %v", err)
	}
	log.Printf("Read %d bytes of data", num)
	os.Stdout.Write(buf[:num])
}
