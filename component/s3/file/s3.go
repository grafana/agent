package s3

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	aws_v1 "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/grafana/agent/component"
	"golang.org/x/net/context"
)

func init() {
	component.Register(component.Registration{
		Name:    "s3.file",
		Args:    Arguments{},
		Exports: Exports{},
	})
}

type Arguments struct {
	Path string `hcl:"path,attr"`
	// PollFrequency determines the frequency to check for changes
	// defaults to 5m
	PollFrequency time.Duration `hcl:"poll_frequency"`
	// IsSecret determines if the content should be displayed to the user
	IsSecret bool `hcl:"is_secret,optional"`
}

type Exports struct {
	Content string `hcl:"content,attr"`
}

type S3 struct {
	opts component.Options

	mut        sync.Mutex
	args       Arguments
	health     *component.Health
	s3Bucket   string
	s3FileName string
	downloader *s3manager.Downloader

	updateChan chan ([]byte)
	content    string
	cancel     context.CancelFunc
}

var (
	_ component.Component       = (*S3)(nil)
	_ component.HealthComponent = (*S3)(nil)
)

func New(o component.Options, args Arguments) (*S3, error) {
	// session.NewSession utilizes combo of environment vars, config files
	// and all other default S3/AWS configuration values
	s3Session := session.Must(session.NewSession())
	downloader := s3manager.NewDownloader(s3Session)
	s := &S3{
		opts:       o,
		args:       args,
		downloader: downloader,
		health:     &component.Health{},
		updateChan: make(chan []byte),
	}
	return s, nil
}

func (s *S3) Run(ctx context.Context) error {
	<-ctx.Done()
	if s.cancel != nil {
		s.cancel()
	}

	return nil
}

func (s *S3) Update(args component.Arguments) error {
	newArgs := args.(Arguments)
	if newArgs.PollFrequency <= time.Second*30 {
		return fmt.Errorf("poll_frequency must be greater than 30s")
	}

	parts := strings.Split(newArgs.Path, "/")
	file := parts[len(parts)-1]
	bucket := strings.Join(parts[:len(parts)-1], "/")
	// todo see if we can add some checks around the file/bucket

	s.mut.Lock()
	defer s.mut.Unlock()

	s.s3Bucket = bucket
	s.s3FileName = file
	if s.cancel != nil {
		s.cancel()
	}
	ctx := context.Background()
	ctx, s.cancel = context.WithCancel(ctx)
	s.args = newArgs

	go watcher(s.s3Bucket, s.s3FileName, s.updateChan, s.args.PollFrequency, s.downloader, ctx)
	go s.handleUpdate(ctx)
	return nil

}

func (s *S3) handleUpdate(ctx context.Context) {
	for {
		select {
		case buf := <-s.updateChan:
			s.mut.Lock()
			s.content = string(buf)
			s.handleContentPolling()
			s.mut.Unlock()
		case <-ctx.Done():
			return
		}
	}
}

func (s *S3) handleContentPolling() {
	s.opts.OnStateChange(&Exports{
		Content: s.content,
	})
	s.health.Health = component.HealthTypeHealthy
	s.health.Message = "s3 file updated"
	s.health.UpdateTime = time.Now()

}

func (s *S3) CurrentHealth() component.Health {
	s.mut.Lock()
	defer s.mut.Unlock()
	return *s.health
}

// watcher kicks off watching a file path for the duration specified
func watcher(bucket, file string, out chan ([]byte), frequency time.Duration, downloader *s3manager.Downloader, ctx context.Context) {
	download(bucket, file, out, downloader)
	timer := time.NewTimer(frequency)
	defer timer.Stop()
	for {
		select {
		case <-timer.C:
			download(bucket, file, out, downloader)
		case <-ctx.Done():
			return
		}
	}
}

// download actually downloads the file from s3
func download(bucket, file string, out chan ([]byte), downloader *s3manager.Downloader) error {
	buf := aws_v1.NewWriteAtBuffer([]byte{})
	_, err := downloader.Download(buf, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(file),
	})
	if err != nil {
		return err
	}
	out <- buf.Bytes()
	return nil
}
