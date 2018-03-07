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

// HotplugOptions can be used to configure and register a hotplug event callback.
// By default, the callback will be called when a new device arrives
// or a device leaves.
// Call Arrived to only receive HotplugEventDeviceArrived events,
// and Left to only receive HotplugEventDeviceLeft events.
// Call Enumerate to also receive HotplugEventDeviceArrived events for
// devices that are already plugged in.
// Use ProductID, VendorID and DeviceClass to filter by product id, vendor id
// or device class.
// Call Register to register the callback function. The options cannot
// be changed after registering. The returned HotplugCallback
// can be used to deregister the callback.
// Callbacks are automatically deregistered when the Context is closed.
type HotplugOptions interface {
	Arrived() HotplugOptions
	Left() HotplugOptions
	Enumerate() HotplugOptions
	ProductID(ID) HotplugOptions
	VendorID(ID) HotplugOptions
	DeviceClass(Class) HotplugOptions
	Register(func(HotplugEvent)) (HotplugCallback, error)
}

// HotplugCallback is a registered hotplug callback.
// It only method is Deregister, which can be used to deregister the callback.
// It is safe to call deregister multiple times.
// Callbacks are also automatically deregistered when the Context is closed.
type HotplugCallback interface {
	Deregister()
}

// HotplugEvent is a hotplug event.
// Its methods should only be called during the event callback.
type HotplugEvent interface {
	// Type returns the event's type (HotplugEventDeviceArrived or HotplugEventDeviceLeft).
	Type() HotplugEventType
	// DeviceDesc returns the device's descriptor.
	DeviceDesc() (*DeviceDesc, error)
	// Open opens the device.
	Open() (*Device, error)
}

type hotplugOptions struct {
	ctx       *Context
	events    HotplugEventType
	enumerate bool
	productID int32
	vendorID  int32
	devClass  int32
}

type hotplugEvent struct {
	eventType HotplugEventType
	ctx       *Context
	dev       *libusbDevice
	desc      *DeviceDesc
	err       error
}

type hotplugCallback struct {
	fn func()
}

func (c *hotplugCallback) Deregister() {
	c.fn()
}

// Hotplug sets up a hotplug event callback.
// It returns a HotplugOptions that can be used to configure and register
// the callback.
func (c *Context) Hotplug() HotplugOptions {
	return &hotplugOptions{
		ctx:       c,
		productID: HotplugMatchAny,
		vendorID:  HotplugMatchAny,
		devClass:  HotplugMatchAny,
	}
}

func (h *hotplugOptions) Arrived() HotplugOptions {
	h.events |= HotplugEventDeviceArrived
	return h
}

func (h *hotplugOptions) Left() HotplugOptions {
	h.events |= HotplugEventDeviceLeft
	return h
}

func (h *hotplugOptions) Enumerate() HotplugOptions {
	h.enumerate = true
	return h
}

func (h *hotplugOptions) ProductID(pid ID) HotplugOptions {
	h.productID = int32(pid)
	return h
}

func (h *hotplugOptions) VendorID(vid ID) HotplugOptions {
	h.vendorID = int32(vid)
	return h
}

func (h *hotplugOptions) DeviceClass(devClass Class) HotplugOptions {
	h.devClass = int32(devClass)
	return h
}

func (h *hotplugOptions) Register(fn func(HotplugEvent)) (HotplugCallback, error) {
	if h.events == 0 {
		h.events = HotplugEventAny
	}
	dereg, err := h.ctx.libusb.registerHotplugCallback(h.ctx.ctx, h.events, h.enumerate, h.vendorID, h.productID, h.devClass, func(ctx *libusbContext, dev *libusbDevice, eventType HotplugEventType) {
		desc, err := h.ctx.libusb.getDeviceDesc(dev)
		fn(&hotplugEvent{
			eventType: eventType,
			ctx:       h.ctx,
			dev:       dev,
			desc:      desc,
			err:       err,
		})
	})
	return &hotplugCallback{
		fn: dereg,
	}, err
}

// Type returns the event's type (HotplugEventDeviceArrived or HotplugEventDeviceLeft).
func (e *hotplugEvent) Type() HotplugEventType {
	return e.eventType
}

// DeviceDesc returns the device's descriptor.
func (e *hotplugEvent) DeviceDesc() (*DeviceDesc, error) {
	return e.desc, e.err
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
