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
	"strconv"
	"strings"

	"github.com/kylelemons/gousb/usb"
	"github.com/kylelemons/gousb/usbid"
)

var (
	vidPID   = flag.String("vidpid", "", "VID:PID of the device to which to connect. Exclusive with busaddr flag.")
	busAddr  = flag.String("busaddr", "", "Bus:address of the device to which to connect. Exclusive with vidpid flag.")
	config   = flag.Uint("config", 1, "Endpoint to which to connect")
	iface    = flag.Uint("interface", 0, "Endpoint to which to connect")
	setup    = flag.Uint("setup", 0, "Endpoint to which to connect")
	endpoint = flag.Uint("endpoint", 1, "Endpoint to which to connect")
	debug    = flag.Int("debug", 3, "Debug level for libusb")
	size     = flag.Uint("read_size", 1024, "Number of bytes of data to read in a single transaction.")
	num      = flag.Uint("read_num", 0, "Number of read transactions to perform. 0 means infinite.")
)

func parseVIDPID(vidPid string) (usb.ID, usb.ID, error) {
	s := strings.Split(vidPid, ":")
	if len(s) != 2 {
		return 0, 0, fmt.Errorf("want VID:PID, two 32-bit hex numbers separated by colon, e.g. 1d6b:0002")
	}
	vid, err := strconv.ParseUint(s[0], 16, 32)
	if err != nil {
		return 0, 0, fmt.Errorf("VID must be a hexadecimal 32-bit number, e.g. 1d6b")
	}
	pid, err := strconv.ParseUint(s[1], 16, 32)
	if err != nil {
		return 0, 0, fmt.Errorf("PID must be a hexadecimal 32-bit number, e.g. 1d6b")
	}
	return usb.ID(vid), usb.ID(pid), nil
}

func parseBusAddr(busAddr string) (uint8, uint8, error) {
	s := strings.Split(busAddr, ":")
	if len(s) != 2 {
		return 0, 0, fmt.Errorf("want bus:addr, two 8-bit decimal unsigned integers separated by colon, e.g. 1:1")
	}
	bus, err := strconv.ParseUint(s[0], 10, 8)
	if err != nil {
		return 0, 0, fmt.Errorf("bus number must be an 8-bit decimal unsigned integer")
	}
	addr, err := strconv.ParseUint(s[1], 10, 8)
	if err != nil {
		return 0, 0, fmt.Errorf("device address must be an 8-bit decimal unsigned integer")
	}
	return uint8(bus), uint8(addr), nil
}

func main() {
	flag.Parse()

	// Only one context should be needed for an application.  It should always be closed.
	ctx := usb.NewContext()
	defer ctx.Close()

	ctx.Debug(*debug)

	var devName string
	var vid, pid usb.ID
	var bus, addr uint8
	switch {
	case *vidPID == "" && *busAddr == "":
		log.Fatal("You need to specify the device through a --vidpid flag or through a --busaddr flag.")
	case *vidPID != "" && *busAddr != "":
		log.Fatal("You can't use --vidpid flag together with --busaddr. Pick one.")
	case *vidPID != "":
		var err error
		vid, pid, err = parseVIDPID(*vidPID)
		if err != nil {
			log.Fatalf("Invalid value for --vidpid (%q): %v", *vidPID, err)
		}
		devName = fmt.Sprintf("VID:PID %s:%s", vid, pid)
	default:
		var err error
		bus, addr, err = parseBusAddr(*busAddr)
		if err != nil {
			log.Fatalf("Invalid value for --busaddr (%q): %v", *busAddr, err)
		}
		devName = fmt.Sprintf("bus:addr %d:%d", bus, addr)
	}

	log.Printf("Scanning for device %q...", devName)
	// ListDevices is used to find the devices to open.
	devs, err := ctx.ListDevices(func(desc *usb.Descriptor) bool {
		switch {
		case vid == desc.Vendor && pid == desc.Product:
			return true
		case bus == desc.Bus && addr == desc.Address:
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

	log.Printf("Connecting to endpoint %d...", *endpoint)
	ep, err := dev.OpenEndpoint(uint8(*config), uint8(*iface), uint8(*setup), uint8(*endpoint)|uint8(usb.ENDPOINT_DIR_IN))
	if err != nil {
		log.Fatalf("open: %s", err)
	}
	log.Print("Reading...")

	buf := make([]byte, *size)
	for i := uint(0); *num == 0 || i < *num; i++ {
		num, err := ep.Read(buf)
		if err != nil {
			log.Fatalf("Reading from device failed: %v", err)
		}
		os.Stdout.Write(buf[:num])
	}
}
