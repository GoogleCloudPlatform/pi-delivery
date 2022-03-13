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

package gcs

import (
	"context"
	"io"

	"cloud.google.com/go/storage"
	"github.com/googlecloudplatform/pi-delivery/pkg/obj"
	"google.golang.org/api/option"
)

// Implementations for Google Cloud Storage.

type Client struct {
	c *storage.Client
}

type Bucket struct {
	h *storage.BucketHandle
}

type Object struct {
	h *storage.ObjectHandle
}

// NewClient returns a new client object for Google Cloud Storage.
func NewClient(ctx context.Context, ops ...option.ClientOption) (obj.Client, error) {
	client, err := storage.NewClient(ctx, ops...)
	if err != nil {
		return nil, err
	}
	return &Client{c: client}, nil
}

func (c *Client) Bucket(name string) obj.Bucket {
	return &Bucket{h: c.c.Bucket(name)}
}

func (c *Client) Close() error {
	return c.c.Close()
}

func (b *Bucket) Object(name string) obj.Object {
	return &Object{h: b.h.Object(name)}
}

func (o *Object) NewRangeReader(ctx context.Context, offset, length int64) (io.ReadCloser, error) {
	return o.h.NewRangeReader(ctx, offset, length)
}
