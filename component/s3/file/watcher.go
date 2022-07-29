package s3

import (
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"golang.org/x/net/context"
)

type watcher struct {
	mut        sync.Mutex
	bucket     string
	file       string
	output     chan []byte
	outError   chan error
	frequency  time.Duration
	downloader *s3.Client
}

func newWatcher(
	bucket, file string,
	out chan []byte,
	outError chan error,
	frequency time.Duration,
	downloader *s3.Client,
) (*watcher, error) {
	return &watcher{
		bucket:     bucket,
		file:       file,
		output:     out,
		outError:   outError,
		frequency:  frequency,
		downloader: downloader,
	}, nil
}

func (w *watcher) updateValues(bucket, file string, frequency time.Duration, downloader *s3.Client) {
	w.mut.Lock()
	defer w.mut.Unlock()
	w.bucket = bucket
	w.file = file
	w.frequency = frequency
	w.downloader = downloader
}

func (w *watcher) run(ctx context.Context) {
	err := w.download()
	if err != nil {
		w.outError <- err
	}
	dlTick := time.NewTicker(w.frequency)
	currentFrequency := w.frequency
	defer dlTick.Stop()
	for {
		select {
		case <-dlTick.C:
			err = w.download()
			if err != nil {
				w.outError <- err
			}
			w.mut.Lock()
			if currentFrequency != w.frequency {
				currentFrequency = w.frequency
				dlTick.Reset(currentFrequency)
			}
			w.mut.Unlock()
		case <-ctx.Done():
			return
		}
	}
}

// download actually downloads the file from s3
func (w *watcher) download() error {
	w.mut.Lock()
	defer w.mut.Unlock()
	output, err := w.downloader.GetObject(context.Background(), &s3.GetObjectInput{
		Bucket: aws.String(w.bucket),
		Key:    aws.String(w.file),
	})
	if err != nil {
		return err
	}
	buf := make([]byte, output.ContentLength)
	output.Body.Read(buf)
	w.output <- buf
	return nil
}
