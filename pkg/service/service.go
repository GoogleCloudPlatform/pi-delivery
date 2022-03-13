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

package service

import (
	"context"
	"errors"
	"io"

	"github.com/googlecloudplatform/pi-delivery/pkg/cached"
	"github.com/googlecloudplatform/pi-delivery/pkg/obj"
	"github.com/googlecloudplatform/pi-delivery/pkg/obj/gcs"
	"github.com/googlecloudplatform/pi-delivery/pkg/resultset"
	"github.com/googlecloudplatform/pi-delivery/pkg/unpack"
	"go.uber.org/zap"
)

var errInternal = errors.New("internal error")

type Service struct {
	storage obj.Client
	bucket  obj.Bucket
}

func NewService(ctx context.Context, logger *zap.SugaredLogger, bucketName string) *Service {
	storageClient, err := gcs.NewClient(ctx)
	if err != nil {
		logger.Fatalw("Failed to create a new Storage client",
			"error", err)
	}
	return &Service{
		storage: storageClient,
		bucket:  storageClient.Bucket(bucketName),
	}
}

// Get returns n bytes of pi starting at start.
// The first digit (position 0) is 3 before the decimal point.
func (s *Service) Get(ctx context.Context, logger *zap.SugaredLogger, set resultset.ResultSet, start, n int64) ([]byte, error) {
	logger = logger.With("start", start, "n", n)

	if n == 0 {
		return nil, nil
	}

	// pb.Range.Start counts at the first digit before the decimal point (3)
	// while the rest of the program treats the first digit after the decimal point (1)
	// as the zeroth digit. We need a special handling here.
	zero := start == 0
	unpacked := make([]byte, n)

	off := 0
	if zero {
		n--
		unpacked[0] = set.FirstDigit()
		off = 1
	} else {
		start--
	}

	rr := set.NewReader(ctx, s.bucket)
	defer rr.Close()
	reader := unpack.NewReader(ctx, cached.NewCachedReader(ctx, rr))
	read, err := reader.ReadAt(unpacked[off:], start)

	if err != nil && !errors.Is(err, io.EOF) {
		logger.Errorw("ReadAt returned error",
			"error", err,
		)
		return nil, errInternal
	}
	if zero {
		read++
	}

	return unpacked[:read], nil
}

// Close closes connections used by the service.
func (s *Service) Close() error {
	return s.storage.Close()
}
