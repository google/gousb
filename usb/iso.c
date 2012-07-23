#include <libusb-1.0/libusb.h>
#include <stdio.h>
#include <string.h>

void print_xfer(struct libusb_transfer *xfer);

void callback(struct libusb_transfer *xfer) {
	printf("Callback!\n");
	print_xfer(xfer);
	iso_callback(xfer->user_data);
}

int submit(struct libusb_transfer *xfer) {
	xfer->callback = &callback;
	xfer->status = -1;
	//print_xfer(xfer);
	printf("Transfer submitted\n");

	/* fake
	strcpy(xfer->buffer, "hello");
	xfer->actual_length = 5;
	callback(xfer);
	return 0; */
	return libusb_submit_transfer(xfer);
}

void print_xfer(struct libusb_transfer *xfer) {
	int i;

	printf("Transfer:\n");
	printf("  dev_handle:   %p\n", xfer->dev_handle);
	printf("  flags:        %08x\n", xfer->flags);
	printf("  endpoint:     %x\n", xfer->endpoint);
	printf("  type:         %x\n", xfer->type);
	printf("  timeout:      %dms\n", xfer->timeout);
	printf("  status:       %x\n", xfer->status);
	printf("  length:       %d (act: %d)\n", xfer->length, xfer->actual_length);
	printf("  callback:     %p\n", xfer->callback);
	printf("  user_data:    %p\n", xfer->user_data);
	printf("  buffer:       %p\n", xfer->buffer);
	printf("  num_iso_pkts: %d\n", xfer->num_iso_packets);
	printf("  packets:\n");
	for (i = 0; i < xfer->num_iso_packets; i++) {
		printf("    [%04d] %d (act: %d) %x\n", i,
			xfer->iso_packet_desc[i].length,
			xfer->iso_packet_desc[i].actual_length,
			xfer->iso_packet_desc[i].status);
	}
}
