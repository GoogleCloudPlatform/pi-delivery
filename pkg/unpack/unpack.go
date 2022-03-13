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
	"encoding/binary"
	"errors"
	"fmt"
	"strconv"

	"github.com/googlecloudplatform/pi-delivery/pkg/ycd"
)

var ErrUnknownRadix error = errors.New("Unpack: unknown radix")
var ErrBufferTooSmall error = errors.New("Unpack: destination buffer is too small")
var ErrInvalidWord error = errors.New("Unpack: invalid word")

const (
	zeros    = "0000000000000000000"
	WordSize = ycd.WordSize
)

func copyWithZero(dst []byte, s string, nz int) int {
	return copy(dst, zeros[:nz]) + copy(dst[nz:], s)
}

// UnpackBlock reads packed digits from packed and writes unpacked strings to unpacked.
func UnpackBlock(unpacked, packed []byte, radix, pre int) (int, error) {
	if len(packed) == 0 || len(unpacked) == 0 {
		return 0, nil
	}

	dpw := ycd.DigitsPerWord(radix)

	unpackedLen := UnpackedLen(int64(len(packed)-1), radix) - int64(pre)
	if int64(len(unpacked)) < unpackedLen {
		return 0, fmt.Errorf("%w: required = %v bytes, actual buffer = %v bytes",
			ErrBufferTooSmall, unpackedLen, len(unpacked))
	}

	// Unpack the first word with pre.
	// Copy dpw-pre bytes.
	s := strconv.FormatUint(binary.LittleEndian.Uint64(packed), radix)
	nz := dpw - len(s)
	if nz < 0 {
		return 0, fmt.Errorf("%w: word = %16x, unpacked = %s",
			ErrInvalidWord, packed[:WordSize], s)
	}
	nzNeeded := nz - pre
	if nzNeeded < 0 {
		nzNeeded = 0
	}
	n := copy(unpacked, zeros[:nzNeeded])
	if n < dpw-pre && n < len(unpacked) {
		if nz < pre {
			n += copy(unpacked[n:], s[pre-nz:dpw-pre-n+(pre-nz)])
		} else {
			n += copy(unpacked[n:], s[:dpw-pre-n])
		}
	}

	if len(packed) == WordSize {
		return n, nil
	}

	// Process until the second last word.
	for i := WordSize; i < len(packed)-WordSize; i += WordSize {
		s := strconv.FormatUint(binary.LittleEndian.Uint64(packed[i:]), radix)
		nz := dpw - len(s)
		if nz < 0 {
			return n, fmt.Errorf("%w: word = %16x, unpacked = %s", ErrInvalidWord,
				packed[i:i+WordSize], s)
		}
		n += copyWithZero(unpacked[n:], s, nz)
	}

	// Process the last word with post.
	s = strconv.FormatUint(binary.LittleEndian.Uint64(packed[len(packed)-WordSize:]), radix)
	nz = dpw - len(s)
	if nz < 0 {
		return n, fmt.Errorf("%w: word = %16x, unpacked = %s", ErrInvalidWord,
			packed[len(packed)-WordSize:], s)
	}
	n += copy(unpacked[n:], zeros[:nz])
	if n < len(unpacked) {
		n += copy(unpacked[n:], s)
	}
	return n, nil
}

// UnpackedLen returns a number of bytes to store
// an unpacked sequence for n bytes of packed bytes.
func UnpackedLen(n int64, radix int) int64 {
	return n / WordSize * int64(ycd.DigitsPerWord(radix))
}

// ToPackedOffsets calculates packed byte offsets for digits [off, off+len) where dpw is digits per word.
// There are overlapping words between block (file) boundaries if a block size is not aligned with words.
// This function takes that into account and returns increased values.
//
// Returned values:
//	- start: starting byte offset of the first word containing off
//	- n: number of bytes needed to [off, off+len)
//	- pre: number of extra digits between the word aligment and off
//	- post: number of extra digits between the word alignment and off+len-1
// n, pre, post will be 0 if len is 0.
func ToPackedOffsets(off, blockSize, len int64, dpw int) (start, n int64, pre, post int) {
	if dpw <= 0 {
		panic("ToPackedOffsets: zero or negative dpw")
	}
	padding := int64(dpw) - blockSize%int64(dpw)
	if padding == int64(dpw) {
		padding = 0
	}

	len += padding * ((off+len)/blockSize - off/blockSize)
	off += padding * (off / blockSize)
	start = off / int64(dpw)
	pre = int(off - start*int64(dpw))
	if len == 0 {
		n = 0
		post = 0
	} else {
		n = (len + int64(pre) + int64(dpw) - 1) / int64(dpw)
		post = int(n*int64(dpw) - (len + int64(pre)))
	}

	start *= WordSize
	n *= WordSize
	return
}
