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

package ycd

import (
	"fmt"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

const rawTestDataHex = `#Compressed Digit File

FileVersion:	1.1.0

Base:	16

FirstDigits:	3.243f6a8885a308d313198a2e03707344a4093822299f31d008

TotalDigits:	0

Blocksize:	1000000
BlockID:	0

EndHeader

`

const rawTestDataDec = `#Compressed Digit File

FileVersion:	1.1.0

Base:	10

FirstDigits:	3.14159265358979323846264338327950288419716939937510

TotalDigits:	50000001

Blocksize:	1000000
BlockID:	50

EndHeader

`

func TestYCD_Parse(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		raw  string
		want *YCDFile
	}{
		{
			name: "hex",
			raw:  strings.ReplaceAll(rawTestDataHex, "\n", "\r\n") + "\x00",
			want: &YCDFile{
				Header: &Header{
					FileVersion: "1.1.0",
					Radix:       16,
					FirstDigits: "3.243f6a8885a308d313198a2e03707344a4093822299f31d008",
					TotalDigits: 0,
					BlockSize:   1000000,
					BlockID:     0,
					Length:      192,
				},
				FirstDigitOffset: 195,
			},
		},
		{
			name: "dec",
			raw:  strings.ReplaceAll(rawTestDataDec, "\n", "\r\n") + "\x00",
			want: &YCDFile{
				Header: &Header{
					FileVersion: "1.1.0",
					Radix:       10,
					FirstDigits: "3.14159265358979323846264338327950288419716939937510",
					TotalDigits: 50000001,
					BlockSize:   1000000,
					BlockID:     50,
					Length:      200,
				},
				FirstDigitOffset: 203,
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, err := Parse(strings.NewReader(tc.raw))
			if err != nil {
				t.Errorf("Parse() failed: %v", err)
			}
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("Parse() = (-want, +got):\n%s", diff)
			}
		})
	}
}

func TestYCD_DigitsPerWord(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		radix int
		want  int
	}{
		{10, 19},
		{16, 16},
	}
	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Radix %d", tc.radix), func(t *testing.T) {
			got := DigitsPerWord(tc.radix)
			if got != tc.want {
				t.Errorf("DigitsPerWord(%d) = got %d, want %d", tc.radix, got, tc.want)
			}
		})
	}
}
