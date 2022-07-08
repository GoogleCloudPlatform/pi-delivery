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
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"github.com/googlecloudplatform/pi-delivery/gen/index"
	"github.com/googlecloudplatform/pi-delivery/pkg/obj"
	"github.com/googlecloudplatform/pi-delivery/pkg/obj/gcs"
	"github.com/googlecloudplatform/pi-delivery/pkg/unpack"
	"github.com/sethvargo/go-retry"
	"go.uber.org/zap"
)

const (
	TOTAL_NUMBERS     = 1_000_000_00
	DIGITS_PER_NUMBER = 8
	CHUNK_SIZE        = 100_000_000
	WORKERS           = 256
	SEQUENCE          = "3141592653589793238462643383279502884197169399375105820974944592307816406286208998628034825342117067"
	MIN_MATCH         = 10
)

var logger *zap.SugaredLogger
var wg sync.WaitGroup

type workerContextKey string

type task struct {
	start  int64
	n      int64
	cancel context.CancelFunc
}

func process(ctx context.Context, task *task, logger *zap.SugaredLogger, client obj.Client) error {
	logger.Infof("processing task, start = %d, n = %v", task.start, task.n)

	rrd := index.Decimal.NewReader(ctx, client.Bucket(index.BucketName))
	defer rrd.Close()
	urd := unpack.NewReader(ctx, rrd)
	if _, err := urd.Seek(task.start, io.SeekStart); err != nil {
		return err
	}
	buf := make([]byte, task.n)
	n, err := io.ReadFull(urd, buf)
	if n < MIN_MATCH {
		return err
	}
	buf = buf[:n]
	off := 0
	for len(buf) >= MIN_MATCH {
		i := bytes.Index(buf, []byte(SEQUENCE[:MIN_MATCH]))
		if i < 0 {
			break
		}
		match := buf[i:]
		if len(match) < len(SEQUENCE) {
			p := make([]byte, len(SEQUENCE)-len(match))
			if n, err := io.ReadFull(urd, p); n == 0 {
				return err
			}
			match = append(match, p...)
		}
		l := 0
		for ; l < len(SEQUENCE)-MIN_MATCH; l++ {
			if match[MIN_MATCH+l] != SEQUENCE[MIN_MATCH+l] {
				break
			}
		}
		matchLen := MIN_MATCH + l
		fmt.Printf("%v, %v, %s\n",
			task.start+int64(off+i+1), matchLen, string(match[:matchLen]))
		buf = buf[i+matchLen:]
		off += i + matchLen
	}

	logger.Infof("digits processed: %d + %d digits",
		task.start, task.n)
	return nil
}

func worker(ctx context.Context, taskChan <-chan task, client obj.Client) {
	defer wg.Done()
	logger := logger.With("worker id", ctx.Value(workerContextKey("workerId")))
	defer logger.Sync()
	defer logger.Infow("worker exiting")

	logger.Info("worker started")
	b := retry.WithMaxRetries(3, retry.NewExponential(1*time.Second))
	for task := range taskChan {
		select {
		case <-ctx.Done():
			return
		default:
		}

		if err := retry.Do(ctx, b, func(ctx context.Context) error {
			if err := process(ctx, &task, logger, client); err != nil {
				return retry.RetryableError(err)
			}
			return nil
		}); err != nil {
			logger.Errorw("process failed", "error", err)
			task.cancel()
		}
	}

}

func main() {
	l, _ := zap.NewDevelopment()
	defer l.Sync()
	zap.ReplaceGlobals(l)
	logger = l.Sugar()

	start := flag.Int64("s", 0, "Start offset")
	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())
	client, err := gcs.NewClient(ctx)
	if err != nil {
		logger.Errorf("couldn't create a GCS client: %v", err)
		os.Exit(1)
	}
	defer client.Close()

	taskChan := make(chan task, 256)

	for i := 0; i < WORKERS; i++ {
		wg.Add(1)
		ctx = context.WithValue(ctx, workerContextKey("workerId"), i)
		go worker(ctx, taskChan, client)
	}

	for i := *start; i < index.Decimal.TotalDigits(); i += CHUNK_SIZE {
		task := task{
			start:  i,
			n:      CHUNK_SIZE,
			cancel: cancel,
		}
		taskChan <- task
		if ctx.Err() != nil {
			logger.Errorf("context error: %v", ctx.Err())
			break
		}
	}
	close(taskChan)
	wg.Wait()
}
