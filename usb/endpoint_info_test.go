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

import "time"

// IN bulk endpoint
var testBulkInEP = EndpointInfo{
	Number:        2,
	Direction:     EndpointDirectionIn,
	MaxPacketSize: 512,
	TransferType:  TransferTypeBulk,
}

var testBulkInSetup = InterfaceSetup{
	Number:    0,
	Alternate: 0,
	IfClass:   ClassVendorSpec,
	Endpoints: []EndpointInfo{testBulkInEP},
}

// OUT iso endpoint
var testIsoOutEP = EndpointInfo{
	Number:        6,
	MaxPacketSize: 3 * 1024,
	TransferType:  TransferTypeIsochronous,
	PollInterval:  125 * time.Microsecond,
	UsageType:     IsoUsageTypeData,
}

var testIsoOutSetup = InterfaceSetup{
	Number:    0,
	Alternate: 0,
	IfClass:   ClassVendorSpec,
	Endpoints: []EndpointInfo{testIsoOutEP},
}
