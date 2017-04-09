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
	"strings"
	"time"
)

type EndpointInfo struct {
	Number        uint8
	Direction     EndpointDirection
	MaxPacketSize uint32
	TransferType  TransferType
	PollInterval  time.Duration
	IsoSyncType   IsoSyncType
	UsageType     UsageType
}

func (e EndpointInfo) String() string {
	ret := make([]string, 0, 3)
	ret = append(ret, fmt.Sprintf("Endpoint #%d %s (address 0x%02x) %s", e.Number, e.Direction, uint8(e.Number)|uint8(e.Direction), e.TransferType))
	switch e.TransferType {
	case TransferTypeIsochronous:
		ret = append(ret, fmt.Sprintf("- %s %s", e.IsoSyncType, e.UsageType))
	case TransferTypeInterrupt:
		ret = append(ret, fmt.Sprintf("- %s", e.UsageType))
	}
	ret = append(ret, fmt.Sprintf("[%d bytes]", e.MaxPacketSize))
	return strings.Join(ret, " ")
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
	IfClass    Class
	IfSubClass Class
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
