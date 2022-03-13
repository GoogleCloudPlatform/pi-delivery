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
	"errors"
	"io"

	"github.com/googlecloudplatform/pi-delivery/pkg/obj"
)

// Reader is a reader for a ResultSet, starting at the first figit (offset = 0)
// to the end of the ResultSet. Reader automatically switches to the next object
// as necessary. Alternatively you can also use ReadAt to read a section of ResultSet.
// Must be created by NewReader() and the caller must Close() after use.
type Reader struct {
	set    ResultSet
	bucket obj.Bucket
	ctx    context.Context
	off    int64
	rd     io.ReadCloser
	seeked bool
}

// Reader implements both io.ReaderAt and io.ReadSeekCloser
var _ io.ReadSeekCloser = new(Reader)
var _ io.ReaderAt = new(Reader)

// ReadAt reads len(p) bytes of packed digits starting at byte result offset
// (first byte in the result set is 0).
// Returns io.EOF at the end of the result set.
func (r *Reader) ReadAt(p []byte, off int64) (int, error) {
	n := 0

	for n < len(p) {
		reader, err := newRangeReader(r.ctx, r.set, r.bucket, off+int64(n), int64(len(p)-n))
		if err != nil {
			return n, err
		}

		for n < len(p) {
			read, err := reader.Read(p[n:])
			n += read
			if err == io.EOF {
				break
			}
			if err != nil {
				_ = reader.Close()
				return n, err
			}
		}
		if err := reader.Close(); err != nil {
			return n, err
		}
	}
	return n, nil
}

// Read reads len(p) bytes of packed digits at the current position.
// Read returns at the end of each block with error == nil.
// Callers should continue to call Read() if it needs more digits.
func (r *Reader) Read(p []byte) (int, error) {
	if r.rd == nil || r.seeked {
		if err := r.Close(); err != nil {
			return 0, err
		}
		reader, err := newRangeReader(r.ctx, r.set, r.bucket, r.off, -1)
		r.rd = reader
		r.seeked = false
		if err != nil {
			return 0, err
		}
	}
	n, err := r.rd.Read(p)
	r.off += int64(n)
	if err == io.EOF {
		// Next Read() call needs to recreate the reader.
		r.seeked = true
		// Ignore EOF because there might be more data.
		return n, nil
	}
	return n, err
}

// Seek sets the byte result offset for the next Read().
func (r *Reader) Seek(offset int64, whence int) (int64, error) {
	off := r.off
	switch whence {
	case io.SeekStart:
		off = offset
	case io.SeekCurrent:
		off += offset
	case io.SeekEnd:
		off = r.set.TotalByteLength() + offset
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

// Close closes the Reader.
func (r *Reader) Close() error {
	if r.rd != nil {
		return r.rd.Close()
	}
	return nil
}

// ResultSet returns the underlying ResultSet.
func (r *Reader) ResultSet() ResultSet {
	return r.set
}
