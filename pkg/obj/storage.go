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

package obj

import (
	"context"
	"io"
)

//go:generate go run github.com/golang/mock/mockgen -source=$GOFILE -destination=./mocks/storage.go

// Client is an interface for object storage.
type Client interface {
	// Bucket returns a handle to a bucket specified by name,
	Bucket(name string) Bucket
	// Error closes the client.
	Close() error
}

// Bucket is an interface for an object storage bucket.
type Bucket interface {
	// Object returns a handle to an object specified by name.
	Object(name string) Object
}

// Object is an interface to an object in object storage.
type Object interface {
	// NewRangeReader returns a new io.ReadCloser for the section [offset, offset+length)
	// for the object.
	NewRangeReader(ctx context.Context, offset, length int64) (io.ReadCloser, error)
}
