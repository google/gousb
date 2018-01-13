// Copyright 2017 the gousb Authors.  All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package gousb

import "context"

func (e *endpoint) newStream(ctx context.Context, size, count int) (*stream, error) {
	var ts []transferIntf
	for i := 0; i < count; i++ {
		t, err := newUSBTransfer(e.ctx, e.h, &e.Desc, size)
		if err != nil {
			for _, t := range ts {
				t.free()
			}
			return nil, err
		}
		ts = append(ts, t)
	}
	return newStream(ctx, ts), nil
}

// NewStreamContext prepares a new read stream that will keep reading data from
// the endpoint until closed, or until context is cancelled.
// Size defines a buffer size for a single read transaction and count
// defines how many transactions should be active at any time.
// By keeping multiple transfers active at the same time, a Stream reduces
// the latency between subsequent transfers and increases reading throughput.
func (e *InEndpoint) NewStreamContext(ctx context.Context, size, count int) (*ReadStream, error) {
	s, err := e.newStream(ctx, size, count)
	if err != nil {
		return nil, err
	}
	s.submitAll()
	return &ReadStream{s: s}, nil
}

// NewStreamContext prepares a new read stream that will keep reading data
// from the endpoint until closed, or until context is cancelled.
// It uses NewStreamContext with a context that is never cancelled.
func (e *InEndpoint) NewStream(size, count int) (*ReadStream, error) {
	return e.NewStreamContext(context.Background(), size, count)
}

// NewStreamContext prepares a new write stream that will write data in the
// background. Size defines a buffer size for a single write transaction and
// count defines how many transactions may be active at any time. By buffering
// the writes, a Stream reduces the latency between subsequent transfers and
// increases writing throughput.
// The passed context can be used to cancel the write.
func (e *OutEndpoint) NewStreamContext(ctx context.Context, size, count int) (*WriteStream, error) {
	s, err := e.newStream(ctx, size, count)
	if err != nil {
		return nil, err
	}
	return &WriteStream{s: s}, nil
}

// NewStream prepares a new write stream that will write data in the
// background. It uses NewStreamContext with a context that is never
// cancelled.
func (e *OutEndpoint) NewStream(size, count int) (*WriteStream, error) {
	return e.NewStreamContext(context.Background(), size, count)
}
