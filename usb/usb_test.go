package usb_test

import (
	"testing"

	. "github.com/kylelemons/gousb/usb"
	"github.com/kylelemons/gousb/usbid"
)

func TestNoop(t *testing.T) {
	c := NewContext()
	defer c.Close()
	c.Debug(0)
}

func TestEnum(t *testing.T) {
	c := NewContext()
	defer c.Close()
	c.Debug(0)

	cnt := 0
	devs, err := c.ListDevices(func(bus, addr int, desc *Descriptor) bool {
		cnt++
		return true
	})
	defer func() {
		for _, d := range devs {
			d.Close()
		}
	}()
	if err != nil {
		t.Fatalf("list: %s", err)
	}

	if got, want := len(devs), cnt; got != want {
		t.Errorf("len(devs) = %d, want %d", got, want)
	}

	for _, dev := range devs {
		desc, err := dev.Descriptor()
		if err != nil {
			t.Errorf("desc: %s", err)
			continue
		}
		bus := dev.BusNumber()
		addr := dev.Address()

		t.Logf("%03d:%03d %s", bus, addr, usbid.Describe(desc))
		t.Logf("- Protocol: %s", usbid.Classify(desc))

		cfgs, err := dev.Configurations()
		if err != nil {
			t.Errorf("  - configs: %s", err)
			continue
		}

		for _, cfg := range cfgs {
			t.Logf("- %s:", cfg)
			for _, alt := range cfg.Interfaces {
				t.Logf("  --------------")
				for _, iface := range alt {
					t.Logf("  - %s", iface)
					t.Logf("    - %s", usbid.Classify(iface))
					for _, end := range iface.Endpoints {
						t.Logf("    - %s", end)
					}
				}
			}
			t.Logf("  --------------")
		}
	}
}
