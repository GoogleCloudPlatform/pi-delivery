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

package cached

import (
	"context"
	"io"
	"sync"

	"github.com/googlecloudplatform/pi-delivery/pkg/resultset"
)

const cacheSize = 1 * 1024 * 1024 // 1 MiB
var radixes = map[int]int{10: 0, 16: 1}

type cache struct {
	once  sync.Once
	lock  sync.RWMutex
	cache []byte
}

var _cache = make([]cache, len(radixes))

// UpstreamReader is the reader CachedReader reads from.
type UpstreamReader interface {
	io.ReadSeeker
	io.ReaderAt
	// ResultSet returns the upstream result set.
	ResultSet() resultset.ResultSet
}

// CachedReader provides a cache support for up to the first cacheSize bytes
// on top of the UpstreamReader.
type CachedReader struct {
	off   int64
	rd    UpstreamReader
	ctx   context.Context
	cache *cache
}

var _ io.ReadSeeker = new(CachedReader)
var _ io.ReaderAt = new(CachedReader)

// NewCachedReader returns a new CachedReader for upstream rd.
func NewCachedReader(ctx context.Context, rd UpstreamReader) *CachedReader {
	cache := &_cache[radixes[rd.ResultSet().Radix()]]

	cache.once.Do(func() {
		cache.cache = make([]byte, 0, cacheSize)
	})

	return &CachedReader{
		ctx:   ctx,
		rd:    rd,
		off:   0,
		cache: cache,
	}
}

// ReadAt reads len(p) bytes of packed results from offset off.
func (r *CachedReader) ReadAt(p []byte, off int64) (int, error) {
	n := 0
	if read, ok := r.readCache(p, off); ok {
		n += read
		if n == len(p) {
			return n, nil
		}
	}
	read, err := r.rd.ReadAt(p[n:], off+int64(n))
	r.updateCache(p[n:n+read], off+int64(n))
	return n + read, err
}

// Read reads len(p) bytes of packed results from the current offset.
func (r *CachedReader) Read(p []byte) (int, error) {
	if n, ok := r.readCache(p, r.off); ok {
		r.Seek(int64(n), io.SeekCurrent)
		return n, nil
	}
	n, err := r.rd.Read(p)
	r.updateCache(p[:n], r.off)
	r.off += int64(n)
	return n, err
}

// Seek updates the offset for the next Read.
func (r *CachedReader) Seek(offset int64, whence int) (int64, error) {
	off, err := r.rd.Seek(offset, whence)
	r.off = off
	return off, err
}

func (r *CachedReader) readCache(p []byte, offset int64) (int, bool) {
	r.cache.lock.RLock()
	defer r.cache.lock.RUnlock()
	if int64(len(r.cache.cache)) <= offset {
		return 0, false
	}
	return copy(p, r.cache.cache[offset:]), true
}

func (r *CachedReader) updateCache(p []byte, offset int64) {
	// This is lazy so just update if the data is contiguous.
	// Check boundaries with a read lock first.
	r.cache.lock.RLock()
	if int64(len(r.cache.cache)) < offset ||
		int64(cap(r.cache.cache)) <= offset {
		r.cache.lock.RUnlock()
		return
	}
	r.cache.lock.RUnlock()

	r.cache.lock.Lock()
	defer r.cache.lock.Unlock()
	overlap := len(r.cache.cache) - int(offset)
	n := len(p) - overlap
	if overlap >= 0 && n > 0 {
		if len(r.cache.cache)+n > cap(r.cache.cache) {
			n = cap(r.cache.cache) - len(r.cache.cache)
		}
		r.cache.cache = append(r.cache.cache, p[overlap:overlap+n]...)
	}
}

// ResultSet returns the upstream ResultSet.
func (r *CachedReader) ResultSet() resultset.ResultSet {
	return r.rd.ResultSet()
}
