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
	"bufio"
	"io"
)

type YCDFile struct {
	Header           *Header
	Name             string
	FirstDigitOffset int
}

// WordSize is the size of a word (64 bits / 8 bytes).
const WordSize = 8

// DigitsPerWord returns the number of digits per word (64 bits).
func DigitsPerWord(radix int) int {
	if radix == 10 {
		return 19
	} else if radix == 16 {
		return 16
	} else {
		panic("unknown radix")
	}
}

// Parse parses the header of a ycd file and returns the field values.
func Parse(reader io.Reader) (*YCDFile, error) {
	br := bufio.NewReader(reader)
	y := new(YCDFile)

	// First parse the header
	if h, err := parseHeader(br); err == nil {
		y.Header = h
	} else {
		return nil, err
	}
	// There is a nil character before the first digit.
	y.FirstDigitOffset = y.Header.Length
	for {
		b, err := br.ReadByte()
		if err != nil {
			return nil, err
		}
		y.FirstDigitOffset++
		if b == 0 {
			break
		}
	}
	return y, nil
}

// BlockByteLength returns the total byte length of the block
// rounding up to the word alignment.
func (y *YCDFile) BlockByteLength() int64 {
	dpw := int64(DigitsPerWord(y.Header.Radix))
	return (y.Header.BlockSize + dpw - 1) / dpw * WordSize
}
