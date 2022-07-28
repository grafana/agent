package s3

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	aws_config "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go/aws/session"
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
	Path string `river:"path,attr"`
	// PollFrequency determines the frequency to check for changes
	// defaults to 5m
	PollFrequency time.Duration `river:"poll_frequency,attr"`
	// IsSecret determines if the content should be displayed to the user
	IsSecret bool `river:"is_secret,attr,optional"`
	// Options allows you to override default settings
	Options AWSOptions `river:"options,block,optional"`
}

type AWSOptions struct {
	AccessKey  string `river:"key,attr,optional"`
	Secret     Secret `river:"secret,attr,optional"`
	Endpoint   string `river:"endpoint,attr,optional"`
	DisableSSL bool   `river:"disable_ssl,attr,optional"`
}

type Exports struct {
	Content string `river:"content,attr"`
}

type S3 struct {
	opts component.Options

	mut        sync.Mutex
	args       Arguments
	health     *component.Health
	watcher    *watcher
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
	s3cfg, err := generateS3Config(args)
	if err != nil {
		return nil, err
	}
	s3Session := session.Must(session.NewSession(s3cfg))

	downloader := s3manager.NewDownloader(s3Session)
	bucket, file := getPathBucketAndFile(args.Path)
	s := &S3{
		opts:       o,
		args:       args,
		health:     &component.Health{},
		updateChan: make(chan []byte),
	}

	w, err := newWatcher(bucket, file, s.updateChan, args.PollFrequency, downloader)
	if err != nil {
		return nil, err
	}
	s.watcher = w
	return s, nil
}

func (s *S3) Run(ctx context.Context) error {
	go s.handleContentUpdate(ctx)
	go s.watcher.run(ctx)
	<-ctx.Done()
	if s.cancel != nil {
		s.cancel()
	}

	return nil
}

func generateS3Config(args Arguments) (*aws.Config, error) {
	configOptions := make([]aws_config.LoadOptionsFunc, 0)
	// Override the endpoint
	if args.Options.Endpoint != "" {
		endFunc := aws.EndpointResolverFunc(func(service, region string) (aws.Endpoint, error) {
			return aws.Endpoint{URL: args.Options.Endpoint}, nil
		})
		endResolver := aws_config.WithEndpointResolver(endFunc)
		configOptions = append(configOptions, endResolver)
	}

	// This incredibly nested option turns of ssl
	if args.Options.DisableSSL {
		httpOverride := config.WithHTTPClient(
			&http.Client{
				Transport: &http.Transport{
					TLSClientConfig: &tls.Config{
						InsecureSkipVerify: args.Options.DisableSSL,
					},
				},
			},
		)
		configOptions = append(configOptions, httpOverride)
	}
	// check credenentials
	if args.Options.AccessKey != "" {
		if args.Options.Secret == "" {
			return nil, fmt.Errorf("if accesskey or secret are specified then the other must also be specified")
		}
		credFunc := aws.CredentialsProviderFunc(func(ctx context.Context) (aws.Credentials, error) {
			return aws.Credentials{
				AccessKeyID:     args.Options.AccessKey,
				SecretAccessKey: args.Options.Secret,
			}, nil
		})
		credProvider := aws_config.WithCredentialsProvider(credFunc)
		configOptions = append(configOptions, credProvider)
	}

	cfg, err := aws_config.LoadDefaultConfig(context.Background(), configOptions[0])
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}

func (s *S3) Update(args component.Arguments) error {
	newArgs := args.(Arguments)
	if newArgs.PollFrequency <= time.Second*30 {
		return fmt.Errorf("poll_frequency must be greater than 30s")
	}

	s.mut.Lock()
	defer s.mut.Unlock()
	s.args = newArgs

	return nil
}

func getPathBucketAndFile(path string) (bucket, file string) {
	parts := strings.Split(path, "/")
	file = parts[len(parts)-1]
	bucket = strings.Join(parts[:len(parts)-1], "/")
	// TODO see if we can add some checks around the file/bucket
	return
}

func (s *S3) handleContentUpdate(ctx context.Context) {
	for {
		select {
		case buf := <-s.updateChan:
			s.mut.Lock()
			strBuf := string(buf)
			if strBuf != s.content {
				s.content = string(buf)
				s.handleContentPolling()
			}
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
