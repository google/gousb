package usb

// To enable internal debugging:
//   -ldflags "-X github.com/kylelemons/gousb/usb.debugInternal true"

import (
	"io"
	"io/ioutil"
	"log" // TODO(kevlar): make a logger
	"os"
)

var debug *log.Logger
var debugInternal string

func init() {
	var out io.Writer = ioutil.Discard
	if debugInternal != "" {
		out = os.Stderr
	}
	debug = log.New(out, "usb", log.LstdFlags|log.Lshortfile)
}
