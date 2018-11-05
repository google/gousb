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
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/google/gousb"
)

var (
	vidPID    = flag.String("vidpid", "", "VID:PID of the device to which to connect. Exclusive with busaddr flag.")
	busAddr   = flag.String("busaddr", "", "Bus:address of the device to which to connect. Exclusive with vidpid flag.")
	config    = flag.Int("config", 1, "Configuration number to use with the device.")
	iface     = flag.Int("interface", 0, "Interface to use on the device.")
	alternate = flag.Int("alternate", 0, "Alternate setting to use on the interface.")
	endpoint  = flag.Int("endpoint", 1, "Endpoint number to which to connect (without the leading 0x8).")
	debug     = flag.Int("debug", 3, "Debug level for libusb.")
	size      = flag.Int("read_size", 1024, "Number of bytes of data to read in a single transaction.")
	bufSize   = flag.Int("buffer_size", 0, "Number of buffer transfers, for data prefetching.")
	num       = flag.Int("read_num", 0, "Number of read transactions to perform. 0 means infinite.")
	timeout   = flag.Duration("timeout", 0, "Timeout for the command. 0 means infinite.")
)

func parseVIDPID(vidPid string) (gousb.ID, gousb.ID, error) {
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
	return gousb.ID(vid), gousb.ID(pid), nil
}

func parseBusAddr(busAddr string) (int, int, error) {
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
	return int(bus), int(addr), nil
}

type contextReader interface {
	ReadContext(context.Context, []byte) (int, error)
}

func main() {
	flag.Parse()

	// Only one context should be needed for an application.  It should always be closed.
	ctx := gousb.NewContext()
	defer ctx.Close()

	ctx.Debug(*debug)

	var devName string
	var vid, pid gousb.ID
	var bus, addr int
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
	// OpenDevices is used to find the devices to open.
	devs, err := ctx.OpenDevices(func(desc *gousb.DeviceDesc) bool {
		switch {
		case vid == desc.Vendor && pid == desc.Product:
			return true
		case bus == desc.Bus && addr == desc.Address:
			return true
		}
		return false
	})
	// All Devices returned from OpenDevices must be closed.
	defer func() {
		for _, d := range devs {
			d.Close()
		}
	}()

	// OpenDevices can occasionally fail, so be sure to check its return value.
	if err != nil {
		log.Printf("Warning: OpenDevices: %s.", err)
	}
	switch {
	case len(devs) == 0:
		log.Fatal("No matching devices found.")
	case len(devs) > 1:
		log.Printf("Warning: multiple devices found. Using bus %d, addr %d.", devs[0].Desc.Bus, devs[0].Desc.Address)
		for _, d := range devs[1:] {
			d.Close()
		}
		devs = devs[:1]
	}
	dev := devs[0]

	log.Print("Enabling autodetach")
	dev.SetAutoDetach(true)

	log.Printf("Setting configuration %d...", *config)
	cfg, err := dev.Config(*config)
	if err != nil {
		log.Fatalf("dev.Config(%d): %v", *config, err)
	}
	log.Printf("Claiming interface %d (alt setting %d)...", *iface, *alternate)
	intf, err := cfg.Interface(*iface, *alternate)
	if err != nil {
		log.Fatalf("cfg.Interface(%d, %d): %v", *iface, *alternate, err)
	}

	log.Printf("Using endpoint %d...", *endpoint)
	ep, err := intf.InEndpoint(*endpoint)
	if err != nil {
		log.Fatalf("dev.InEndpoint(): %s", err)
	}
	log.Printf("Found endpoint: %s", ep)
	var rdr contextReader = ep
	if *bufSize > 1 {
		log.Print("Creating buffer...")
		s, err := ep.NewStream(*size, *bufSize)
		if err != nil {
			log.Fatalf("ep.NewStream(): %v", err)
		}
		defer s.Close()
		rdr = s
	}

	opCtx := context.Background()
	if *timeout > 0 {
		var done func()
		opCtx, done = context.WithTimeout(opCtx, *timeout)
		defer done()
	}
	buf := make([]byte, *size)
	log.Print("Reading...")
	for i := 0; *num == 0 || i < *num; i++ {
		num, err := rdr.ReadContext(opCtx, buf)
		if err != nil {
			log.Fatalf("Reading from device failed: %v", err)
		}
		os.Stdout.Write(buf[:num])
	}
}
