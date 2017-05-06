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

package gousb

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

// Descriptor structs based on USB 2.0 spec, section 9.6

const (
	deviceDescLen = 18
	configDescLen = 9
	intfDescLen   = 9
	epDescLen     = 7
)

// decoded device descriptor, 18 bytes
type usbDeviceDescriptor struct {
	BLength            uint8  // 0
	BDescriptorType    uint8  // 1
	BCDUSB             uint16 // 2:3
	BDeviceClass       uint8  // 4
	BDeviceSubClass    uint8  // 5
	BDeviceProtocol    uint8  // 6
	BMaxPacketSize0    uint8  // 7
	IDVendor           uint16 // 8:9
	IDProduct          uint16 // 10:11
	BCDDevice          uint16 // 12:13
	IManufacturer      uint8  // 14
	IProduct           uint8  // 15
	ISerialNumber      uint8  // 16
	BNumConfigurations uint8  // 17
}

func deviceDescFromBytes(descBytes []byte) (*DeviceDesc, error) {
	if len(descBytes) < deviceDescLen {
		return nil, fmt.Errorf("device descriptor is %d bytes, got only %d bytes", descBytes, len(descBytes))
	}
	d := &usbDeviceDescriptor{}
	b := bytes.NewReader(descBytes[:deviceDescLen])
	err := binary.Read(b, binary.LittleEndian, d)
	if err != nil {
		return nil, fmt.Errorf("failed to build the device descriptor: %v", err)
	}
	dev := &DeviceDesc{
		Spec:                 BCD(d.BCDUSB),
		Class:                Class(d.BDeviceClass),
		SubClass:             Class(d.BDeviceSubClass),
		Protocol:             Protocol(d.BDeviceProtocol),
		MaxControlPacketSize: int(d.BMaxPacketSize0),
		Vendor:               ID(d.IDVendor),
		Product:              ID(d.IDProduct),
		Device:               BCD(d.BCDDevice),
		Configs:              make(map[int]ConfigDesc, d.BNumConfigurations),
	}
	return dev, nil
}

// decoded configuration descriptor, 9 bytes followed by bNumInterfaces interface descriptors
type usbConfigDescriptor struct {
	bLength             uint8  // 0
	bDescriptorType     uint8  // 1
	wTotalLength        uint16 // 2:3
	bNumInterfaces      uint8  // 4
	bConfigurationValue uint8  // 5
	iConfiguration      uint8  // 6
	bmAttributes        uint8  // 7
	bMaxPower           uint8  // 8
}

// decoded interface descriptor, 9 bytes followed by bNumEndpoints endpoint descriptors.
type usbInterfaceDescriptor struct {
	bLength            uint8 // 0
	bDescriptorType    uint8 // 1
	bInterfaceNumber   uint8 // 2
	bAlternateSetting  uint8 // 3
	bNumEndpoints      uint8 // 4
	bInterfaceClass    uint8 // 5
	bInterfaceSubClass uint8 // 6
	bInterfaceProtocol uint8 // 7
	iInterface         uint8 // 8
}

// decoded endpoint descriptor, 7 bytes.
type usbEndpointDescriptor struct {
	bLength          uint8  // 0
	bDescriptorType  uint8  // 1
	bEndpointAddress uint8  // 2
	bmAttributes     uint8  // 3
	wMaxPacketSize   uint16 // 4:5
	bInterval        uint8  // 6
}
