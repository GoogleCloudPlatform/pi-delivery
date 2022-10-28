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
	"context"
	"io"

	"github.com/golang/mock/gomock"
	"github.com/googlecloudplatform/pi-delivery/pkg/obj"
	mock_obj "github.com/googlecloudplatform/pi-delivery/pkg/obj/mocks"
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

// GenTestByteSeq returns a byte slice for tests with length n.
func GenTestByteSeq(n int) []byte {
	buf := make([]byte, n)
	for i := 0; i < n; i++ {
		buf[i] = byte(i)
	}
	return buf
}

// NewMockBucket returns a mock bucket for set that returns testBuf data.
func NewMockBucket(ctx context.Context, ctrl *gomock.Controller, set resultset.ResultSet, testBuf []byte) obj.Bucket {
	bucket := mock_obj.NewMockBucket(ctrl)
	for i, f := range set {
		i := i
		obj := mock_obj.NewMockObject(ctrl)
		obj.EXPECT().NewRangeReader(
			gomock.AssignableToTypeOf(ctx),
			gomock.Any(),
			gomock.Any(),
		).DoAndReturn(
			func(ctx context.Context, off, length int64) (io.ReadCloser, error) {
				return NewTestReader(set, i, testBuf, off, length)
			},
		).AnyTimes()
		bucket.EXPECT().Object(f.Name).Return(obj).AnyTimes()
	}
	return bucket
}
