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
	"bytes"
	"encoding/binary"
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"

	"github.com/googlecloudplatform/pi-delivery/pkg/ycd"
)

const zeros = "0000000000000000000"

func main() {
	radix := flag.Int("r", 10, "radix")
	blockSize := flag.Int("b", 100, "block size")
	flag.Parse()

	dpw := ycd.DigitsPerWord(*radix)

	block := make([]byte, *blockSize)
	uw := make([]byte, dpw)
	bin := make([]byte, 8)
	for {
		n, err := io.ReadFull(os.Stdin, block)
		if err != nil && err != io.ErrUnexpectedEOF {
			if err != io.EOF {
				fmt.Fprintf(os.Stderr, "ReadFull: %v", err)
			}
			break
		}

		rd := bytes.NewReader(block[:n])
		for {
			copy(uw, zeros)
			_, err := io.ReadFull(rd, uw)
			if err != nil && err != io.ErrUnexpectedEOF {
				if err != io.EOF {
					fmt.Fprintf(os.Stderr, "ReadFull: %v", err)
				}
				break
			}
			word, err := strconv.ParseUint(string(uw), *radix, 64)
			if err != nil {
				fmt.Fprintf(os.Stderr, "ParseUint: %v", err)
				break
			}
			binary.LittleEndian.PutUint64(bin, word)
			for _, v := range bin {
				fmt.Fprintf(os.Stdout, "0x%02x, ", v)
			}
			fmt.Fprintln(os.Stdout)
		}
		fmt.Fprintln(os.Stdout, "// Block Boundary")
	}
}
