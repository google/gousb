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

// Package usb provides a wrapper around libusb-1.0.
package usb

type Context struct {
	ctx  *libusbContext
	done chan struct{}
}

func (c *Context) Debug(level int) {
	libusb.setDebug(c.ctx, level)
}

func NewContext() *Context {
	c, err := libusb.init()
	if err != nil {
		panic(err)
	}
	ctx := &Context{
		ctx:  c,
		done: make(chan struct{}),
	}
	go libusb.handleEvents(ctx.ctx, ctx.done)
	return ctx
}

// ListDevices calls each with each enumerated device.
// If the function returns true, the device is opened and a Device is returned if the operation succeeds.
// Every Device returned (whether an error is also returned or not) must be closed.
// If there are any errors enumerating the devices,
// the final one is returned along with any successfully opened devices.
func (c *Context) ListDevices(each func(desc *Descriptor) bool) ([]*Device, error) {
	list, err := libusb.getDevices(c.ctx)
	if err != nil {
		return nil, err
	}

	var reterr error
	var ret []*Device
	for _, dev := range list {
		desc, err := libusb.getDeviceDesc(dev)
		if err != nil {
			libusb.dereference(dev)
			reterr = err
			continue
		}

		if each(desc) {
			handle, err := libusb.open(dev)
			if err != nil {
				reterr = err
				continue
			}
			ret = append(ret, newDevice(handle, desc))
		} else {
			libusb.dereference(dev)
		}
	}
	return ret, reterr
}

// OpenDeviceWithVidPid opens Device from specific VendorId and ProductId.
// If none is found, it returns nil and nil error. If there are multiple devices
// with the same VID/PID, it will return one of them, picked arbitrarily.
// If there were any errors during device list traversal, it is possible
// it will return a non-nil device and non-nil error. A Device.Close() must
// be called to release the device if the returned device wasn't nil.
func (c *Context) OpenDeviceWithVidPid(vid, pid int) (*Device, error) {
	var found bool
	devs, err := c.ListDevices(func(desc *Descriptor) bool {
		if found {
			return false
		}
		if desc.Vendor == ID(vid) && desc.Product == ID(pid) {
			found = true
			return true
		}
		return false
	})
	if len(devs) == 0 {
		return nil, err
	}
	return devs[0], nil
}

func (c *Context) Close() error {
	c.done <- struct{}{}
	if c.ctx != nil {
		libusb.exit(c.ctx)
	}
	c.ctx = nil
	return nil
}
