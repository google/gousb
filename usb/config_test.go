package usb

import "testing"

func TestEndpointInfo(t *testing.T) {
	for _, tc := range []struct {
		ep   EndpointInfo
		want string
	}{
		{
			ep: EndpointInfo{
				Address:       0x86,
				Attributes:    0x02,
				MaxPacketSize: 512,
			},
			want: "Endpoint #6 IN  bulk - unsynchronized data [512 0]",
		},
		{
			ep: EndpointInfo{
				Address:       0x02,
				Attributes:    0x05,
				MaxPacketSize: 512,
				MaxIsoPacket:  512,
			},
			want: "Endpoint #2 OUT isochronous - asynchronous data [512 512]",
		},
	} {
		if got := tc.ep.String(); got != tc.want {
			t.Errorf("%#v.String(): got %q, want %q", tc.ep, got, tc.want)
		}
	}
}
