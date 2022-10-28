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

package unpack

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/googlecloudplatform/pi-delivery/pkg/ycd"
)

func TestUnpack_ToPackedOffsets(t *testing.T) {
	testCases := []struct {
		radix             int
		blockSize         int64
		off, len          int64
		wantStart, wantN  int64
		wantPre, wantPost int
	}{
		{10, 40, 0, 0, 0, 0, 0, 0},
		{10, 40, 0, 1, 0, 8, 0, 18},
		{10, 40, 0, 19, 0, 8, 0, 0},
		{10, 40, 18, 1, 0, 8, 18, 0},
		{10, 40, 18, 2, 0, 16, 18, 18},
		{10, 60, 19, 38, 8, 16, 0, 0},
		{10, 60, 20, 38, 8, 24, 1, 18},
		{10, 40, 39, 42, 16, 40, 1, 18},
		{10, 30, 29, 2, 8, 16, 10, 18},
		{16, 30, 0, 0, 0, 0, 0, 0},
		{16, 30, 0, 1, 0, 8, 0, 15},
		{16, 30, 0, 16, 0, 8, 0, 0},
		{16, 30, 15, 1, 0, 8, 15, 0},
		{16, 30, 15, 2, 0, 16, 15, 15},
		{16, 16, 16, 32, 8, 16, 0, 0},
		{16, 16, 17, 32, 8, 24, 1, 15},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Radix %d BlockSize %d Off %d Len %d", tc.radix, tc.blockSize, tc.off, tc.len), func(t *testing.T) {
			start, n, pre, post := ToPackedOffsets(tc.off, tc.blockSize, tc.len, ycd.DigitsPerWord(tc.radix))
			if start != tc.wantStart || n != tc.wantN || pre != tc.wantPre || post != tc.wantPost {
				t.Errorf("ToPackedOffsets() = (start, n, pre, post): got (%d, %d, %d, %d), want (%d, %d, %d, %d)",
					start, n, pre, post, tc.wantStart, tc.wantN, tc.wantPre, tc.wantPost)
			}
		})
	}
}

func TestUnpack_Errors(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		unpacked int
		packed   int
		radix    int
		pre      int
		wantErr  error
	}{
		{0, 0, 10, 0, nil},
		{0, WordSize, 10, 0, nil},
		{0, 2 * WordSize, 10, 0, nil},
		{17, 2 * WordSize, 10, 1, ErrBufferTooSmall},
		{18, 2 * WordSize, 10, 0, ErrBufferTooSmall},
		{0, 0, 16, 0, nil},
		{0, WordSize, 16, 0, nil},
		{0, 2 * WordSize, 16, 0, nil},
		{14, 2 * WordSize, 16, 1, ErrBufferTooSmall},
		{15, 2 * WordSize, 16, 0, ErrBufferTooSmall},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(fmt.Sprintf("Unpack %d %d %d %d", tc.unpacked, tc.packed, tc.radix, tc.pre), func(t *testing.T) {
			t.Parallel()
			var unpacked, packed []byte
			if tc.unpacked > 0 {
				unpacked = make([]byte, tc.unpacked)
			}
			if tc.packed > 0 {
				packed = make([]byte, tc.packed)
			}
			n, err := UnpackBlock(unpacked, packed, tc.radix, tc.pre)
			if !cmp.Equal(err, tc.wantErr, cmpopts.EquateErrors()) {
				t.Errorf("UnpackBlock() = got error %v, want %v", err, tc.wantErr)
			}
			if n != 0 {
				t.Errorf("UnpackBlock(): n = got %d, want 0", n)
			}
		})
	}
}

func TestUnpack_Unpack(t *testing.T) {
	t.Parallel()

	longPackedDec := []byte{
		0x8e, 0x22, 0xa2, 0x31, 0xfe, 0xa8, 0x16, 0x83,
		0x43, 0xe1, 0x29, 0xbc, 0x73, 0xf4, 0x7c, 0x0c,
		0x03, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	}
	wantLongDec := []byte(
		"9445923078164062862" +
			"0899862803482534211" +
			"0000000000000000003")

	longPackedHex := []byte{
		0x7a, 0x13, 0x6c, 0x0b, 0xef, 0x6e, 0x98, 0x2a,
		0xfb, 0x7e, 0x50, 0xf0, 0x3b, 0xba, 0x76, 0x01,
		0xff, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	}
	wantLongHex := []byte(
		"2a986eef0b6c137a" +
			"0176ba3bf0507efb" +
			"00000000000000ff")

	testCases := []struct {
		radix  int
		packed []byte
		want   []byte
		pre    int
		post   int
	}{
		{10, []byte{0, 0, 0, 0, 0, 0, 0, 0}, []byte("0000000000000000000"), 0, 0},
		{10, []byte{0, 0, 0, 0, 0, 0, 0, 0}, []byte("00000000000000000"), 2, 0},
		{10, []byte{0x60, 0xe2, 0x3e, 0xb8, 0xae, 0x61, 0xa6, 0x13}, []byte("1415926535897932384"), 0, 0},
		{10, []byte{0x60, 0xe2, 0x3e, 0xb8, 0xae, 0x61, 0xa6, 0x13}, []byte("141592653589793238"), 0, 1},
		{10, []byte{0x60, 0xe2, 0x3e, 0xb8, 0xae, 0x61, 0xa6, 0x13}, []byte("415926535897932384"), 1, 0},
		{10, []byte{0x60, 0xe2, 0x3e, 0xb8, 0xae, 0x61, 0xa6, 0x13}, []byte("41592653589793238"), 1, 1},
		{10, []byte{0x60, 0xe2, 0x3e, 0xb8, 0xae, 0x61, 0xa6, 0x13}, []byte("6"), 6, 12},
		{10, []byte{0x00, 0x00, 0xf4, 0x44, 0x82, 0x91, 0x63, 0x45}, []byte("5000000000000000000"), 0, 0},
		{10, []byte{0x00, 0x00, 0xf4, 0x44, 0x82, 0x91, 0x63, 0x45}, []byte("5"), 0, 18},
		{10, []byte{0x00, 0x00, 0xf4, 0x44, 0x82, 0x91, 0x63, 0x45}, []byte("0"), 18, 0},
		{10, longPackedDec, wantLongDec, 0, 0},
		{10, longPackedDec, wantLongDec[1:], 1, 0},
		{10, longPackedDec, wantLongDec[1 : len(wantLongDec)-1], 1, 1},
		{10, longPackedDec, wantLongDec[1 : len(wantLongDec)-18], 1, 18},
		{10, longPackedDec, wantLongDec[18 : len(wantLongDec)-18], 18, 18},
		{16, []byte{0, 0, 0, 0, 0, 0, 0, 0}, []byte("0000000000000000"), 0, 0},
		{16, []byte{0xff, 0xff, 0xff, 0xff, 0x00, 0x00, 0x00, 0x00}, []byte("ffffff"), 10, 0},
		{16, []byte{0xff, 0xff, 0xff, 0xff, 0x00, 0x00, 0x00, 0x00}, []byte("ffff"), 10, 2},
		{16, longPackedHex, wantLongHex, 0, 0},
		{16, longPackedHex, wantLongHex[1:], 1, 0},
		{16, longPackedHex, wantLongHex[1 : len(wantLongHex)-1], 1, 1},
		{16, longPackedHex, wantLongHex[15 : len(wantLongHex)-1], 15, 1},
		{16, longPackedHex, wantLongHex[15 : len(wantLongHex)-15], 15, 15},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(fmt.Sprintf("Radix %d Expected %s Pre %d Post %d", tc.radix, tc.want, tc.pre, tc.post), func(t *testing.T) {
			t.Parallel()
			unpacked := make([]byte, UnpackedLen(int64(len(tc.packed)), tc.radix)-int64(tc.pre+tc.post))
			n, err := UnpackBlock(unpacked, tc.packed, tc.radix, tc.pre)
			if err != nil {
				t.Errorf("UnpackBlock() failed: %v", err)
			}
			if n != len(tc.want) {
				t.Errorf("UnpackBlock(): n = got %d, want %d", n, len(tc.want))
			}
			if diff := cmp.Diff(tc.want, unpacked); diff != "" {
				t.Errorf("UnpackBlock() = (-want, +got):\n%s", diff)
			}
		})
	}
}
