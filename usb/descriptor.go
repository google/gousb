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

type Descriptor struct {
	// Bus information
	Bus     uint8 // The bus on which the device was detected
	Address uint8 // The address of the device on the bus

	// Version information
	Spec   BCD // USB Specification Release Number
	Device BCD // The device version

	// Product information
	Vendor  ID // The Vendor identifer
	Product ID // The Product identifier

	// Protocol information
	Class    uint8 // The class of this device
	SubClass uint8 // The sub-class (within the class) of this device
	Protocol uint8 // The protocol (within the sub-class) of this device

	// Configuration information
	Configs []ConfigInfo
}
