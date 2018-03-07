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

#include <libusb.h>
#include "_cgo_export.h"

int gousb_hotplug_register_callback(
	libusb_context* ctx,
	libusb_hotplug_event events,
	libusb_hotplug_flag flags,
	int vid,
	int pid,
	int dev_class,
	void *user_data,
	libusb_hotplug_callback_handle *handle
) {
	return libusb_hotplug_register_callback(
		ctx, events, flags, vid, pid, dev_class, (libusb_hotplug_callback_fn)(goHotplugCallback), user_data, handle
	);
}
