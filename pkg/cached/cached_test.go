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

package cached

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"testing"
	"testing/iotest"

	"github.com/golang/mock/gomock"
	"github.com/google/go-cmp/cmp"
	mock_obj "github.com/googlecloudplatform/pi-delivery/pkg/obj/mocks"
	"github.com/googlecloudplatform/pi-delivery/pkg/resultset"
	"github.com/googlecloudplatform/pi-delivery/pkg/tests"
	"github.com/googlecloudplatform/pi-delivery/pkg/ycd"
)

func TestCachedReader_New(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	mockCtrl := gomock.NewController(t)

	testSet := resultset.ResultSet{
		{
			Header: &ycd.Header{
				Radix:       10,
				TotalDigits: int64(0),
				BlockSize:   int64(1000),
				BlockID:     int64(0),
				Length:      198,
			},
			Name:             "Pi - Dec - Chudnovsky/Pi - Dec - Chudnovsky - 0.ycd",
			FirstDigitOffset: 201,
		},
	}

	ur := testSet.NewReader(ctx, mock_obj.NewMockBucket(mockCtrl))
	t.Cleanup(func() {
		if err := ur.Close(); err != nil {
			t.Errorf("Close() failed: %v", err)
		}
	})
	if ur == nil {
		t.Fatal("NewReader(): got nil, want non-nil")
	}

	rd := NewCachedReader(ctx, ur)
	if rd == nil {
		t.Errorf("NewCacheReader(): got nil, want non-nil")
	}
	if diff := cmp.Diff(testSet, rd.ResultSet()); diff != "" {
		t.Errorf("reader.ResultSet() = (-want, got):\n%s", diff)
	}
}

func TestCachedReader_Cached(t *testing.T) {
	t.Parallel()

	mockCtrl := gomock.NewController(t)
	testSet := resultset.ResultSet{
		{
			Header: &ycd.Header{
				Radix:       10,
				TotalDigits: int64(0),
				BlockSize:   int64(1000),
				BlockID:     int64(0),
				Length:      198,
			},
			Name:             "Pi - Dec - Chudnovsky/Pi - Dec - Chudnovsky - 0.ycd",
			FirstDigitOffset: 201,
		},
	}
	ctx := context.Background()
	testBuf := tests.GenTestByteSeq(int(testSet.TotalByteLength()))

	bucket := mock_obj.NewMockBucket(mockCtrl)
	object := mock_obj.NewMockObject(mockCtrl)

	// Check if the cache is working around boundaries.
	bucket.EXPECT().
		Object(testSet[0].Name).
		Return(object).
		MaxTimes(4)

	gomock.InOrder(
		object.EXPECT().
			NewRangeReader(ctx, int64(testSet[0].FirstDigitOffset), int64(10)).
			Return(io.NopCloser(bytes.NewReader(testBuf)), nil).
			MaxTimes(1),
		object.EXPECT().
			NewRangeReader(ctx, int64(testSet[0].FirstDigitOffset)+20, int64(10)).
			Return(io.NopCloser(bytes.NewReader(testBuf[20:])), nil).
			MaxTimes(1),
		object.EXPECT().
			NewRangeReader(ctx, int64(testSet[0].FirstDigitOffset)+10, int64(10)).
			Return(io.NopCloser(bytes.NewReader(testBuf[10:])), nil).
			MaxTimes(1),
		object.EXPECT().
			NewRangeReader(ctx, int64(testSet[0].FirstDigitOffset)+20, int64(10)).
			Return(io.NopCloser(bytes.NewReader(testBuf[20:])), nil).
			MaxTimes(1),
	)

	ur := testSet.NewReader(ctx, bucket)
	t.Cleanup(func() {
		if err := ur.Close(); err != nil {
			t.Errorf("Close() failed: %v", err)
		}
	})
	rd := NewCachedReader(ctx, ur)

	testCases := []struct {
		off int64
		n   int
	}{
		{0, 10},
		{20, 10},
		{10, 10},
		{20, 10},
		{0, 30},
	}
	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%d, %d", tc.off, tc.n), func(t *testing.T) {
			buf := make([]byte, tc.n)
			n, err := rd.ReadAt(buf, tc.off)
			if err != nil {
				t.Errorf("ReadAt(%p, %d) failed: %v", buf, tc.off, err)
			}
			if got, want := n, tc.n; got != want {
				t.Errorf("ReadAt(%p, %d): n = got %d, want %d", buf, tc.off, got, want)
			}
			if diff := cmp.Diff(testBuf[tc.off:tc.off+int64(tc.n)], buf); diff != "" {
				t.Errorf("ReadAt(%p, %d) = (-want, +got):\n%s", buf, tc.off, diff)
			}
		})
	}
}

func TestCacheReader_Simple(t *testing.T) {
	t.Parallel()

	mockCtrl := gomock.NewController(t)
	testSet := resultset.ResultSet{
		{
			Header: &ycd.Header{
				Radix:       10,
				TotalDigits: int64(0),
				BlockSize:   int64(1000),
				BlockID:     int64(0),
				Length:      198,
			},
			Name:             "Pi - Dec - Chudnovsky/Pi - Dec - Chudnovsky - 0.ycd",
			FirstDigitOffset: 201,
		},
	}
	ctx := context.Background()
	testBuf := tests.GenTestByteSeq(int(testSet.TotalByteLength()))
	bucket := tests.NewMockBucket(ctx, mockCtrl, testSet, testBuf)

	ur := testSet.NewReader(ctx, bucket)
	t.Cleanup(func() {
		if err := ur.Close(); err != nil {
			t.Errorf("Close() failed: %v", err)
		}
	})

	rd := NewCachedReader(ctx, ur)
	testCases := []struct {
		off int64
		n   int
	}{
		{15, 30},
		{25, 10},
		{11, 2},
		{0, 50},
		{52, 1},
		{54, 10},
		{51, 1},
		{50, 15},
	}
	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%d, %d", tc.off, tc.n), func(t *testing.T) {
			buf := make([]byte, tc.n)
			n, err := rd.ReadAt(buf, tc.off)
			if err != nil {
				t.Errorf("ReadAt(%p, %d) failed: %v", buf, tc.off, err)
			}
			if got, want := n, tc.n; got != want {
				t.Errorf("ReadAt(%p, %d): n = got %d, want %d", buf, tc.off, got, want)
			}
			if diff := cmp.Diff(testBuf[tc.off:tc.off+int64(tc.n)], buf); diff != "" {
				t.Errorf("ReadAt(%p, %d) = (-want, +got):\n%s", buf, tc.off, diff)
			}
		})
	}
}

func TestCachedReader_IOTest(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		set  resultset.ResultSet
	}{
		{
			name: "small",
			set: resultset.ResultSet{
				{
					Header: &ycd.Header{
						Radix:       10,
						TotalDigits: int64(0),
						BlockSize:   int64(100),
						BlockID:     int64(0),
						Length:      198,
					},
					Name:             "Pi - Dec - Chudnovsky/Pi - Dec - Chudnovsky - 0.ycd",
					FirstDigitOffset: 201,
				},
				{
					Header: &ycd.Header{
						Radix:       10,
						TotalDigits: int64(0),
						BlockSize:   int64(100),
						BlockID:     int64(1),
						Length:      198,
					},
					Name:             "Pi - Dec - Chudnovsky/Pi - Dec - Chudnovsky - 1.ycd",
					FirstDigitOffset: 201,
				},
			},
		},
		{
			name: "large",
			set: resultset.ResultSet{
				{
					Header: &ycd.Header{
						FileVersion: "1.1.0",
						Radix:       16,
						FirstDigits: "3.243f6a8885a308d313198a2e03707344a4093822299f31d008",
						TotalDigits: int64(0),
						BlockSize:   int64(1200000),
						BlockID:     int64(0),
						Length:      198,
					},
					Name:             "Pi - Hex - Chudnovsky/Pi - Hex - Chudnovsky - 0.ycd",
					FirstDigitOffset: 201,
				},
				{
					Header: &ycd.Header{
						FileVersion: "1.1.0",
						Radix:       16,
						FirstDigits: "3.243f6a8885a308d313198a2e03707344a4093822299f31d008",
						TotalDigits: int64(0),
						BlockSize:   int64(1200000),
						BlockID:     int64(1),
						Length:      198,
					},
					Name:             "Pi - Hex - Chudnovsky/Pi - Hex - Chudnovsky - 1.ycd",
					FirstDigitOffset: 201,
				},
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			mockCtrl := gomock.NewController(t)
			testBuf := tests.GenTestByteSeq(int(tc.set.TotalByteLength()))
			bucket := tests.NewMockBucket(ctx, mockCtrl, tc.set, testBuf)

			ur := tc.set.NewReader(ctx, bucket)
			t.Cleanup(func() {
				if err := ur.Close(); err != nil {
					t.Errorf("Close() failed: %v", err)
				}
			})
			rd := NewCachedReader(ctx, ur)
			if err := iotest.TestReader(rd, testBuf); err != nil {
				t.Errorf("TestReader() failed: %v", err)
			}
		})
	}
}
