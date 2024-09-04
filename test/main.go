package main

import (
	"fmt"

	"github.com/blacknode/gousb"
)

func main() {
	ctx := gousb.NewContext()
	defer ctx.Close()

	ctx.OpenDevices(func(desc *gousb.DeviceDesc) bool {
		fmt.Println(desc.Class, desc.SubClass, desc.Device, gousb.ClassPrinter, desc.Class, gousb.ClassPrinter == desc.Class)

		return false
	})
}
