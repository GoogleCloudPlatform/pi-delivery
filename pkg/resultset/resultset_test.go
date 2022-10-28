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
	"fmt"
	"sort"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/googlecloudplatform/pi-delivery/pkg/resultset"
	"github.com/googlecloudplatform/pi-delivery/pkg/ycd"
)

func TestResultSet_Sort(t *testing.T) {
	t.Parallel()

	set := resultset.ResultSet{
		{
			Header: &ycd.Header{BlockID: 2},
		},
		{
			Header: &ycd.Header{BlockID: 0},
		},
		{
			Header: &ycd.Header{BlockID: 1},
		},
	}
	if got, want := set.Len(), len(set); got != want {
		t.Errorf("Len() = got %d, want %d", got, want)
	}
	testCases := []struct {
		i, j int
		want bool
	}{
		{1, 0, true},
		{0, 1, false},
	}
	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Less(%d, %d)", tc.i, tc.j), func(t *testing.T) {
			if got := set.Less(tc.i, tc.j); got != tc.want {
				t.Errorf("Less(%d, %d) = got %v, want %v", tc.i, tc.j, got, tc.want)
			}
		})
	}

	sort.Sort(set)
	want := resultset.ResultSet{
		{
			Header: &ycd.Header{BlockID: 0},
		},
		{
			Header: &ycd.Header{BlockID: 1},
		},
		{
			Header: &ycd.Header{BlockID: 2},
		},
	}
	if diff := cmp.Diff(want, set); diff != "" {
		t.Errorf("Sorted list = (-want, +got):\n%s", diff)
	}
}

func TestResultSet_Sets(t *testing.T) {
	t.Parallel()

	type offTestCase struct {
		off, wantID, wantOff int64
	}
	testCases := []struct {
		name                string
		set                 resultset.ResultSet
		wantBlockSize       int64
		wantTotalDigits     int64
		wantBlockByteLength int64
		wantTotalByteLength int64
		wantDigitsPerWord   int
		wantRadix           int
		wantFirstDigit      byte
		offTestCases        []offTestCase
	}{
		{
			name: "decimal",
			set: resultset.ResultSet{
				{
					Header: &ycd.Header{
						FileVersion: "1.1.0",
						Radix:       10,
						FirstDigits: "3.14159265358979323846264338327950288419716939937510",
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
						FileVersion: "1.1.0",
						Radix:       10,
						FirstDigits: "3.14159265358979323846264338327950288419716939937510",
						TotalDigits: int64(0),
						BlockSize:   int64(100),
						BlockID:     int64(1),
						Length:      198,
					},
					Name:             "Pi - Dec - Chudnovsky/Pi - Dec - Chudnovsky - 1.ycd",
					FirstDigitOffset: 201,
				},
				{
					Header: &ycd.Header{
						FileVersion: "1.1.0",
						Radix:       10,
						FirstDigits: "3.14159265358979323846264338327950288419716939937510",
						TotalDigits: int64(0),
						BlockSize:   int64(100),
						BlockID:     int64(2),
						Length:      198,
					},
					Name:             "Pi - Dec - Chudnovsky/Pi - Dec - Chudnovsky - 2.ycd",
					FirstDigitOffset: 201,
				},
			},
			wantBlockSize:       100,
			wantTotalDigits:     300,
			wantBlockByteLength: 48,
			wantTotalByteLength: 144,
			wantDigitsPerWord:   19,
			wantRadix:           10,
			wantFirstDigit:      '3',
			offTestCases: []offTestCase{
				{0, 0, 0},
				{47, 0, 47},
				{48, 1, 0},
				{143, 2, 47},
				{144, 3, 0},
			},
		},
		{
			name: "hexadecimal",
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
				{
					Header: &ycd.Header{
						FileVersion: "1.1.0",
						Radix:       16,
						FirstDigits: "3.243f6a8885a308d313198a2e03707344a4093822299f31d008",
						TotalDigits: int64(0),
						BlockSize:   int64(100),
						BlockID:     int64(2),
						Length:      198,
					},
					Name:             "Pi - Hex - Chudnovsky/Pi - Hex - Chudnovsky - 2.ycd",
					FirstDigitOffset: 201,
				},
			},
			wantBlockSize:       100,
			wantTotalDigits:     300,
			wantBlockByteLength: 56,
			wantTotalByteLength: 168,
			wantDigitsPerWord:   16,
			wantRadix:           16,
			wantFirstDigit:      '3',
			offTestCases: []offTestCase{
				{0, 0, 0},
				{55, 0, 55},
				{56, 1, 0},
				{167, 2, 55},
				{168, 3, 0},
			},
		},
		{
			name: "decimal partial",
			set: resultset.ResultSet{
				{
					Header: &ycd.Header{
						FileVersion: "1.1.0",
						Radix:       10,
						FirstDigits: "3.14159265358979323846264338327950288419716939937510",
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
						FileVersion: "1.1.0",
						Radix:       10,
						FirstDigits: "3.14159265358979323846264338327950288419716939937510",
						TotalDigits: int64(150),
						BlockSize:   int64(100),
						BlockID:     int64(1),
						Length:      198,
					},
					Name:             "Pi - Dec - Chudnovsky/Pi - Dec - Chudnovsky - 1.ycd",
					FirstDigitOffset: 201,
				},
			},
			wantBlockSize:       100,
			wantTotalDigits:     150,
			wantBlockByteLength: 48,
			wantTotalByteLength: 96, // this doesn't reflect the partial block.
			wantDigitsPerWord:   19,
			wantRadix:           10,
			wantFirstDigit:      '3',
			offTestCases:        []offTestCase{},
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			set := tc.set
			if got := set.BlockSize(); got != tc.wantBlockSize {
				t.Errorf("BlockSize() = got %d, want %d", got, tc.wantBlockSize)
			}
			if got := set.TotalDigits(); got != tc.wantTotalDigits {
				t.Errorf("TotalDigits() = got %d, want %d", got, tc.wantTotalDigits)
			}
			if got := set.BlockByteLength(); got != tc.wantBlockByteLength {
				t.Errorf("BlockByteLength() = got %d, want %d", got, tc.wantBlockByteLength)
			}
			if got := set.TotalByteLength(); got != tc.wantTotalByteLength {
				t.Errorf("TotalByteLength() = got %d, want %d", got, tc.wantTotalByteLength)
			}
			if got := set.DigitsPerWord(); got != tc.wantDigitsPerWord {
				t.Errorf("DigitsPerWord() = got %d, want %d", got, tc.wantDigitsPerWord)
			}
			if got := set.Radix(); got != tc.wantRadix {
				t.Errorf("Radix() = got %d, want %d", got, tc.wantRadix)
			}
			if got := set.FirstDigit(); got != tc.wantFirstDigit {
				t.Errorf("FirstDigit() = got %d, want %d", got, tc.wantFirstDigit)
			}
			for _, tc := range tc.offTestCases {
				t.Run(fmt.Sprintf("off %d", tc.off), func(t *testing.T) {
					id, off := set.OffsetToBlockPos(tc.off)
					if id != tc.wantID || off != tc.wantOff {
						t.Errorf("OffsetToBlockPos(%d) = got (%d, %d), want (%d, %d)", tc.off, id, off, tc.wantID, tc.wantOff)
					}
				})
			}
		})
	}
}
