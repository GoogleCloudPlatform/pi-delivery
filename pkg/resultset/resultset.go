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

package resultset

import (
	"context"
	"io"
	"sort"

	"github.com/googlecloudplatform/pi-delivery/pkg/obj"
	"github.com/googlecloudplatform/pi-delivery/pkg/ycd"
)

// ResultSet is a list of YCD files consisting the same pi calculation result.
type ResultSet []*ycd.YCDFile

var _ sort.Interface = new(ResultSet)

// Len is the number of elements in the collection.
func (s ResultSet) Len() int {
	return len(s)
}

// Less reports whether the element with index i must sort before the element with index j.
func (s ResultSet) Less(i, j int) bool {
	return s[i].Header.BlockID < s[j].Header.BlockID
}

// Swap swaps the elements with indexes i and j.
func (s ResultSet) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

// NewReader returns a new ResultSetReader with bucket.
func (s ResultSet) NewReader(ctx context.Context, bucket obj.Bucket) *Reader {
	return &Reader{
		bucket: bucket,
		set:    s,
	}
}

// TotalDigits returns the total number of digits in the array.
// y-cruncher doesn't seem to set TotalDigits unless a particular ycd file has
// a smaller number of digits smaller than the block size.
// If any file has the total digit field set, this simply returns the value of the field.
// Otherwise, it returns the sum of block sizes.
func (s ResultSet) TotalDigits() int64 {
	total := int64(0)
	for _, v := range s {
		if v.Header.TotalDigits != 0 {
			return v.Header.TotalDigits
		}
		total += v.Header.BlockSize
	}
	return total
}

// BlockSize returns the number of digits in each block.
func (s ResultSet) BlockSize() int64 {
	if len(s) == 0 {
		return 0
	}
	return s[0].Header.BlockSize
}

// BlockByteLength returns the byte length of each block.
func (s ResultSet) BlockByteLength() int64 {
	if len(s) == 0 {
		return 0
	}
	return s[0].BlockByteLength()
}

// OffSetToBlockPos converts the byte offset off to a set of blockID and blockOff.
func (s ResultSet) OffsetToBlockPos(off int64) (blockID, blockOff int64) {
	if s.BlockSize() == 0 {
		return 0, off
	}
	return off / s.BlockByteLength(), off % s.BlockByteLength()
}

// TotalByteLength returns the total byte length of the digits.
// It doesn't include headers.
func (s ResultSet) TotalByteLength() int64 {
	if len(s) == 0 {
		return 0
	}
	return s[0].BlockByteLength() * int64(len(s))
}

// DigitsPerWord returns the number of digits per word.
func (s ResultSet) DigitsPerWord() int {
	if len(s) == 0 {
		return 0
	}
	return ycd.DigitsPerWord(s.Radix())
}

// Radix returns the base of the result set.
func (s ResultSet) Radix() int {
	if len(s) == 0 {
		return 0
	}
	return s[0].Header.Radix
}

// FirstDigit returns the first digit of the result.
func (s ResultSet) FirstDigit() byte {
	if len(s) == 0 {
		return 0
	}
	return s[0].Header.FirstDigits[0]
}

// newRangeReader returns a io.ReadCloser for section [off, off+length) in the resultset.
func newRangeReader(ctx context.Context, set ResultSet, bucket obj.Bucket, off, length int64) (io.ReadCloser, error) {
	if off >= set.TotalByteLength() {
		return nil, io.EOF
	}
	block, blockOff := set.OffsetToBlockPos(off)
	obj := bucket.Object(set[block].Name)
	blockByteLen := set[block].BlockByteLength()
	if length < 0 {
		length = blockByteLen - blockOff
	} else if blockByteLen < blockOff+length {
		length = blockByteLen - blockOff
	}

	return obj.NewRangeReader(ctx, blockOff+int64(set[block].FirstDigitOffset), length)
}
