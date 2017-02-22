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

package usb

// IN bulk endpoint
var testBulkInEP = EndpointInfo{
	Address:       0x82,
	Attributes:    uint8(TRANSFER_TYPE_BULK),
	MaxPacketSize: 512,
	PollInterval:  1,
}

var testBulkInSetup = InterfaceSetup{
	Number:    0,
	Alternate: 0,
	IfClass:   uint8(CLASS_VENDOR_SPEC),
	Endpoints: []EndpointInfo{testBulkInEP},
}

// OUT iso endpoint
var testIsoOutEP = EndpointInfo{
	Address:       0x06,
	Attributes:    uint8(TRANSFER_TYPE_ISOCHRONOUS),
	MaxPacketSize: 3<<11 + 1024,
	MaxIsoPacket:  3 * 1024,
	PollInterval:  1,
}

var testIsoOutSetup = InterfaceSetup{
	Number:    0,
	Alternate: 0,
	IfClass:   uint8(CLASS_VENDOR_SPEC),
	Endpoints: []EndpointInfo{testIsoOutEP},
}
