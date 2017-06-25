// +build !linux

// Licensed under the Apache License, Version 2.0 (the "License");
// File originally made by jpoirier (https://github.com/jpoirier)
// Addapted to the google/gousb library by nicovell3

package gousb


func (d *Device) detachKernelDriver() (err error) {
	return
}

func (d *Device) attachKernelDriver() (err error) {
	return
}  
