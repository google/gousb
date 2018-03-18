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

package gousb

type HotplugEvent interface {
	// Type returns the event's type (HotplugEventDeviceArrived or HotplugEventDeviceLeft).
	Type() HotplugEventType
	// IsEnumerated returns true if the device was already plugged in when the callback was registered.
	IsEnumerated() bool
	// DeviceDesc returns the device's descriptor.
	DeviceDesc() (*DeviceDesc, error)
	// Open opens the device.
	Open() (*Device, error)
	// Deregister deregisters the callback registration after the callback function returns.
	Deregister()
}

// RegisterHotplug registers a hotplug callback function.
// The callback will receive arrive events for all currently plugged in devices.
// These events will return true from IsEnumerated().
// Note that events are delivered concurrently. You may receive arrive and leave events
// concurrently with enumerated arrive events. You may also receive arrive events twice
// for the same device, and you may receive a leave event for a device for which
// you never received an arrive event.
func (c *Context) RegisterHotplug(fn func(HotplugEvent)) (func(), error) {
	dereg, err := c.libusb.registerHotplugCallback(c.ctx, HotplugEventAny, false, hotplugMatchAny, hotplugMatchAny, hotplugMatchAny, func(ctx *libusbContext, dev *libusbDevice, eventType HotplugEventType) bool {
		desc, err := c.libusb.getDeviceDesc(dev)
		e := &hotplugEvent{
			eventType:  eventType,
			ctx:        c,
			dev:        dev,
			desc:       desc,
			err:        err,
			enumerated: false,
		}
		fn(e)
		return e.deregister
	})
	if err != nil {
		return nil, err
	}
	// enumerate devices
	// this is done in gousb to properly support cancellation and to distinguish enumerated devices
	list, err := c.libusb.getDevices(c.ctx)
	if err != nil {
		dereg()
		return nil, err
	}
	for _, dev := range list {
		desc, err := c.libusb.getDeviceDesc(dev)
		e := &hotplugEvent{
			eventType:  HotplugEventDeviceArrived,
			ctx:        c,
			dev:        dev,
			desc:       desc,
			err:        err,
			enumerated: true,
		}
		fn(e)
		if e.deregister {
			dereg()
			break
		}
	}
	return dereg, nil
}

type hotplugEvent struct {
	eventType  HotplugEventType
	ctx        *Context
	dev        *libusbDevice
	desc       *DeviceDesc
	err        error
	enumerated bool
	deregister bool
}

// Type returns the event's type (HotplugEventDeviceArrived or HotplugEventDeviceLeft).
func (e *hotplugEvent) Type() HotplugEventType {
	return e.eventType
}

// DeviceDesc returns the device's descriptor.
func (e *hotplugEvent) DeviceDesc() (*DeviceDesc, error) {
	return e.desc, e.err
}

// IsEnumerated returns true if the device was already plugged in when the callback was registered.
func (e *hotplugEvent) IsEnumerated() bool {
	return e.enumerated
}

// Open opens the device.
func (e *hotplugEvent) Open() (*Device, error) {
	if e.err != nil {
		return nil, e.err
	}
	handle, err := e.ctx.libusb.open(e.dev)
	if err != nil {
		return nil, err
	}
	return &Device{handle: handle, ctx: e.ctx, Desc: e.desc}, nil
}

// Deregister deregisters the callback registration after the callback function returns.
func (e *hotplugEvent) Deregister() {
	e.deregister = true
}
