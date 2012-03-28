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

	logDevice := func(t *testing.T, desc *Descriptor) {
		t.Logf("%03d.%03d %s", desc.Bus, desc.Address, usbid.Describe(desc))
		t.Logf("- Protocol: %s", usbid.Classify(desc))

		for _, cfg := range desc.Configs {
			t.Logf("- %s:", cfg)
			for _, alt := range cfg.Interfaces {
				t.Logf("  --------------")
				for _, iface := range alt {
					t.Logf("  - %s", iface)
					t.Logf("    - %s", usbid.Classify(iface))
					for _, end := range iface.Endpoints {
						t.Logf("    - %s (packet size: %d bytes)", end, end.MaxPacketSize)
					}
				}
			}
			t.Logf("  --------------")
		}
	}

	descs := []*Descriptor{}
	devs, err := c.ListDevices(func(desc *Descriptor) bool {
		logDevice(t, desc)
		descs = append(descs, desc)
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

	if got, want := len(devs), len(descs); got != want {
		t.Fatalf("len(devs) = %d, want %d", got, want)
	}

	for i := range devs {
		if got, want := devs[i].Descriptor, descs[i]; got != want {
			t.Errorf("dev[%d].Descriptor = %p, want %p", i, got, want)
		}
	}
}
