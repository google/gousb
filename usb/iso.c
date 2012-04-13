#include <libusb-1.0/libusb.h>

void _callback(struct libusb_transfer *xfer) {
	iso_callback(xfer->user_data);
}

libusb_transfer_cb_fn callback() {
  return &_callback;
}
