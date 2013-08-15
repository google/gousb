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

package usbid

import (
	"reflect"
	"strings"
	"testing"

	"github.com/kylelemons/gousb/usb"
)

func TestParse(t *testing.T) {
	tests := []struct {
		Input   string
		Vendors map[usb.ID]*Vendor
		Classes map[uint8]*Class
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
		Vendors: map[usb.ID]*Vendor{
			0xabcd: {
				Name: "Vendor One",
				Product: map[usb.ID]*Product{
					0x0123: {Name: "Product One"},
					0x0124: {Name: "Product Two"},
				},
			},
			0xefef: {
				Name: "Vendor Two",
				Product: map[usb.ID]*Product{
					0x0aba: {
						Name: "Product",
						Interface: map[usb.ID]string{
							0x12: "Interface One",
							0x24: "Interface Two",
						},
					},
					0x0abb: {
						Name: "Product",
						Interface: map[usb.ID]string{
							0x12: "Interface",
						},
					},
				},
			},
		},
		Classes: map[uint8]*Class{
			0x00: {
				Name: "(Defined at Interface level)",
			},
			0x01: {
				Name: "Audio",
				SubClass: map[uint8]*SubClass{
					0x01: {Name: "Control Device"},
					0x02: {Name: "Streaming"},
					0x03: {Name: "MIDI Streaming"},
				},
			},
			0x02: {
				Name: "Communications",
				SubClass: map[uint8]*SubClass{
					0x01: {Name: "Direct Line"},
					0x02: {
						Name: "Abstract (modem)",
						Protocol: map[uint8]string{
							0x00: "None",
							0x01: "AT-commands (v.25ter)",
							0x02: "AT-commands (PCCA101)",
							0x03: "AT-commands (PCCA101 + wakeup)",
							0x04: "AT-commands (GSM)",
							0x05: "AT-commands (3G)",
							0x06: "AT-commands (CDMA)",
							0xfe: "Defined by command set descriptor",
							0xff: "Vendor Specific (MSFT RNDIS?)",
						},
					},
				},
			},
		},
	}}

	for idx, test := range tests {
		vendors, classes, err := ParseIDs(strings.NewReader(test.Input))
		if err != nil {
			t.Errorf("%d. ParseIDs: %s", idx, err)
			continue
		}
		if got, want := vendors, test.Vendors; !reflect.DeepEqual(got, want) {
			t.Errorf("%d. Vendor parse mismatch", idx)
			t.Errorf(" - got:  %+v", got)
			t.Errorf(" - want: %+v", want)
			continue
		}
		if got, want := classes, test.Classes; !reflect.DeepEqual(got, want) {
			t.Errorf("%d. Classes parse mismatch", idx)
			t.Errorf(" - got:  %+v", got)
			t.Errorf(" - want: %+v", want)
			continue
		}
	}
}
