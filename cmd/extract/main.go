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

package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/googlecloudplatform/pi-delivery/gen/index"
	"github.com/googlecloudplatform/pi-delivery/pkg/obj/gcs"
	"github.com/googlecloudplatform/pi-delivery/pkg/unpack"
)

func main() {
	start := flag.Int64("s", 0, "Start offset")
	n := flag.Int64("n", 100, "Number of digits to read")
	outfile := flag.String("o", "-", "Output file")
	useReadAt := flag.Bool("a", false, "Use ReadAt")
	flag.Parse()

	if *n <= 0 {
		return
	}
	if *start < 0 {
		*start += index.Decimal.TotalDigits()
	}

	out := os.Stdout
	if *outfile != "-" {
		f, err := os.OpenFile(*outfile, os.O_CREATE|os.O_RDWR, 0644)
		if err != nil {
			fmt.Fprintf(os.Stderr, "couldn't open %s: %v", *outfile, err)
			os.Exit(1)
		}
		defer f.Close()
		out = f
	}

	ctx := context.Background()
	sc, err := gcs.NewClient(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "couldn't initialize storage client: %v\n", err)
		os.Exit(1)
	}
	defer sc.Close()

	unpackReader := unpack.NewReader(ctx, index.Decimal.NewReader(ctx, sc.Bucket(index.BucketName)))

	var reader io.Reader
	if *useReadAt {
		reader = io.NewSectionReader(unpackReader, *start, *n)
	} else {
		if _, err := unpackReader.Seek(*start, io.SeekStart); err != nil {
			fmt.Fprintf(os.Stderr, "seek failed: %v\n", err)
			os.Exit(1)
		}
		reader = unpackReader
	}
	written, err := io.CopyN(out, reader, *n)
	if err != nil {
		fmt.Fprintf(os.Stderr, "I/O error: %v\n", err)
		os.Exit(1)
	}
	fmt.Fprintf(os.Stderr, "extracted %d digits\n", written)
}
