// Copyright 2022 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package tests

import (
	"bytes"
	"io"

	"github.com/googlecloudplatform/pi-delivery/pkg/resultset"
)

// NewTestReader returns an io.ReadCloser for buf [off, off+length]
// based on the offsets in set.
func NewTestReader(set resultset.ResultSet, idx int, buf []byte, off, length int64) (io.ReadCloser, error) {
	if idx >= len(set) {
		return nil, io.EOF
	}
	off -= int64(set[idx].FirstDigitOffset)
	if off >= set[idx].BlockByteLength() {
		return nil, io.EOF
	}

	base := int64(idx) * set.BlockByteLength()
	next := int64(idx+1) * set.BlockByteLength()
	if length < 0 {
		return io.NopCloser(
			bytes.NewReader(
				buf[off+base : next],
			),
		), nil
	}
	end := off + length
	if end > set.BlockByteLength() {
		end = set.BlockByteLength()
	}
	if end+base > int64(len(buf)) {
		end = int64(len(buf)) - base
	}
	if off+base == end+base {
		// zero bytes remaining
		return nil, io.EOF
	}
	return io.NopCloser(
		bytes.NewReader(
			buf[off+base : end+base]),
	), nil
}
