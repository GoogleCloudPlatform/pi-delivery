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
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/googlecloudplatform/pi-delivery/gen/index"
	"go.uber.org/zap"
)

func TestService_SimpleGet(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	testCases := []struct {
		radix    int
		start, n int64
		want     string
	}{
		{10, 0, 1, "3"},
		{10, 1, 1, "1"},
		{10, 0, 50, "31415926535897932384626433832795028841971693993751"},
		{10, 1, 50, "14159265358979323846264338327950288419716939937510"},
		{16, 0, 1, "3"},
		{16, 1, 1, "2"},
		{16, 0, 50, "3243f6a8885a308d313198a2e03707344a4093822299f31d00"},
		{16, 1, 50, "243f6a8885a308d313198a2e03707344a4093822299f31d008"},
	}

	l, _ := zap.NewDevelopment()
	s := l.Sugar()
	serv := NewService(ctx, s, index.BucketName)
	if serv == nil {
		t.Fatal("NewService() got nil, want non-nil")
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(fmt.Sprintf("Radix %d Start %d N %d", tc.radix, tc.start, tc.n), func(t *testing.T) {
			t.Parallel()
			set := index.Decimal
			if tc.radix == 16 {
				set = index.Hexadecimal
			}
			got, err := serv.Get(ctx, s, set, tc.start, tc.n)
			if err != nil {
				t.Errorf("Get() failed: %v", err)
			}
			if diff := cmp.Diff(tc.want, string(got)); diff != "" {
				t.Errorf("Get() = (-want, +got):\n%s", diff)
			}
		})
	}
}
