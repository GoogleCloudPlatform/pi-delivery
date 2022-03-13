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
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/googlecloudplatform/pi-delivery/pkg/resultset"
	"github.com/googlecloudplatform/pi-delivery/pkg/ycd"
)

// UpstreamReader is the reader UnpackReader reads from.
type UpstreamReader interface {
	io.ReadSeeker
	io.ReaderAt
	// Result returns the upstream result set.
	ResultSet() resultset.ResultSet
}

// UnpackReader reads from the Upstream Reader and converts packed digits
// to unpacked string representation ("14159...").
// Note the first offset is still the first digit after the decimal point as in
// the packed format.
type UnpackReader struct {
	radix       int
	off         int64
	totalDigits int64
	blockSize   int64
	rd          UpstreamReader
	seeked      bool
	unread      []byte
}

var _ io.ReadSeeker = new(UnpackReader)
var _ io.ReaderAt = new(UnpackReader)

var ErrNotFullWord = errors.New("read bytes are not full words")

// NewReader returns a new UnpackReader for UpstreamReader rd
func NewReader(ctx context.Context, rd UpstreamReader) *UnpackReader {
	return &UnpackReader{
		radix:       rd.ResultSet().Radix(),
		totalDigits: rd.ResultSet().TotalDigits(),
		blockSize:   rd.ResultSet().BlockSize(),
		rd:          rd,
	}
}

// ReadAt reads len(p) bytes of unpacked digits starting at the off-th digit.
// ReadAt(p, 0) returns 141592... for decimal results.
// Note that YCD files starts at the second digit after the decimal point
// so we'll treat the 0-th digit specifically.
func (r *UnpackReader) ReadAt(p []byte, off int64) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}
	if off >= r.totalDigits {
		return 0, io.EOF
	}

	start, n, pre, _ := ToPackedOffsets(off, r.blockSize, int64(len(p)), ycd.DigitsPerWord(r.radix))
	packed := make([]byte, n)
	read, err := r.rd.ReadAt(packed, start)
	if read == 0 {
		return 0, err
	}
	if read%WordSize != 0 {
		return 0, fmt.Errorf("read %v bytes: %w", read, ErrNotFullWord)
	}
	remaining := len(p)
	if remaining > int(r.totalDigits-off) {
		remaining = int(r.totalDigits - off)
		err = io.EOF
	}
	written, perr := r.unpack(p[:remaining], packed[:read], off, pre)
	if perr != nil {
		return written, fmt.Errorf("unpack error at off %v: %w", off, perr)
	}
	return written, err
}

// Read reads len(p) bytes of unpacked digits starting at the current reader offset.
func (r *UnpackReader) Read(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}
	if r.off >= r.totalDigits {
		return 0, io.EOF
	}

	written := 0
	read := 0

	dpw := ycd.DigitsPerWord(r.radix)
	start, packedN, pre, post := ToPackedOffsets(r.off, r.blockSize, int64(len(p)), dpw)
	if r.seeked {
		if _, err := r.rd.Seek(start, io.SeekStart); err != nil {
			return written, err
		}
		r.unread = nil
		r.seeked = false
	}

	packed := make([]byte, packedN)
	if len(r.unread) > 0 {
		read += copy(packed, r.unread)
		if post == 0 || packedN > 2*WordSize {
			r.unread = nil
		}
	}

	n, err := io.ReadFull(r.rd, packed[read:])
	read += n

	remaining := len(p)
	if remaining > int(r.totalDigits-r.off) {
		remaining = int(r.totalDigits - r.off)
	}
	n, perr := r.unpack(p[:remaining], packed[:read], r.off, pre)
	r.off += int64(n)
	written += n

	if read%WordSize != 0 {
		return written, fmt.Errorf("off %v, read bytes %v: %w", r.off, n, ErrNotFullWord)
	}
	if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
		return written, fmt.Errorf("read error at off %v: %w", r.off, err)
	}
	if perr != nil {
		poff, _ := r.rd.Seek(0, io.SeekCurrent)
		return written, fmt.Errorf("unpack error at off %v, packed off %v: %w", r.off, poff, err)
	}

	if int64(read) == packedN && post > 0 {
		r.unread = make([]byte, WordSize)
		copy(r.unread, packed[read-WordSize:])
	}

	if err == io.ErrUnexpectedEOF {
		return written, io.EOF
	}
	return written, err
}

// Seek updates the offset for the next Read.
func (r *UnpackReader) Seek(offset int64, whence int) (int64, error) {
	off := r.off
	switch whence {
	case io.SeekStart:
		off = offset
	case io.SeekCurrent:
		off += offset
	case io.SeekEnd:
		off = r.totalDigits + offset
	}
	if off < 0 {
		return r.off, errors.New("Seek: negative offset")
	}
	if r.off != off {
		r.off = off
		r.seeked = true
	}

	return off, nil
}

func (r *UnpackReader) unpack(unpacked, packed []byte, offset int64, pre int) (int, error) {
	poff := 0
	written := 0
	dpw := ycd.DigitsPerWord(r.radix)

	for poff < len(packed) && written < len(unpacked) {
		remaining := len(unpacked) - written
		reqDigits := remaining
		if offset%r.blockSize+int64(remaining) > r.blockSize {
			reqDigits = int(r.blockSize - offset%r.blockSize)

		}
		reqBytes := (reqDigits + dpw - 1) / dpw * WordSize
		n, err := UnpackBlock(unpacked[written:written+reqDigits], packed[poff:poff+reqBytes], r.radix, pre)
		poff += reqBytes
		written += n
		offset += int64(n)

		if err != nil {
			return written, err
		}
		pre = 0
	}
	return written, nil
}
