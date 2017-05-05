// Copyright 2017 the gousb Authors.  All rights reserved.
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

import "github.com/google/gousb"

const testDBPath = "testdata/testdb.txt"

var (
	testDBVendors = map[gousb.ID]*Vendor{
		0xabcd: {
			Name: "Vendor One",
			Product: map[gousb.ID]*Product{
				0x0123: {Name: "Product One"},
				0x0124: {Name: "Product Two"},
			},
		},
		0xefef: {
			Name: "Vendor Two",
			Product: map[gousb.ID]*Product{
				0x0aba: {
					Name: "Product",
					Interface: map[gousb.ID]string{
						0x12: "Interface One",
						0x24: "Interface Two",
					},
				},
				0x0abb: {
					Name: "Product",
					Interface: map[gousb.ID]string{
						0x12: "Interface",
					},
				},
			},
		},
	}
	testDBClasses = map[gousb.Class]*Class{
		0x00: {
			Name: "(Defined at Interface level)",
		},
		0x01: {
			Name: "Audio",
			SubClass: map[gousb.Class]*SubClass{
				0x01: {Name: "Control Device"},
				0x02: {Name: "Streaming"},
				0x03: {Name: "MIDI Streaming"},
			},
		},
		0x02: {
			Name: "Communications",
			SubClass: map[gousb.Class]*SubClass{
				0x01: {Name: "Direct Line"},
				0x02: {
					Name: "Abstract (modem)",
					Protocol: map[gousb.Protocol]string{
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
	}
)
