package smoke

import (
	"context"
	"testing"

	"github.com/go-kit/log"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
)

// Note: these tests are mostly worthless at this point
// but would allow easy debugging of tasks as they become more
// complex. Using https://pkg.go.dev/k8s.io/client-go/testing#ObjectTracker
// to mock responses from the fake client is also possible.

func Test_deletePodBySelectorTask_Run(t1 *testing.T) {
	type fields struct {
		logger    log.Logger
		clientset kubernetes.Interface
		namespace string
		selector  string
	}
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "deletePodBySelectorTask",
			fields: fields{
				logger:    log.NewNopLogger(),
				clientset: fake.NewSimpleClientset(),
				namespace: "foo",
				selector:  "foo=bar",
			},
		},
	}
	for _, tt := range tests {
		t1.Run(tt.name, func(t1 *testing.T) {
			t := &deletePodBySelectorTask{
				logger:    tt.fields.logger,
				clientset: tt.fields.clientset,
				namespace: tt.fields.namespace,
				selector:  tt.fields.selector,
			}
			if err := t.Run(tt.args.ctx); (err != nil) != tt.wantErr {
				t1.Errorf("Run() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_deletePodTask_Run(t1 *testing.T) {
	type fields struct {
		logger    log.Logger
		clientset kubernetes.Interface
		namespace string
		pod       string
	}
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "deletePodTask",
			fields: fields{
				logger:    log.NewNopLogger(),
				clientset: fake.NewSimpleClientset(),
				namespace: "foo",
				pod:       "bar",
			},
		},
	}
	for _, tt := range tests {
		t1.Run(tt.name, func(t1 *testing.T) {
			t := &deletePodTask{
				logger:    tt.fields.logger,
				clientset: tt.fields.clientset,
				namespace: tt.fields.namespace,
				pod:       tt.fields.pod,
			}
			if err := t.Run(tt.args.ctx); (err != nil) != tt.wantErr {
				t1.Errorf("Run() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_scaleDeploymentTask_Run(t1 *testing.T) {
	type fields struct {
		logger      log.Logger
		clientset   kubernetes.Interface
		namespace   string
		deployment  string
		maxReplicas int
		minReplicas int
	}
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "scaleDeploymentTask",
			fields: fields{
				logger:      log.NewNopLogger(),
				clientset:   fake.NewSimpleClientset(),
				namespace:   "foo",
				deployment:  "bar",
				maxReplicas: 11,
				minReplicas: 2,
			},
		},
	}
	for _, tt := range tests {
		t1.Run(tt.name, func(t1 *testing.T) {
			t := &scaleDeploymentTask{
				logger:      tt.fields.logger,
				clientset:   tt.fields.clientset,
				namespace:   tt.fields.namespace,
				deployment:  tt.fields.deployment,
				maxReplicas: tt.fields.maxReplicas,
				minReplicas: tt.fields.minReplicas,
			}
			if err := t.Run(tt.args.ctx); (err != nil) != tt.wantErr {
				t1.Errorf("Run() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
