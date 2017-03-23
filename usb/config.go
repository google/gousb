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

package usb

import (
	"fmt"
)

type EndpointInfo struct {
	Address       uint8
	Attributes    uint8
	MaxPacketSize uint16
	MaxIsoPacket  uint32
	PollInterval  uint8
	RefreshRate   uint8
	SynchAddress  uint8
}

func (e EndpointInfo) Number() int {
	return int(e.Address) & ENDPOINT_NUM_MASK
}

func (e EndpointInfo) TransferType() TransferType {
	return TransferType(e.Attributes) & TRANSFER_TYPE_MASK
}

func (e EndpointInfo) Direction() EndpointDirection {
	return EndpointDirection(e.Address) & ENDPOINT_DIR_MASK
}

func (e EndpointInfo) String() string {
	return fmt.Sprintf("Endpoint #%d %-3s %s - %s %s [%d %d]",
		e.Number(), e.Direction(), e.TransferType(),
		IsoSyncType(e.Attributes)&ISO_SYNC_TYPE_MASK,
		IsoUsageType(e.Attributes)&ISO_USAGE_TYPE_MASK,
		e.MaxPacketSize, e.MaxIsoPacket,
	)
}

type InterfaceInfo struct {
	Number uint8
	Setups []InterfaceSetup
}

func (i InterfaceInfo) String() string {
	return fmt.Sprintf("Interface %02x (%d setups)", i.Number, len(i.Setups))
}

type InterfaceSetup struct {
	Number     uint8
	Alternate  uint8
	IfClass    uint8
	IfSubClass uint8
	IfProtocol uint8
	Endpoints  []EndpointInfo
}

func (a InterfaceSetup) String() string {
	return fmt.Sprintf("Interface %02x Setup %02x", a.Number, a.Alternate)
}

type ConfigInfo struct {
	Config     uint8
	Attributes uint8
	MaxPower   uint8
	Interfaces []InterfaceInfo
}

func (c ConfigInfo) String() string {
	return fmt.Sprintf("Config %02x", c.Config)
}
