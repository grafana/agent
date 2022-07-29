package s3

import (
	"io"
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
	dlTicker   *time.Ticker
	downloader *s3.Client
}

func newWatcher(
	bucket, file string,
	out chan []byte,
	outError chan error,
	frequency time.Duration,
	downloader *s3.Client,
) *watcher {

	return &watcher{
		bucket:     bucket,
		file:       file,
		output:     out,
		outError:   outError,
		dlTicker:   time.NewTicker(frequency),
		downloader: downloader,
	}
}

func (w *watcher) updateValues(bucket, file string, frequency time.Duration, downloader *s3.Client) {
	w.mut.Lock()
	defer w.mut.Unlock()
	w.bucket = bucket
	w.file = file
	w.dlTicker.Reset(frequency)
	w.downloader = downloader
}

func (w *watcher) run(ctx context.Context) {
	err := w.download(ctx)
	if err != nil {
		w.outError <- err
	}
	defer w.dlTicker.Stop()
	for {
		select {
		case <-w.dlTicker.C:
			err = w.download(ctx)
			if err != nil {
				w.outError <- err
			}
		case <-ctx.Done():
			return
		}
	}
}

// download actually downloads the file from s3
func (w *watcher) download(ctx context.Context) error {
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
	_, err = output.Body.Read(buf)
	if err != nil && err != io.EOF {
		return err
	}
	select {
	case <-ctx.Done():
		return nil
	case w.output <- buf:
	}

	return nil
}
