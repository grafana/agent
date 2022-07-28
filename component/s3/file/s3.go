package s3

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	aws_config "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/grafana/agent/component"
	"github.com/grafana/agent/pkg/flow/rivertypes"
	"github.com/grafana/agent/pkg/river"
	"golang.org/x/net/context"
)

func init() {
	component.Register(component.Registration{
		Name:    "s3.file",
		Args:    Arguments{},
		Exports: Exports{},
	})
}

// Arguments implements the input for the s3 component
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

// AWSOptions implements specific AWS configuration options
type AWSOptions struct {
	AccessKey    string            `river:"key,attr,optional"`
	Secret       rivertypes.Secret `river:"secret,attr,optional"`
	Endpoint     string            `river:"endpoint,attr,optional"`
	DisableSSL   bool              `river:"disable_ssl,attr,optional"`
	UsePathStyle bool              `river:"use_path_style,attr,optional"`
}

var DefaultArguments = Arguments{
	PollFrequency: 10 * time.Minute,
}

// UnmarshalRiver implements the unmarshaller
func (a *Arguments) UnmarshalRiver(f func(v interface{}) error) error {
	*a = DefaultArguments
	type arguments Arguments
	return f((*arguments)(a))
}

// Exports implements the file content
type Exports struct {
	Content string `river:"content,attr"`
}

type S3 struct {
	mut        sync.Mutex
	opts       component.Options
	args       Arguments
	health     *component.Health
	watcher    *watcher
	updateChan chan []byte
	errorChan  chan error
	content    string
	cancel     context.CancelFunc
}

var (
	_ component.Component       = (*S3)(nil)
	_ component.HealthComponent = (*S3)(nil)
	_ river.Unmarshaler         = (*Arguments)(nil)
)

// New initializes the s3 component
func New(o component.Options, args Arguments) (*S3, error) {
	s3cfg, err := generateS3Config(args)
	if err != nil {
		return nil, err
	}

	s3Client := s3.NewFromConfig(*s3cfg, func(s3o *s3.Options) {
		s3o.UsePathStyle = args.Options.UsePathStyle
	})

	bucket, file := getPathBucketAndFile(args.Path)
	s := &S3{
		opts:       o,
		args:       args,
		health:     &component.Health{},
		updateChan: make(chan []byte),
		errorChan:  make(chan error),
	}

	w, err := newWatcher(bucket, file, s.updateChan, s.errorChan, args.PollFrequency, s3Client)
	if err != nil {
		return nil, err
	}
	s.watcher = w
	return s, nil
}

// Run activates the content handler and watcher
func (s *S3) Run(ctx context.Context) error {
	go s.handleContentUpdate(ctx)
	go s.watcher.run(ctx)
	<-ctx.Done()
	if s.cancel != nil {
		s.cancel()
	}

	return nil
}

// Update is called whenever the arguments have changed
func (s *S3) Update(args component.Arguments) error {
	newArgs := args.(Arguments)
	if newArgs.PollFrequency <= time.Second*30 {
		return fmt.Errorf("poll_frequency must be greater than 30s")
	}

	// TODO detect if args are different
	s3cfg, err := generateS3Config(newArgs)
	if err != nil {
		return nil
	}
	s3Client := s3.NewFromConfig(*s3cfg, func(s3o *s3.Options) {
		s3o.UsePathStyle = newArgs.Options.UsePathStyle
	})

	bucket, file := getPathBucketAndFile(newArgs.Path)

	s.mut.Lock()
	defer s.mut.Unlock()
	s.args = newArgs
	s.watcher.updateValues(bucket, file, newArgs.PollFrequency, s3Client)

	return nil
}

// CurrentHealth returns the health of the component
func (s *S3) CurrentHealth() component.Health {
	s.mut.Lock()
	defer s.mut.Unlock()
	return *s.health
}

func generateS3Config(args Arguments) (*aws.Config, error) {
	configOptions := make([]func(*aws_config.LoadOptions) error, 0)
	// Override the endpoint
	if args.Options.Endpoint != "" {
		endFunc := aws.EndpointResolverFunc(func(service, region string) (aws.Endpoint, error) {
			return aws.Endpoint{URL: args.Options.Endpoint}, nil
		})
		endResolver := aws_config.WithEndpointResolver(endFunc)
		configOptions = append(configOptions, endResolver)
	}

	// This incredibly nested option turns off ssl
	if args.Options.DisableSSL {
		httpOverride := aws_config.WithHTTPClient(
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
	// check credentials
	if args.Options.AccessKey != "" {
		if args.Options.Secret == "" {
			return nil, fmt.Errorf("if accesskey or secret are specified then the other must also be specified")
		}
		credFunc := aws.CredentialsProviderFunc(func(ctx context.Context) (aws.Credentials, error) {
			return aws.Credentials{
				AccessKeyID:     args.Options.AccessKey,
				SecretAccessKey: string(args.Options.Secret),
			}, nil
		})
		credProvider := aws_config.WithCredentialsProvider(credFunc)
		configOptions = append(configOptions, credProvider)
	}

	cfg, err := aws_config.LoadDefaultConfig(context.TODO(), configOptions...)
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}

// handleContentUpdate reads from the update and error channels setting as appropriate
func (s *S3) handleContentUpdate(ctx context.Context) {
	for {
		select {
		case buf := <-s.updateChan:
			s.mut.Lock()
			strBuf := string(buf)
			// Only update if changed
			if strBuf != s.content {
				s.content = string(buf)
				s.handleContentPolling(nil)
			}
			s.mut.Unlock()
		case err := <-s.errorChan:
			s.mut.Lock()
			s.handleContentPolling(err)
			s.mut.Unlock()

		case <-ctx.Done():
			return
		}
	}
}

func (s *S3) handleContentPolling(err error) {
		if err == nil {
s.opts.OnStateChange(&Exports{
				Content: s.content,
			})

		s.health.Health = component.HealthTypeHealthy
		s.health.Message = "s3 file updated"
	} else {
		s.health.Health = component.HealthTypeUnhealthy
		s.health.Message = err.Error()
	}
	s.health.UpdateTime = time.Now()
}

func getPathBucketAndFile(path string) (bucket, file string) {
	parts := strings.Split(path, "/")
	file = parts[len(parts)-1]
	bucket = strings.Join(parts[:len(parts)-1], "/")
	bucket = strings.ReplaceAll(bucket, "s3://", "")
	// TODO see if we can add some checks around the file/bucket
	return
}
