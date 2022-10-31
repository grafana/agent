package opamp

import (
	"context"
	"math/rand"
	"os"
	"runtime"
	"time"

	"github.com/go-kit/log"

	"github.com/go-kit/log/level"
	"github.com/oklog/ulid/v2"

	"github.com/open-telemetry/opamp-go/client"
	"github.com/open-telemetry/opamp-go/client/types"
	"github.com/open-telemetry/opamp-go/protobufs"
)

type OpAMP struct {
	logger          log.Logger
	instanceId      ulid.ULID
	effectiveConfig string

	agentDescription *protobufs.AgentDescription

	opampClient client.OpAMPClient

	remoteConfigStatus *protobufs.RemoteConfigStatus
}

func NewOpAMP(logger log.Logger) (*OpAMP, error) {
	o := &OpAMP{
		logger: logger,
	}

	o.createAgentIdentity()
	level.Debug(o.logger).Log("msg", "Agent starting", "id", o.instanceId.String())

	if err := o.start(); err != nil {
		return nil, err
	}
	return o, nil
}

func (o *OpAMP) start() error {
	o.opampClient = client.NewHTTP(nil)

	settings := types.StartSettings{
		OpAMPServerURL: "http://127.0.0.1:4320/v1/opamp",
		InstanceUid:    o.instanceId.String(),
		Callbacks: types.CallbacksStruct{
			OnConnectFunc: func() {
				level.Debug(o.logger).Log("msg", "Connected to the server.")
			},
			OnConnectFailedFunc: func(err error) {
				level.Error(o.logger).Log("msg", "Failed to connect to the server", "err", err)
			},
			OnErrorFunc: func(err *protobufs.ServerErrorResponse) {
				level.Error(o.logger).Log("msg", "Server returned an error response", "err", err.ErrorMessage)
			},
			SaveRemoteConfigStatusFunc: func(_ context.Context, status *protobufs.RemoteConfigStatus) {
				o.remoteConfigStatus = status
			},
			GetEffectiveConfigFunc: func(ctx context.Context) (*protobufs.EffectiveConfig, error) {
				return o.composeEffectiveConfig(), nil
			},
			OnMessageFunc: o.onMessage,
		},
		RemoteConfigStatus: o.remoteConfigStatus,
		Capabilities:       protobufs.AgentCapabilities_AgentCapabilities_AcceptsRemoteConfig,
	}
	err := o.opampClient.SetAgentDescription(o.agentDescription)
	if err != nil {
		return err
	}

	level.Debug(o.logger).Log("msg", "Starting OpAMP client...")

	err = o.opampClient.Start(context.Background(), settings)
	if err != nil {
		return err
	}

	level.Debug(o.logger).Log("msg", "OpAMP Client started.")

	return nil
}

func (o *OpAMP) createAgentIdentity() {
	// Generate instance id.
	entropy := ulid.Monotonic(rand.New(rand.NewSource(0)), 0)
	o.instanceId = ulid.MustNew(ulid.Timestamp(time.Now()), entropy)

	hostname, _ := os.Hostname()

	// Create Agent description.
	o.agentDescription = &protobufs.AgentDescription{
		IdentifyingAttributes: []*protobufs.KeyValue{
			{
				Key: "service.name",
				Value: &protobufs.AnyValue{
					Value: &protobufs.AnyValue_StringValue{StringValue: "grafana-agent"},
				},
			},
			{
				Key: "service.version",
				Value: &protobufs.AnyValue{
					Value: &protobufs.AnyValue_StringValue{StringValue: "last"},
				},
			},
		},
		NonIdentifyingAttributes: []*protobufs.KeyValue{
			{
				Key: "os.family",
				Value: &protobufs.AnyValue{
					Value: &protobufs.AnyValue_StringValue{
						StringValue: runtime.GOOS,
					},
				},
			},
			{
				Key: "host.name",
				Value: &protobufs.AnyValue{
					Value: &protobufs.AnyValue_StringValue{
						StringValue: hostname,
					},
				},
			},
		},
	}
}

func (o *OpAMP) updateAgentIdentity(instanceId ulid.ULID) {
	level.Debug(o.logger).Log("msg", "Agent identify is being changed from id=%v to id=%v",
		o.instanceId.String(),
		instanceId.String())
	o.instanceId = instanceId
}

func (o *OpAMP) composeEffectiveConfig() *protobufs.EffectiveConfig {
	return &protobufs.EffectiveConfig{
		ConfigMap: &protobufs.AgentConfigMap{
			ConfigMap: map[string]*protobufs.AgentConfigFile{
				"": {Body: []byte(o.effectiveConfig)},
			},
		},
	}
}

func (o *OpAMP) applyRemoteConfig(config *protobufs.AgentRemoteConfig) error {
	if config == nil {
		return nil
	}
	level.Debug(o.logger).Log("msg", "Received remote config from server", "hash", config.ConfigHash, "map", config.Config)
	return nil
}

func (o *OpAMP) Stop() {
	level.Debug(o.logger).Log("msg", "Agent shutting down...")
	if o.opampClient != nil {
		_ = o.opampClient.Stop(context.Background())
	}
}

func (o *OpAMP) onMessage(ctx context.Context, msg *types.MessageData) {
	if msg.RemoteConfig != nil {
		err := o.applyRemoteConfig(msg.RemoteConfig)
		if err != nil {
			o.opampClient.SetRemoteConfigStatus(&protobufs.RemoteConfigStatus{
				LastRemoteConfigHash: msg.RemoteConfig.ConfigHash,
				Status:               protobufs.RemoteConfigStatuses_RemoteConfigStatuses_FAILED,
				ErrorMessage:         err.Error(),
			})
		} else {
			o.opampClient.SetRemoteConfigStatus(&protobufs.RemoteConfigStatus{
				LastRemoteConfigHash: msg.RemoteConfig.ConfigHash,
				Status:               protobufs.RemoteConfigStatuses_RemoteConfigStatuses_APPLIED,
			})
		}
	}

	if msg.AgentIdentification != nil {
		newInstanceId, err := ulid.Parse(msg.AgentIdentification.NewInstanceUid)
		if err != nil {
			level.Error(o.logger).Log("err", err.Error())
		}
		o.updateAgentIdentity(newInstanceId)
	}
}
