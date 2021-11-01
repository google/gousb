#!/bin/sh
echo 'mode: count' > coverage.merged &&\
        go list ./... | xargs -n1 -I'{}' sh -c ': > coverage.tmp; go test -v -covermode=count -coverprofile=coverage.tmp {} && tail -n +2 coverage.tmp >> coverage.merged' &&\
        rm coverage.tmp

$HOME/gopath/bin/goveralls -coverprofile=coverage.merged -service=github -ignore libusb.go,error.go ||\
        true
