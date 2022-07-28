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
	err := w.download(w.bucket, w.file, w.output, w.downloader)
	if err != nil {
		w.outError <- err
	}
	timer := time.NewTimer(w.frequency)
	currentFrequency := w.frequency
	defer timer.Stop()
	for {
		select {
		case <-timer.C:
			err = w.download(w.bucket, w.file, w.output, w.downloader)
			if err != nil {
				w.outError <- err
			}
			if currentFrequency != w.frequency {
				currentFrequency = w.frequency
				timer.Stop()
				timer = time.NewTimer(w.frequency)
			}
		case <-ctx.Done():
			return
		}
	}
}

// download actually downloads the file from s3
func (w *watcher) download(bucket, file string, out chan []byte, downloader *s3.Client) error {
	w.mut.Lock()
	defer w.mut.Unlock()
	output, err := downloader.GetObject(context.Background(), &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(file),
	})
	if err != nil {
		return err
	}
	buf := make([]byte, output.ContentLength)
	output.Body.Read(buf)
	out <- buf
	return nil
}
