package usb

import (
	"testing"
)

func TestBCD(t *testing.T) {
	tests := []struct {
		BCD
		Int int
		Str string
	}{
		{0x1234, 1234, "12.34"},
	}

	for _, test := range tests {
		if got, want := test.BCD.Int(), test.Int; got != want {
			t.Errorf("Int(%x) = %d, want %d", test.BCD, got, want)
		}
		if got, want := test.BCD.String(), test.Str; got != want {
			t.Errorf("String(%x) = %q, want %q", test.BCD, got, want)
		}
	}
}
