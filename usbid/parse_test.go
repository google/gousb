package usbid

import (
	"reflect"
	"strings"
	"testing"

	"github.com/kylelemons/gousb/usb"
)

func TestParse(t *testing.T) {
	tests := []struct {
		Input  string
		Output map[usb.ID]*Vendor
	}{{
		Input: `
# Skip comment
abcd  Vendor One
	0123  Product One
	0124  Product Two
efef  Vendor Two
	0aba  Product
		12  Interface One
		24  Interface Two
	0abb  Product
		12  Interface

C 00  (Defined at Interface level)
C 01  Audio
	01  Control Device
	02  Streaming
	03  MIDI Streaming
C 02  Communications
	01  Direct Line
	02  Abstract (modem)
		00  None
		01  AT-commands (v.25ter)
		02  AT-commands (PCCA101)
		03  AT-commands (PCCA101 + wakeup)
		04  AT-commands (GSM)
		05  AT-commands (3G)
		06  AT-commands (CDMA)
		fe  Defined by command set descriptor
		ff  Vendor Specific (MSFT RNDIS?)
`,
		Output: map[usb.ID]*Vendor{
			0xabcd: {
				Name: "Vendor One",
				Devices: map[usb.ID]*Device{
					0x0123: {Name: "Product One"},
					0x0124: {Name: "Product Two"},
				},
			},
			0xefef: {
				Name: "Vendor Two",
				Devices: map[usb.ID]*Device{
					0x0aba: {
						Name: "Product",
						Interfaces: map[usb.ID]string{
							0x12: "Interface One",
							0x24: "Interface Two",
						},
					},
					0x0abb: {
						Name: "Product",
						Interfaces: map[usb.ID]string{
							0x12: "Interface",
						},
					},
				},
			},
		},
	}}

	for idx, test := range tests {
		vendors, err := ParseIDs(strings.NewReader(test.Input))
		if err != nil {
			t.Errorf("%d. ParseIDs: %s", idx, err)
			continue
		}
		if got, want := vendors, test.Output; !reflect.DeepEqual(got, want) {
			t.Errorf("%d. Parse mismatch", idx)
			t.Errorf(" - got:  %v", idx, got)
			t.Errorf(" - want: %v", idx, want)
			continue
		}
	}
}
