package sd

import (
	"fmt"
	"strings"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/grafana/agent/component/pyroscope/ebpf/ebpfspy/metrics"
	lru "github.com/hashicorp/golang-lru/v2"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/model/labels"

	"github.com/grafana/agent/component/discovery"
)

const (
	labelContainerID    = "__container_id__"
	labelServiceName    = "service_name"
	labelServiceNameK8s = "__meta_kubernetes_pod_annotation_pyroscope_io_service_name"
	metricValue         = "process_cpu"
)

type Target struct {
	labels labels.Labels

	fingerprint           uint64
	fingerprintCalculated bool
}

func NewTarget(cid containerID, target discovery.Target) (*Target, error) {
	serviceName := target[labelServiceName]
	if serviceName == "" {
		serviceName = inferServiceName(target)
	}

	lset := make(map[string]string, len(target))
	for k, v := range target {
		if strings.HasPrefix(k, model.ReservedLabelPrefix) && k != labels.MetricName {
			continue
		}
		lset[k] = v
	}
	if lset[labels.MetricName] == "" {
		lset[labels.MetricName] = metricValue
	}
	if lset[labelServiceName] == "" {
		lset[labelServiceName] = serviceName
	}
	if cid != "" {
		lset[labelContainerID] = string(cid)
	}
	return &Target{
		labels: labels.FromMap(lset),
	}, nil
}

func inferServiceName(target discovery.Target) string {
	k8sServiceName := target[labelServiceNameK8s]
	if k8sServiceName != "" {
		return k8sServiceName
	}
	k8sNamespace := target["__meta_kubernetes_namespace"]
	k8sContainer := target["__meta_kubernetes_pod_container_name"]
	if k8sNamespace != "" && k8sContainer != "" {
		return fmt.Sprintf("ebpf/%s/%s", k8sNamespace, k8sContainer)
	}
	dockerContainer := target["__meta_docker_container_name"]
	if dockerContainer != "" {
		return dockerContainer
	}
	return "unspecified"
}

func (t *Target) Labels() (uint64, labels.Labels) {
	if !t.fingerprintCalculated {
		t.fingerprint = t.labels.Hash()
		t.fingerprintCalculated = true
	}
	return t.fingerprint, t.labels
}

type containerID string

type TargetFinder struct {
	l          log.Logger
	cid2target map[containerID]*Target

	// todo make it never evict during a reset
	containerIDCache *lru.Cache[uint32, containerID]
	defaultTarget    *Target
	metrics          *metrics.Metrics
}

func NewTargetFinder(l log.Logger, containerIdCacheSize int, metrics *metrics.Metrics) (*TargetFinder, error) {
	containerIDCache, err := lru.New[uint32, containerID](containerIdCacheSize)
	if err != nil {
		return nil, fmt.Errorf("containerIDCache create: %w", err)
	}
	return &TargetFinder{
		l:                l,
		containerIDCache: containerIDCache,
		metrics:          metrics,
	}, nil
}

type TargetsOptions struct {
	Targets            []discovery.Target
	TargetsOnly        bool
	DefaultTarget      discovery.Target
	ContainerCacheSize int
}

func (s *TargetFinder) SetTargets(opts TargetsOptions) {
	_ = level.Debug(s.l).Log("msg", "set targets", "count", len(opts.Targets))
	containerID2Target := make(map[containerID]*Target)
	for _, target := range opts.Targets {
		cid := containerIDFromTarget(target)
		if cid != "" {
			t, err := NewTarget(cid, target)
			if err != nil {
				_ = level.Error(s.l).Log(
					"msg", "target skipped",
					"target", target.Labels().String(),
					"err", err,
				)
				continue
			}
			containerID2Target[cid] = t
		}
	}
	if len(opts.Targets) > 0 && len(containerID2Target) == 0 {
		_ = level.Warn(s.l).Log("msg", "No container IDs found in targets")
	}
	s.cid2target = containerID2Target
	if opts.TargetsOnly {
		s.defaultTarget = nil
	} else {
		t, err := NewTarget("", opts.DefaultTarget)
		if err != nil {
			_ = level.Error(s.l).Log(
				"msg", "default target skipped",
				"target", opts.DefaultTarget.Labels().String(),
				"err", err,
			)
			s.defaultTarget = nil
		} else {
			s.defaultTarget = t
		}
	}
	_ = level.Debug(s.l).Log("msg", "created targets", "count", len(s.cid2target))
}

func (s *TargetFinder) FindTarget(pid uint32) *Target {
	res := s.findTarget(pid)
	if res != nil {
		return res
	}
	return s.defaultTarget
}

func (s *TargetFinder) findTarget(pid uint32) *Target {
	cid, ok := s.containerIDCache.Get(pid)
	if ok && cid != "" {
		return s.cid2target[cid]
	}

	cid = getContainerIDFromPID(pid)
	s.containerIDCache.Add(pid, cid)
	return s.cid2target[cid]
}

func (s *TargetFinder) ResizeContainerIDCache(size int) {
	s.containerIDCache.Resize(size)
}

func (s *TargetFinder) DebugInfo() []string {
	debugTargets := make([]string, 0, len(s.cid2target))
	for _, target := range s.cid2target {
		_, labels := target.Labels()
		debugTargets = append(debugTargets, labels.String())
	}
	return debugTargets
}

func (s *TargetFinder) Targets() []*Target {
	res := make([]*Target, 0, len(s.cid2target))
	for _, target := range s.cid2target {
		res = append(res, target)
	}
	return res
}

func containerIDFromTarget(target discovery.Target) containerID {
	cid, ok := target[labelContainerID]
	if ok && cid != "" {
		return containerID(cid)
	}
	cid, ok = target["__meta_kubernetes_pod_container_id"]
	if ok && cid != "" {
		return getContainerIDFromK8S(cid)
	}
	cid, ok = target["__meta_docker_container_id"]
	if ok && cid != "" {
		return containerID(cid)
	}
	return ""
}
