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
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
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

func TestYCD_ParseHex(t *testing.T) {
	crlf := strings.ReplaceAll(rawTestDataHex, "\n", "\r\n")
	crlf += "\x00"

	ycd, err := Parse(strings.NewReader(crlf))
	if err != nil {
		t.Fatalf("failed to parse the test data: %v", err)
	}
	if assert.NotNil(t, ycd, "Parse should return a non-nil value") &&
		assert.NotNil(t, ycd.Header, "Header should return not be nil") {
		assert.Equal(t, "1.1.0", ycd.Header.FileVersion)
		assert.Equal(t, 16, ycd.Header.Radix)
		assert.Equal(t, "3.243f6a8885a308d313198a2e03707344a4093822299f31d008", ycd.Header.FirstDigits)
		assert.Zero(t, ycd.Header.TotalDigits)
		assert.Equal(t, int64(1000000), ycd.Header.BlockSize)
		assert.Zero(t, ycd.Header.BlockID)
		assert.Equal(t, 192, ycd.Header.Length)
		assert.Equal(t, 195, ycd.FirstDigitOffset)
	}
}

func TestYCD_ParseDec(t *testing.T) {
	crlf := strings.ReplaceAll(rawTestDataDec, "\n", "\r\n")
	crlf += "\x00"

	ycd, err := Parse(strings.NewReader(crlf))
	if err != nil {
		t.Fatalf("failed to parse the test data: %v", err)
	}
	if assert.NotNil(t, ycd, "Parse should return a non-nil value") &&
		assert.NotNil(t, ycd.Header, "Header should return not be nil") {
		assert.Equal(t, "1.1.0", ycd.Header.FileVersion)
		assert.Equal(t, 10, ycd.Header.Radix)
		assert.Equal(t, "3.14159265358979323846264338327950288419716939937510", ycd.Header.FirstDigits)
		assert.Equal(t, int64(50000001), ycd.Header.TotalDigits)
		assert.Equal(t, int64(1000000), ycd.Header.BlockSize)
		assert.Equal(t, int64(50), ycd.Header.BlockID)
		assert.Equal(t, 200, ycd.Header.Length)
		assert.Equal(t, 203, ycd.FirstDigitOffset)
	}
}

func TestYCD_DigitsPerWord(t *testing.T) {
	assert.Equal(t, 19, DigitsPerWord(10))
	assert.Equal(t, 16, DigitsPerWord(16))
}
