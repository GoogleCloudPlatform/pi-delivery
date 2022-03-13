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

package ycd

import (
	"bufio"
	"fmt"
	"strconv"
	"strings"
)

type Header struct {
	// FileVersion is the version of the ycd file.
	// Currently it's 1.1.0 and this code is tested against the version.
	FileVersion string

	// Radis is the radix of the file. 10 or 16.
	Radix int

	// FirstDigits is always the first several digits of pi?
	// e.g. 3.14159265358979323846264338327950288419716939937510 for decimal and
	// 3.243f6a8885a308d313198a2e03707344a4093822299f31d008 for hexadecimal.
	FirstDigits string

	// TotalDigits is zero if the file has n == BlockSize.
	// otherwise it's the number of digits in the file.
	TotalDigits int64

	// BlockSize is digits per file.
	BlockSize int64

	// BlockID is the position of the current file.
	BlockID int64

	// Length is the total byte length of the header in the file.
	// It is the offset of the empty line after EndHeader.
	Length int
}

func parseInt64(s string) (int64, error) {
	return strconv.ParseInt(s, 10, 64)
}

func (h *Header) validate() error {
	if h.FileVersion != "1.1.0" {
		return fmt.Errorf("unknown file version: %s", h.FileVersion)
	}
	if h.Radix != 10 && h.Radix != 16 {
		return fmt.Errorf("unknown radix: %v", h.Radix)
	}
	return nil
}

func parseHeader(reader *bufio.Reader) (*Header, error) {
	var h Header
	length := 0

	line, err := reader.ReadString('\n')
	if err != nil {
		return nil, err
	}
	length += len(line)
	if strings.TrimSpace(line) != "#Compressed Digit File" {
		return nil, fmt.Errorf("first line should be '#Compressed Digit File': %s", strings.TrimSpace(line))
	}

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			return nil, err
		}
		length += len(line)
		// The delimiter is \r\n so it's two runes.
		if len(line) == 2 {
			continue
		}
		tokens := strings.Split(line, ":")
		key := strings.TrimSpace(tokens[0])
		if key == "EndHeader" {
			break
		}

		value := strings.TrimSpace(tokens[1])
		switch key {
		case "FileVersion":
			h.FileVersion = value
		case "Base":
			if i, err := strconv.Atoi(value); err == nil {
				h.Radix = i
			} else {
				return nil, err
			}
		case "FirstDigits":
			h.FirstDigits = value
		case "TotalDigits":
			if i, err := parseInt64(value); err == nil {
				h.TotalDigits = i
			} else {
				return nil, err
			}
		case "Blocksize":
			if i, err := parseInt64(value); err == nil {
				h.BlockSize = i
			} else {
				return nil, err
			}
		case "BlockID":
			if i, err := parseInt64(value); err == nil {
				h.BlockID = i
			} else {
				return nil, err
			}
		default:
			return nil, fmt.Errorf("unknown header key: %s", key)
		}
	}
	if err := h.validate(); err != nil {
		return nil, err
	}

	h.Length = length
	return &h, nil
}
