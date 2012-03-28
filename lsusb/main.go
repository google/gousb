package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/kylelemons/gousb/usb"
	"github.com/kylelemons/gousb/usbid"
)

var (
	debug = flag.Int("debug", 0, "libusb debug level (0..3)")
)

func main() {
	flag.Parse()

	// Only one context should be needed for an application.  It should always be closed.
	ctx := usb.NewContext()
	defer ctx.Close()

	// Debugging can be turned on; this shows some of the inner workings of the libusb package.
	ctx.Debug(*debug)

	// ListDevices is used to find the devices to open.
	devs, err := ctx.ListDevices(func(bus, addr int, desc *usb.Descriptor) bool {
		// After inspecting the descriptor, return true or false depending on whether
		// the device is "interesting" or not.  Any descriptor for which true is returned
		// generates a DeviceInfo which is retuned in a slice.
		return true
	})

	// All DeviceInfo returned from ListDevices must be closed.
	defer func() {
		for _, d := range devs {
			d.Close()
		}
	}()

	// ListDevices can occaionally fail, so be sure to check its return value.
	if err != nil {
		log.Fatalf("list: %s", err)
	}

	for _, dev := range devs {
		// The descriptor (which contains useful information) can always be obtained again
		// from a DeviceInfo.
		desc, err := dev.Descriptor()
		if err != nil {
			log.Printf("desc: %s", err)
			continue
		}

		// The DeviceInfo contains the bus number and bus address of the device.
		bus := dev.BusNumber()
		addr := dev.Address()

		// The usbid package can be used to print out human readable information.
		fmt.Printf("%03d:%03d %s\n", bus, addr, usbid.Describe(desc))
		fmt.Printf("  Protocol: %s\n", usbid.Classify(desc))

		// The configurations can be examined from the DeviceInfo, though they can only
		// be set once the device is opened.  All configuration references must be closed,
		// to free up the memory in libusb.
		cfgs, err := dev.Configurations()
		defer func() {
			for _, cfg := range cfgs {
				cfg.Close()
			}
		}()
		if err != nil {
			log.Printf("  - configs: %s", err)
			continue
		}

		// This loop just uses more of the built-in and usbid pretty printing to list
		// the USB devices.
		for _, cfg := range cfgs {
			fmt.Printf("  %s:\n", cfg)
			for _, alt := range cfg.Interfaces {
				fmt.Printf("    --------------\n")
				for _, iface := range alt {
					fmt.Printf("    %s\n", iface)
					fmt.Printf("      %s\n", usbid.Classify(iface))
					for _, end := range iface.Endpoints {
						fmt.Printf("      %s\n", end)
					}
				}
			}
			fmt.Printf("    --------------\n")
		}

		// To actually interact with a device, DevInfo.Open() returns a device handle
		// which can be used to do more I/O with the device and its endpoints.
	}
}
