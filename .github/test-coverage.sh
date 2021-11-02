#!/bin/sh
echo 'mode: count' > coverage.merged &&\
        go list -f '{{.Dir}}' ./... | xargs -n1 -I'{}' sh -c ': > coverage.tmp; go test -v -covermode=count -coverprofile=coverage.tmp {} && tail -n +2 coverage.tmp >> coverage.merged' &&\
        rm coverage.tmp
