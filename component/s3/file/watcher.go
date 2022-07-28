package s3

import (
	"time"

	aws_v1 "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"golang.org/x/net/context"
)

type watcher struct {
	bucket     string
	file       string
	output     chan ([]byte)
	frequency  time.Duration
	downloader *s3manager.Downloader
}

func newWatcher(
	bucket, file string,
	out chan ([]byte),
	frequency time.Duration,
	downloader *s3manager.Downloader,
) (*watcher, error) {
	return &watcher{
		bucket:     bucket,
		file:       file,
		output:     out,
		frequency:  frequency,
		downloader: downloader,
	}, nil
}

func (w *watcher) run(ctx context.Context) {
	download(w.bucket, w.file, w.output, w.downloader)
	timer := time.NewTimer(w.frequency)
	defer timer.Stop()
	for {
		select {
		case <-timer.C:
			download(w.bucket, w.file, w.output, w.downloader)
		case <-ctx.Done():
			return
		}
	}
}

// download actually downloads the file from s3
func download(bucket, file string, out chan ([]byte), downloader *s3manager.Downloader) error {
	buf := aws_v1.NewWriteAtBuffer([]byte{})
	_, err := downloader.Download(buf, &s3.GetObjectInput{
		Bucket: aws_v1.String(bucket),
		Key:    aws_v1.String(file),
	})
	if err != nil {
		return err
	}
	out <- buf.Bytes()
	return nil
}
