package usb

import "testing"

func TestEndpointInfo(t *testing.T) {
	for _, tc := range []struct {
		ep   EndpointInfo
		want string
	}{
		{
			ep: EndpointInfo{
				Number:        6,
				Direction:     EndpointDirectionIn,
				TransferType:  TransferTypeBulk,
				MaxPacketSize: 512,
			},
			want: "Endpoint #6 IN (address 0x86) bulk [512 bytes]",
		},
		{
			ep: EndpointInfo{
				Number:        2,
				Direction:     EndpointDirectionOut,
				TransferType:  TransferTypeIsochronous,
				MaxPacketSize: 512,
				IsoSyncType:   IsoSyncTypeAsync,
				UsageType:     IsoUsageTypeData,
			},
			want: "Endpoint #2 OUT (address 0x02) isochronous - asynchronous data [512 bytes]",
		},
		{
			ep: EndpointInfo{
				Number:        3,
				Direction:     EndpointDirectionIn,
				TransferType:  TransferTypeInterrupt,
				MaxPacketSize: 16,
				UsageType:     InterruptUsageTypePeriodic,
			},
			want: "Endpoint #3 IN (address 0x83) interrupt - periodic [16 bytes]",
		},
	} {
		if got := tc.ep.String(); got != tc.want {
			t.Errorf("%#v.String(): got %q, want %q", tc.ep, got, tc.want)
		}
	}
}
