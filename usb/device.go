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
	"sync"
	"time"
)

var DefaultReadTimeout = 1 * time.Second
var DefaultWriteTimeout = 1 * time.Second
var DefaultControlTimeout = 250 * time.Millisecond //5 * time.Second

type Device struct {
	handle *libusbDevHandle

	// Embed the device information for easy access
	*Descriptor

	// Timeouts
	ReadTimeout    time.Duration
	WriteTimeout   time.Duration
	ControlTimeout time.Duration

	// Claimed interfaces
	lock    *sync.Mutex
	claimed map[uint8]int
}

func newDevice(handle *libusbDevHandle, desc *Descriptor) *Device {
	ifaces := 0
	d := &Device{
		handle:         handle,
		Descriptor:     desc,
		ReadTimeout:    DefaultReadTimeout,
		WriteTimeout:   DefaultWriteTimeout,
		ControlTimeout: DefaultControlTimeout,
		lock:           new(sync.Mutex),
		claimed:        make(map[uint8]int, ifaces),
	}

	return d
}

func (d *Device) Reset() error {
	return libusb.reset(d.handle)
}

func (d *Device) Control(rType, request uint8, val, idx uint16, data []byte) (int, error) {
	return libusb.control(d.handle, d.ControlTimeout, rType, request, val, idx, data)
}

// ActiveConfig returns the config id (not the index) of the active configuration.
// This corresponds to the ConfigInfo.Config field.
func (d *Device) ActiveConfig() (uint8, error) {
	return libusb.getConfig(d.handle)
}

// SetConfig attempts to change the active configuration.
// The cfg provided is the config id (not the index) of the configuration to set,
// which corresponds to the ConfigInfo.Config field.
func (d *Device) SetConfig(cfg uint8) error {
	return libusb.setConfig(d.handle, cfg)
}

// Close the device.
func (d *Device) Close() error {
	if d.handle == nil {
		return fmt.Errorf("usb: double close on device")
	}
	d.lock.Lock()
	defer d.lock.Unlock()
	for iface := range d.claimed {
		libusb.release(d.handle, iface)
	}
	libusb.close(d.handle)
	d.handle = nil
	return nil
}

func (d *Device) OpenEndpoint(cfgNum, ifNum, setNum, epNum uint8) (Endpoint, error) {
	var cfg *ConfigInfo
	for _, c := range d.Configs {
		if c.Config == cfgNum {
			debug.Printf("found conf: %#v\n", c)
			cfg = &c
			break
		}
	}
	if cfg == nil {
		return nil, fmt.Errorf("usb: unknown configuration %02x", cfgNum)
	}

	var intf *InterfaceInfo
	for _, i := range cfg.Interfaces {
		if i.Number == ifNum {
			debug.Printf("found iface: %#v\n", i)
			intf = &i
			break
		}
	}
	if intf == nil {
		return nil, fmt.Errorf("usb: unknown interface %02x", ifNum)
	}

	var setAlternate bool
	var ifs *InterfaceSetup
	for i, s := range intf.Setups {
		if s.Alternate == setNum {
			setAlternate = i != 0
			debug.Printf("found setup: %#v [default: %v]\n", s, !setAlternate)
			ifs = &s
		}
	}
	if ifs == nil {
		return nil, fmt.Errorf("usb: unknown setup %02x", setNum)
	}

	var ep *EndpointInfo
	for _, e := range ifs.Endpoints {
		if e.Address == epNum {
			debug.Printf("found ep %02x in %#v\n", epNum, *ifs)
			ep = &e
		}
	}
	if ep == nil {
		return nil, fmt.Errorf("usb: unknown endpoint %02x", epNum)
	}

	end := newEndpoint(d.handle, *ifs, *ep, d.ReadTimeout, d.WriteTimeout)

	// Set the configuration
	activeConf, err := libusb.getConfig(d.handle)
	if err != nil {
		return nil, fmt.Errorf("usb: getcfg: %s", err)
	}
	if activeConf != cfgNum {
		if err := libusb.setConfig(d.handle, cfgNum); err != nil {
			return nil, fmt.Errorf("usb: setcfg: %s", err)
		}
	}

	// Claim the interface
	if err := libusb.claim(d.handle, ifNum); err != nil {
		return nil, fmt.Errorf("usb: claim: %s", err)
	}

	// Increment the claim count
	d.lock.Lock()
	d.claimed[ifNum]++
	d.lock.Unlock() // unlock immediately because the next calls may block

	// Choose the alternate
	if setAlternate {
		if err := libusb.setAlt(d.handle, ifNum, setNum); err != nil {
			return nil, fmt.Errorf("usb: setalt: %s", err)
		}
	}

	return end, nil
}

func (d *Device) GetStringDescriptor(desc_index int) (string, error) {
	return libusb.getStringDesc(d.handle, desc_index)
}

// SetAutoDetach enables/disables libusb's automatic kernel driver detachment.
// When autodetach is enabled libusb will automatically detach the kernel driver
// on the interface and reattach it when releasing the interface.
// Automatic kernel driver detachment is disabled on newly opened device handles by default.
func (d *Device) SetAutoDetach(autodetach bool) error {
	var autodetachInt int
	switch autodetach {
	case true:
		autodetachInt = 1
	case false:
		autodetachInt = 0
	}
	return libusb.setAutoDetach(d.handle, autodetachInt)
}
