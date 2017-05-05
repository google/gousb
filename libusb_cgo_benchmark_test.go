package gousb

import "testing"

func BenchmarkCGo(b *testing.B) {
	for _, bc := range []struct {
		name  string
		bfunc func(*libusbContext, int)
	}{
		{
			name: "simple function",
			bfunc: func(ctx *libusbContext, N int) {
				for i := 0; i < N; i++ {
					libusbSetDebug(ctx, i&1)
				}
			},
		},
		{
			name: "method",
			bfunc: func(ctx *libusbContext, N int) {
				impl := libusbImpl{}
				for i := 0; i < N; i++ {
					impl.setDebug(ctx, i&1)
				}
			},
		},
		{
			name: "interface",
			bfunc: func(ctx *libusbContext, N int) {
				var intf libusbIntf = libusbImpl{}
				for i := 0; i < N; i++ {
					intf.setDebug(ctx, i&1)
				}
			},
		},
	} {
		b.Run(bc.name, func(b *testing.B) {
			ctx, err := libusbImpl{}.init()
			if err != nil {
				b.Fatalf("libusb_init() failed: %v", err)
			}
			defer libusbImpl{}.exit(ctx)
			b.ResetTimer()
			bc.bfunc(ctx, b.N)
		})
	}
}
