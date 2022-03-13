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

func TestResultSet_ReadAt(t *testing.T) {
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

	bucket := mock_obj.NewMockBucket(mockCtrl)
	object := mock_obj.NewMockObject(mockCtrl)

	reader := testSet.NewReader(ctx, bucket)
	require.NotNil(t, reader)
	defer assert.NoError(t, reader.Close())

	assert.Equal(t, testSet, reader.ResultSet())

	// Zero byte read
	n, err := reader.ReadAt(nil, 0)
	assert.Zero(t, n)
	assert.NoError(t, err)

	testBuf := genTestByteSeq(int(testSet.TotalByteLength()))

	// Simple read from the start
	buf := make([]byte, 8)
	bucket.EXPECT().
		Object(testSet[0].Name).
		Return(object)

	object.EXPECT().
		NewRangeReader(ctx, int64(testSet[0].FirstDigitOffset), int64(len(buf))).
		Return(io.NopCloser(bytes.NewReader(testBuf[:48])), nil)

	n, err = reader.ReadAt(buf, 0)
	assert.Equal(t, len(buf), n)
	assert.NoError(t, err)
	assert.Equal(t, testBuf[:8], buf)

	// Simple read with offset
	buf = make([]byte, 16)
	bucket.EXPECT().
		Object(testSet[0].Name).
		Return(object)

	object.EXPECT().
		NewRangeReader(ctx, int64(testSet[0].FirstDigitOffset+16), int64(len(buf))).
		Return(io.NopCloser(bytes.NewReader(testBuf[16:])), nil)

	n, err = reader.ReadAt(buf, 16)
	assert.Equal(t, len(buf), n)
	assert.NoError(t, err)
	assert.Equal(t, testBuf[16:32], buf)

	// Reading across object boundaries
	buf = make([]byte, 24)
	gomock.InOrder(
		bucket.EXPECT().
			Object(testSet[0].Name).
			Return(object),
		bucket.EXPECT().
			Object(testSet[1].Name).
			Return(object),
	)

	gomock.InOrder(
		object.EXPECT().
			NewRangeReader(ctx, int64(testSet[0].FirstDigitOffset+40), int64(8)).
			Return(io.NopCloser(bytes.NewReader(testBuf[40:48])), nil),
		object.EXPECT().
			NewRangeReader(ctx, int64(testSet[1].FirstDigitOffset), int64(len(buf)-8)).
			Return(io.NopCloser(bytes.NewReader(testBuf[48:])), nil),
	)

	n, err = reader.ReadAt(buf, 40)
	assert.Equal(t, len(buf), n)
	assert.NoError(t, err)
	assert.Equal(t, testBuf[40:40+len(buf)], buf)

	// Reading past EOF
	buf = make([]byte, 16)
	bucket.EXPECT().
		Object(testSet[1].Name).
		Return(object)
	object.EXPECT().
		NewRangeReader(ctx, int64(testSet[1].FirstDigitOffset+40), int64(8)).
		Return(io.NopCloser(bytes.NewReader(testBuf[88:])), nil)

	n, err = reader.ReadAt(buf, 88)
	assert.Equal(t, 8, n)
	assert.ErrorIs(t, err, io.EOF)
	assert.Equal(t, append(testBuf[88:], make([]byte, 8)...), buf)
}

func TestResultSet_IOTest(t *testing.T) {
	t.Parallel()
	mockCtrl := gomock.NewController(t)

	testSet := resultset.ResultSet{
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

	reader := testSet.NewReader(ctx, bucket)
	require.NotNil(t, reader)
	defer assert.NoError(t, reader.Close())

	assert.Equal(t, testSet, reader.ResultSet())
	assert.NoError(t, iotest.TestReader(reader, testBuf))
}

func TestResultSet_PartialBlock(t *testing.T) {
	t.Parallel()
	mockCtrl := gomock.NewController(t)

	testSet := resultset.ResultSet{
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

	reader := testSet.NewReader(ctx, bucket)
	require.NotNil(t, reader)
	defer assert.NoError(t, reader.Close())

	assert.Equal(t, testSet, reader.ResultSet())
	assert.NoError(t, iotest.TestReader(reader, testBuf))
}
