package sd

import (
	"strings"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/model/labels"

	"github.com/grafana/agent/component/discovery"
)

const (
	labelContainerID = "__container_id__"
	labelServiceName = "service_name"
	metricValue      = "process_cpu"
)

type Target struct {
	labels labels.Labels

	fingerprint           uint64
	fingerprintCalculated bool
}

func NewTarget(cid containerID, defaultServiceName string, target discovery.Target) *Target {
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
		lset[labelServiceName] = defaultServiceName
	}
	if cid != "" {
		lset[labelContainerID] = string(cid)
	}
	return &Target{
		labels: labels.FromMap(lset),
	}
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
	// todo make it configurable lru
	pid2cid       map[uint32]containerID
	defaultTarget *Target
}

func NewTargetFinder(l log.Logger) *TargetFinder {
	return &TargetFinder{
		l:       l,
		pid2cid: make(map[uint32]containerID),
	}
}

type Options struct {
	DefaultServiceName string
	Targets            []discovery.Target
	TargetsOnly        bool
	DefaultTarget      discovery.Target
}

func (s *TargetFinder) SetTargets(opts Options) {
	containerID2Target := make(map[containerID]*Target)
	for _, target := range opts.Targets {
		cid := containerIDFromTarget(target)
		if cid != "" {
			defaultServiceName := defaultServiceName(opts.DefaultServiceName, target)
			containerID2Target[cid] = NewTarget(cid, defaultServiceName, target)
		}
	}
	if len(opts.Targets) > 0 && len(containerID2Target) == 0 {
		_ = level.Warn(s.l).Log("msg", "No container IDs found in targets")
	}
	s.cid2target = containerID2Target
	if opts.TargetsOnly {
		s.defaultTarget = nil
	} else {
		s.defaultTarget = NewTarget("", opts.DefaultServiceName, opts.DefaultTarget)
	}
}

func (s *TargetFinder) FindTarget(pid uint32) *Target {
	res := s.findTarget(pid)
	if res != nil {
		return res
	}
	return s.defaultTarget
}

func (s *TargetFinder) findTarget(pid uint32) *Target {
	cid, ok := s.pid2cid[pid]
	if ok && cid != "" {
		return s.cid2target[cid]
	}
	if len(s.pid2cid) > 1024 { // todo make it configurable lru
		s.pid2cid = make(map[uint32]containerID)
	}
	cid = getContainerIDFromPID(pid)
	s.pid2cid[pid] = cid
	return s.cid2target[cid]
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

func defaultServiceName(fallback string, target discovery.Target) string {
	serviceName := target[labelServiceName]
	if serviceName != "" {
		return serviceName
	}

	k8sNamespace := target["__meta_kubernetes_namespace"]
	k8sPod := target["__meta_kubernetes_pod_name"]
	if k8sNamespace != "" && k8sPod != "" {
		return k8sNamespace + "/" + k8sPod
	}

	dockerContainerName := target["__meta_docker_container_name"]
	if dockerContainerName != "" {
		return dockerContainerName
	}

	return fallback
}
