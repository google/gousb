package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/kylelemons/gousb/usb"
	"github.com/kylelemons/gousb/usbid"
)

var (
	ctx = usb.NewContext()
)

var (
	debug = flag.Int("debug", 0, "libusb debug level (0..3)")
)

func main() {
	flag.Parse()
	ctx.Debug(*debug)

	devs, err := ctx.ListDevices(func(bus, addr int, desc *usb.Descriptor) bool { return true })
	defer func() {
		for _, d := range devs {
			d.Close()
		}
	}()
	if err != nil {
		log.Fatalf("list: %s", err)
	}

	for _, dev := range devs {
		desc, err := dev.Descriptor()
		if err != nil {
			log.Printf("desc: %s", err)
			continue
		}
		bus := dev.BusNumber()
		addr := dev.Address()

		fmt.Printf("%03d:%03d %s\n", bus, addr, usbid.Describe(desc))
		fmt.Printf("  Protocol: %s\n", usbid.Classify(desc))

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
	}
}
