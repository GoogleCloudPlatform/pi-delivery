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
	"io"
	"testing"
	"testing/iotest"

	"github.com/golang/mock/gomock"
	mock_obj "github.com/googlecloudplatform/pi-delivery/pkg/obj/mocks"
	"github.com/googlecloudplatform/pi-delivery/pkg/resultset"
	"github.com/googlecloudplatform/pi-delivery/pkg/tests"
	"github.com/googlecloudplatform/pi-delivery/pkg/ycd"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func genTestByteSeq(n int) []byte {
	buf := make([]byte, n)
	for i := 0; i < n; i++ {
		buf[i] = byte(i)
	}
	return buf
}

func TestCachedReader_Simple(t *testing.T) {
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
	testBuf := genTestByteSeq(int(testSet.TotalByteLength()))

	bucket := mock_obj.NewMockBucket(mockCtrl)
	object := mock_obj.NewMockObject(mockCtrl)

	rr := testSet.NewReader(ctx, bucket)
	require.NotNil(t, rr)

	reader := NewCachedReader(ctx, rr)
	require.NotNil(t, reader)

	assert.Equal(t, testSet, reader.ResultSet())

	test := func(off int64, n int) {
		buf := make([]byte, n)
		if m, err := reader.ReadAt(buf, off); assert.NoError(t, err) {
			assert.Equal(t, n, m)
			assert.Equal(t, testBuf[off:off+int64(n)], buf)
		}
	}

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

	test(0, 10)
	test(20, 10)
	test(10, 10)
	test(20, 10)
	test(0, 30)

	// Make sure the cache is correctly constructed.
	bucket.EXPECT().Object(testSet[0].Name).Return(object).AnyTimes()

	object.EXPECT().NewRangeReader(
		gomock.AssignableToTypeOf(ctx),
		gomock.Any(),
		gomock.Any(),
	).DoAndReturn(
		func(ctx context.Context, off, length int64) (io.ReadCloser, error) {
			return tests.NewTestReader(testSet, 0, testBuf, off, length)
		},
	).AnyTimes()

	test(15, 30)
	test(25, 10)
	test(11, 2)
	test(0, 50)
	test(52, 1)
	test(54, 10)
	test(51, 1)
	test(50, 15)
}

func TestCachedReader_IOTestSmall(t *testing.T) {
	t.Parallel()
	mockCtrl := gomock.NewController(t)

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
	ctx := context.Background()
	testBuf := genTestByteSeq(int(testSet.TotalByteLength()))

	bucket := mock_obj.NewMockBucket(mockCtrl)
	obj0 := mock_obj.NewMockObject(mockCtrl)
	obj1 := mock_obj.NewMockObject(mockCtrl)

	bucket.EXPECT().Object(testSet[0].Name).Return(obj0).AnyTimes()
	bucket.EXPECT().Object(testSet[1].Name).Return(obj1).AnyTimes()

	obj0.EXPECT().NewRangeReader(
		gomock.AssignableToTypeOf(ctx),
		gomock.Any(),
		gomock.Any(),
	).DoAndReturn(
		func(ctx context.Context, off, length int64) (io.ReadCloser, error) {
			return tests.NewTestReader(testSet, 0, testBuf, off, length)
		},
	).AnyTimes()

	obj1.EXPECT().NewRangeReader(
		gomock.AssignableToTypeOf(ctx),
		gomock.Any(),
		gomock.Any(),
	).DoAndReturn(
		func(ctx context.Context, off, length int64) (io.ReadCloser, error) {
			return tests.NewTestReader(testSet, 1, testBuf, off, length)
		},
	).AnyTimes()

	rr := testSet.NewReader(ctx, bucket)
	require.NotNil(t, rr)
	defer assert.NoError(t, rr.Close())

	reader := NewCachedReader(ctx, rr)
	require.NotNil(t, reader)

	assert.Equal(t, testSet, reader.ResultSet())

	assert.NoError(t, iotest.TestReader(reader, testBuf))
}

func TestCachedReader_IOTestLarge(t *testing.T) {
	t.Parallel()
	mockCtrl := gomock.NewController(t)

	testSet := resultset.ResultSet{
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
	}
	ctx := context.Background()
	testBuf := genTestByteSeq(int(testSet.TotalByteLength()))

	bucket := mock_obj.NewMockBucket(mockCtrl)
	obj0 := mock_obj.NewMockObject(mockCtrl)
	obj1 := mock_obj.NewMockObject(mockCtrl)

	bucket.EXPECT().Object(testSet[0].Name).Return(obj0).AnyTimes()
	bucket.EXPECT().Object(testSet[1].Name).Return(obj1).AnyTimes()

	obj0.EXPECT().NewRangeReader(
		gomock.AssignableToTypeOf(ctx),
		gomock.Any(),
		gomock.Any(),
	).DoAndReturn(
		func(ctx context.Context, off, length int64) (io.ReadCloser, error) {
			return tests.NewTestReader(testSet, 0, testBuf, off, length)
		},
	).AnyTimes()

	obj1.EXPECT().NewRangeReader(
		gomock.AssignableToTypeOf(ctx),
		gomock.Any(),
		gomock.Any(),
	).DoAndReturn(
		func(ctx context.Context, off, length int64) (io.ReadCloser, error) {
			return tests.NewTestReader(testSet, 1, testBuf, off, length)
		},
	).AnyTimes()

	rr := testSet.NewReader(ctx, bucket)
	require.NotNil(t, rr)
	defer assert.NoError(t, rr.Close())

	reader := NewCachedReader(ctx, rr)
	require.NotNil(t, reader)

	assert.Equal(t, testSet, reader.ResultSet())

	assert.NoError(t, iotest.TestReader(reader, testBuf))
}
