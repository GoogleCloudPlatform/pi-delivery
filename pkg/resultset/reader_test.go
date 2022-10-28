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

package resultset_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"testing"
	"testing/iotest"

	"github.com/golang/mock/gomock"
	"github.com/google/go-cmp/cmp"
	"github.com/googlecloudplatform/pi-delivery/pkg/obj"
	mock_obj "github.com/googlecloudplatform/pi-delivery/pkg/obj/mocks"
	"github.com/googlecloudplatform/pi-delivery/pkg/resultset"
	"github.com/googlecloudplatform/pi-delivery/pkg/tests"
	"github.com/googlecloudplatform/pi-delivery/pkg/ycd"
)

func TestResultSet_ReadAt(t *testing.T) {
	t.Parallel()

	testSet := resultset.ResultSet{
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
	}

	testBuf := tests.GenTestByteSeq(int(testSet.TotalByteLength()))
	testCases := []struct {
		name    string
		off     int64
		n       int
		mock    func(context.Context, *gomock.Controller) obj.Bucket
		wantErr error
		wantN   int
		want    []byte
	}{
		{
			name: "zero bytes",
			off:  0,
			n:    0,
			mock: func(_ context.Context, _ *gomock.Controller) obj.Bucket {
				return nil
			},
			want: []byte{},
		},
		{
			name: "simple from the start",
			off:  0,
			n:    8,
			mock: func(ctx context.Context, ctrl *gomock.Controller) obj.Bucket {
				bucket := mock_obj.NewMockBucket(ctrl)
				object := mock_obj.NewMockObject(ctrl)
				bucket.EXPECT().
					Object(testSet[0].Name).
					Return(object)
				object.EXPECT().
					NewRangeReader(ctx, int64(testSet[0].FirstDigitOffset), int64(8)).
					Return(io.NopCloser(bytes.NewReader(testBuf[:48])), nil)
				return bucket
			},
			wantN: 8,
			want:  testBuf[:8],
		},
		{
			name: "simple with offset",
			off:  16,
			n:    16,
			mock: func(ctx context.Context, ctrl *gomock.Controller) obj.Bucket {
				bucket := mock_obj.NewMockBucket(ctrl)
				object := mock_obj.NewMockObject(ctrl)
				bucket.EXPECT().
					Object(testSet[0].Name).
					Return(object)
				object.EXPECT().
					NewRangeReader(ctx, int64(testSet[0].FirstDigitOffset+16), int64(16)).
					Return(io.NopCloser(bytes.NewReader(testBuf[16:])), nil)
				return bucket
			},
			wantN: 16,
			want:  testBuf[16:32],
		},
		{
			name: "cross boundary",
			off:  40,
			n:    24,
			mock: func(ctx context.Context, ctrl *gomock.Controller) obj.Bucket {
				bucket := mock_obj.NewMockBucket(ctrl)
				object := mock_obj.NewMockObject(ctrl)
				gomock.InOrder(
					bucket.EXPECT().Object(testSet[0].Name).Return(object),
					bucket.EXPECT().Object(testSet[1].Name).Return(object),
				)
				gomock.InOrder(
					object.EXPECT().
						NewRangeReader(ctx, int64(testSet[0].FirstDigitOffset+40), int64(8)).
						Return(io.NopCloser(bytes.NewReader(testBuf[40:48])), nil),
					object.EXPECT().
						NewRangeReader(ctx, int64(testSet[1].FirstDigitOffset), int64(16)).
						Return(io.NopCloser(bytes.NewReader(testBuf[48:])), nil),
				)
				return bucket
			},
			wantN: 24,
			want:  testBuf[40:64],
		},
		{
			name: "reading across EOF",
			off:  88,
			n:    16,
			mock: func(ctx context.Context, ctrl *gomock.Controller) obj.Bucket {
				bucket := mock_obj.NewMockBucket(ctrl)
				object := mock_obj.NewMockObject(ctrl)
				bucket.EXPECT().Object(testSet[1].Name).Return(object)
				object.EXPECT().
					NewRangeReader(ctx, int64(testSet[1].FirstDigitOffset+40), int64(8)).
					Return(io.NopCloser(bytes.NewReader(testBuf[88:])), nil)
				return bucket
			},
			wantErr: io.EOF,
			wantN:   8,
			want:    append(testBuf[88:], make([]byte, 8)...),
		},
		{
			name: "reading past EOF",
			off:  96,
			n:    1,
			mock: func(_ context.Context, _ *gomock.Controller) obj.Bucket {
				return nil
			},
			wantErr: io.EOF,
			wantN:   0,
			want:    make([]byte, 1),
		},
	}
	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%s, off = %d, n = %d", tc.name, tc.off, tc.n), func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			mockCtrl := gomock.NewController(t)

			rd := testSet.NewReader(ctx, tc.mock(ctx, mockCtrl))
			t.Cleanup(func() {
				if err := rd.Close(); err != nil {
					t.Errorf("Close() failed: %v", err)
				}
			})

			buf := make([]byte, tc.n)
			n, err := rd.ReadAt(buf, tc.off)
			if got := err; !errors.Is(got, tc.wantErr) {
				t.Errorf("ReadAt(%p, %d) error got %v, want %v", buf, tc.off, got, tc.wantErr)
			}
			if got := n; got != tc.wantN {
				t.Errorf("ReadAt(%p, %d): n = got %d, want %d", buf, tc.off, got, tc.wantN)
			}
			if diff := cmp.Diff(tc.want, buf); diff != "" {
				t.Errorf("ReadAt(%p, %d) = (-want, +got):\n%s", buf, tc.off, diff)
			}
		})
	}
}

func TestResultSet_IOTest(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		set  resultset.ResultSet
	}{
		{
			name: "whole blocks",
			set: resultset.ResultSet{
				{
					Header: &ycd.Header{
						FileVersion: "1.1.0",
						Radix:       16,
						FirstDigits: "3.243f6a8885a308d313198a2e03707344a4093822299f31d008",
						TotalDigits: int64(0),
						BlockSize:   int64(100),
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
						BlockSize:   int64(100),
						BlockID:     int64(1),
						Length:      198,
					},
					Name:             "Pi - Hex - Chudnovsky/Pi - Hex - Chudnovsky - 1.ycd",
					FirstDigitOffset: 201,
				},
			},
		},
		{
			name: "partial blocks",
			set: resultset.ResultSet{
				{
					Header: &ycd.Header{
						FileVersion: "1.1.0",
						Radix:       16,
						FirstDigits: "3.243f6a8885a308d313198a2e03707344a4093822299f31d008",
						TotalDigits: int64(0),
						BlockSize:   int64(100),
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
						TotalDigits: int64(150),
						BlockSize:   int64(100),
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

			rd := tc.set.NewReader(ctx, bucket)
			t.Cleanup(func() {
				if err := rd.Close(); err != nil {
					t.Errorf("Close() failed: %v", err)
				}
			})
			if err := iotest.TestReader(rd, testBuf); err != nil {
				t.Errorf("TestReader() failed: %v", err)
			}
		})
	}
}
