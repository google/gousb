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

// #include <libusb.h>
import "C"
import "strconv"

type Class uint8

const (
	ClassPerInterface Class = C.LIBUSB_CLASS_PER_INTERFACE
	ClassAudio        Class = C.LIBUSB_CLASS_AUDIO
	ClassComm         Class = C.LIBUSB_CLASS_COMM
	ClassHID          Class = C.LIBUSB_CLASS_HID
	ClassPrinter      Class = C.LIBUSB_CLASS_PRINTER
	ClassPTP          Class = C.LIBUSB_CLASS_PTP
	ClassMassStorage  Class = C.LIBUSB_CLASS_MASS_STORAGE
	ClassHub          Class = C.LIBUSB_CLASS_HUB
	ClassData         Class = C.LIBUSB_CLASS_DATA
	ClassWireless     Class = C.LIBUSB_CLASS_WIRELESS
	ClassApplication  Class = C.LIBUSB_CLASS_APPLICATION
	ClassVendorSpec   Class = C.LIBUSB_CLASS_VENDOR_SPEC
)

var classDescription = map[Class]string{
	ClassPerInterface: "per-interface",
	ClassAudio:        "audio",
	ClassComm:         "communications",
	ClassHID:          "human interface device",
	ClassPrinter:      "printer dclass",
	ClassPTP:          "picture transfer protocol",
	ClassMassStorage:  "mass storage",
	ClassHub:          "hub",
	ClassData:         "data",
	ClassWireless:     "wireless",
	ClassApplication:  "application",
	ClassVendorSpec:   "vendor-specific",
}

func (c Class) String() string {
	if d, ok := classDescription[c]; ok {
		return d
	}
	return strconv.Itoa(int(c))
}

type DescriptorType uint8

const (
	DescriptorTypeDevice    DescriptorType = C.LIBUSB_DT_DEVICE
	DescriptorTypeConfig    DescriptorType = C.LIBUSB_DT_CONFIG
	DescriptorTypeString    DescriptorType = C.LIBUSB_DT_STRING
	DescriptorTypeInterface DescriptorType = C.LIBUSB_DT_INTERFACE
	DescriptorTypeEndpoint  DescriptorType = C.LIBUSB_DT_ENDPOINT
	DescriptorTypeHID       DescriptorType = C.LIBUSB_DT_HID
	DescriptorTypeReport    DescriptorType = C.LIBUSB_DT_REPORT
	DescriptorTypePhysical  DescriptorType = C.LIBUSB_DT_PHYSICAL
	DescriptorTypeHub       DescriptorType = C.LIBUSB_DT_HUB
)

var descriptorTypeDescription = map[DescriptorType]string{
	DescriptorTypeDevice:    "device",
	DescriptorTypeConfig:    "configuration",
	DescriptorTypeString:    "string",
	DescriptorTypeInterface: "interface",
	DescriptorTypeEndpoint:  "endpoint",
	DescriptorTypeHID:       "HID",
	DescriptorTypeReport:    "HID report",
	DescriptorTypePhysical:  "physical",
	DescriptorTypeHub:       "hub",
}

func (dt DescriptorType) String() string {
	return descriptorTypeDescription[dt]
}

type EndpointDirection uint8

const (
	EndpointNumMask                         = 0x0f
	EndpointDirectionMask                   = 0x80
	EndpointDirectionIn   EndpointDirection = C.LIBUSB_ENDPOINT_IN
	EndpointDirectionOut  EndpointDirection = C.LIBUSB_ENDPOINT_OUT
)

var endpointDirectionDescription = map[EndpointDirection]string{
	EndpointDirectionIn:  "IN",
	EndpointDirectionOut: "OUT",
}

func (ed EndpointDirection) String() string {
	return endpointDirectionDescription[ed]
}

type TransferType uint8

const (
	TransferTypeControl     TransferType = C.LIBUSB_TRANSFER_TYPE_CONTROL
	TransferTypeIsochronous TransferType = C.LIBUSB_TRANSFER_TYPE_ISOCHRONOUS
	TransferTypeBulk        TransferType = C.LIBUSB_TRANSFER_TYPE_BULK
	TransferTypeInterrupt   TransferType = C.LIBUSB_TRANSFER_TYPE_INTERRUPT
	TransferTypeMask                     = 0x03
)

var transferTypeDescription = map[TransferType]string{
	TransferTypeControl:     "control",
	TransferTypeIsochronous: "isochronous",
	TransferTypeBulk:        "bulk",
	TransferTypeInterrupt:   "interrupt",
}

func (tt TransferType) String() string {
	return transferTypeDescription[tt]
}

type IsoSyncType uint8

const (
	IsoSyncTypeNone     IsoSyncType = C.LIBUSB_ISO_SYNC_TYPE_NONE << 2
	IsoSyncTypeAsync    IsoSyncType = C.LIBUSB_ISO_SYNC_TYPE_ASYNC << 2
	IsoSyncTypeAdaptive IsoSyncType = C.LIBUSB_ISO_SYNC_TYPE_ADAPTIVE << 2
	IsoSyncTypeSync     IsoSyncType = C.LIBUSB_ISO_SYNC_TYPE_SYNC << 2
	IsoSyncTypeMask                 = 0x0C
)

var isoSyncTypeDescription = map[IsoSyncType]string{
	IsoSyncTypeNone:     "unsynchronized",
	IsoSyncTypeAsync:    "asynchronous",
	IsoSyncTypeAdaptive: "adaptive",
	IsoSyncTypeSync:     "synchronous",
}

func (ist IsoSyncType) String() string {
	return isoSyncTypeDescription[ist]
}

type UsageType uint8

const (
	// Note: USB3.0 defines usage type for both isochronous and interrupt
	// endpoints, with the same constants representing different usage types.
	// UsageType constants do not correspond to bmAttribute values.
	UsageTypeMask                = 0x30
	UsageTypeUndefined UsageType = iota
	IsoUsageTypeData
	IsoUsageTypeFeedback
	IsoUsageTypeImplicit
	InterruptUsageTypePeriodic
	InterruptUsageTypeNotification
)

var usageTypeDescription = map[UsageType]string{
	UsageTypeUndefined:             "undefined usage",
	IsoUsageTypeData:               "data",
	IsoUsageTypeFeedback:           "feedback",
	IsoUsageTypeImplicit:           "implicit data",
	InterruptUsageTypePeriodic:     "periodic",
	InterruptUsageTypeNotification: "notification",
}

func (ut UsageType) String() string {
	return usageTypeDescription[ut]
}

type RequestType uint8

const (
	RequestTypeStandard = C.LIBUSB_REQUEST_TYPE_STANDARD
	RequestTypeClass    = C.LIBUSB_REQUEST_TYPE_CLASS
	RequestTypeVendor   = C.LIBUSB_REQUEST_TYPE_VENDOR
	RequestTypeReserved = C.LIBUSB_REQUEST_TYPE_RESERVED
)

var requestTypeDescription = map[RequestType]string{
	RequestTypeStandard: "standard",
	RequestTypeClass:    "class",
	RequestTypeVendor:   "vendor",
	RequestTypeReserved: "reserved",
}

func (rt RequestType) String() string {
	return requestTypeDescription[rt]
}

type DeviceSpeed int

const (
	DeviceSpeedUnknown DeviceSpeed = C.LIBUSB_SPEED_UNKNOWN
	DeviceSpeedLow     DeviceSpeed = C.LIBUSB_SPEED_LOW
	DeviceSpeedFull    DeviceSpeed = C.LIBUSB_SPEED_FULL
	DeviceSpeedHigh    DeviceSpeed = C.LIBUSB_SPEED_HIGH
	DeviceSpeedSuper   DeviceSpeed = C.LIBUSB_SPEED_SUPER
)

var deviceSpeedDescription = map[DeviceSpeed]string{
	DeviceSpeedUnknown: "unknown",
	DeviceSpeedLow:     "low",
	DeviceSpeedFull:    "full",
	DeviceSpeedHigh:    "high",
	DeviceSpeedSuper:   "super",
}

func (s DeviceSpeed) String() string {
	return deviceSpeedDescription[s]
}
