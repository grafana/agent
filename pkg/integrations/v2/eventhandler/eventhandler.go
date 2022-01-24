// Package eventhandler watches for Kubernetes Event objects and hands them off to
// Agent's Logs subsystem (embedded promtail)
package eventhandler

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"time"

	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/grafana/agent/pkg/integrations/v2"
	"github.com/grafana/agent/pkg/logs"
	"github.com/grafana/loki/clients/pkg/promtail/api"
	"github.com/grafana/loki/pkg/logproto"
	"github.com/prometheus/common/model"
)

const (
	cacheFileMode = 0600
)

// EventHandler watches for Kubernetes Event objects and hands them off to
// Agent's logs subsystem (embedded promtail).
type EventHandler struct {
	LokiClient    *logs.Instance
	Log           log.Logger
	CacheFile     *os.File
	CachePath     string
	LastEvent     *ShippedEvents
	InitEvent     *ShippedEvents
	EventInformer cache.SharedIndexInformer
	ClusterLabel  string
	SendTimeout   time.Duration
	sync.Mutex
}

// ShippedEvents stores a timestamp and map of event ResourceVersions shipped for that timestamp.
// Used to avoid double-shipping events upon restart.
type ShippedEvents struct {
	// shipped event's timestamp
	Timestamp time.Time `json:"ts"`
	// map of event RVs (resource versions) already "shipped" (handed off) for this timestamp.
	// this is to handle the case of a timestamp having multiple events,
	// which happens quite frequently.
	RvMap map[string]struct{} `json:"resourceVersion"`
}

func newEventHandler(l log.Logger, globals integrations.Globals, c *Config) (integrations.Integration, error) {
	var (
		config *rest.Config
		err    error
	)

	if c.InCluster {
		config, err = rest.InClusterConfig()
	} else {
		config, err = clientcmd.BuildConfigFromFlags("", c.KubeconfigPath)
	}
	if err != nil {
		return nil, err
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	// get an informer
	factory := informers.NewSharedInformerFactory(clientset, time.Duration(c.InformerResync)*time.Second)
	eventInformer := factory.Core().V1().Events().Informer()

	eh := &EventHandler{
		LokiClient:    globals.Logs.Instance(c.LogsInstance),
		Log:           l,
		CachePath:     c.CachePath,
		EventInformer: eventInformer,
		ClusterLabel:  c.ClusterName,
		SendTimeout:   time.Duration(c.SendTimeout) * time.Second,
	}
	// set the resource handler fns
	eh.initInformer(eventInformer)
	return eh, nil
}

// Initialize informer by setting event handler fns
func (eh *EventHandler) initInformer(eventsInformer cache.SharedIndexInformer) {
	eventsInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    eh.addEvent,
		UpdateFunc: eh.updateEvent,
		DeleteFunc: eh.deleteEvent,
	})
}

// Handles new event objects
func (eh *EventHandler) addEvent(obj interface{}) {
	e, ok := obj.(*v1.Event)
	if !ok {
		level.Error(eh.Log).Log("msg", "Object not v1.Event", "obj", obj)
		return
	}
	eventTs := getTimestamp(e)

	// if event is older than the one stored in cache on startup, we've shipped it
	if eventTs.Before(eh.InitEvent.Timestamp) {
		return
	}
	// if event is equal and is in map, we've shipped it
	if eventTs.Equal(eh.InitEvent.Timestamp) {
		if _, ok := eh.InitEvent.RvMap[e.ResourceVersion]; ok {
			return
		}
	}

	labels, msg, err := eh.handleEvent(e)
	if err != nil {
		level.Error(eh.Log).Log("msg", "Error processing event", "err", err, "event", e)
	}

	entry := newEntry(msg, eventTs, labels)
	ok = eh.LokiClient.SendEntry(entry, eh.SendTimeout)
	if !ok {
		err = fmt.Errorf("error handing entry off to promtail")
		level.Error(eh.Log).Log("err", err, "entry", entry)
		return
	}
	level.Info(eh.Log).Log("msg", "Shipped entry", "eventRV", e.ResourceVersion, "eventMsg", e.Message)

	// update cache with new "last" event
	err = eh.updateLastEvent(e, eventTs)
	if err != nil {
		level.Error(eh.Log).Log("msg", "Could not write latest event to cache", "err", err)
	}
}

// Handles event object updates. Note that this get triggered on informer resyncs and also
// events occurring more than once (in which case .count is incremented)
func (eh *EventHandler) updateEvent(objOld interface{}, objNew interface{}) {
	eOld, ok := objOld.(*v1.Event)
	if !ok {
		level.Error(eh.Log).Log("msg", "Object not v1.Event", "obj", eOld)
		return
	}

	eNew, ok := objNew.(*v1.Event)
	if !ok {
		level.Error(eh.Log).Log("msg", "Object not v1.Event", "obj", eNew)
		return
	}

	if eOld.GetResourceVersion() == eNew.GetResourceVersion() {
		// ignore resync updates
		level.Debug(eh.Log).Log("msg", "Event RV didn't change, ignoring", "eRV", eNew.ResourceVersion)
		return
	}

	eventTs := getTimestamp(eNew)
	// if event is older than the one stored in cache on startup, we've shipped it
	if eventTs.Before(eh.InitEvent.Timestamp) {
		return
	}
	// if event is equal and is in map, we've shipped it
	if eventTs.Equal(eh.InitEvent.Timestamp) {
		if _, ok := eh.InitEvent.RvMap[eNew.ResourceVersion]; ok {
			return
		}
	}

	labels, msg, err := eh.handleEvent(eNew)
	if err != nil {
		level.Error(eh.Log).Log("msg", "Error processing event", "err", err, "event", eNew)
	}

	entry := newEntry(msg, eventTs, labels)
	ok = eh.LokiClient.SendEntry(entry, eh.SendTimeout)
	if !ok {
		err = fmt.Errorf("error handing entry off to promtail")
		level.Error(eh.Log).Log("err", err, "entry", entry)
		return
	}
	level.Info(eh.Log).Log("msg", "Shipped entry", "eventRV", eNew.ResourceVersion, "eventMsg", eNew.Message)

	err = eh.updateLastEvent(eNew, eventTs)
	if err != nil {
		level.Error(eh.Log).Log("msg", "Could not write latest event to cache", "err", err)
	}
}

func (eh *EventHandler) deleteEvent(obj interface{}) {
	// do nothing...
	// todo: log this maybe??
}

// extract data from event fields and create labels, etc.
// TODO: ship JSON blobs and allow users to configure using pipelines etc.
// instead of hardcoding labels here
func (eh *EventHandler) handleEvent(event *v1.Event) (model.LabelSet, string, error) {
	var (
		msg    strings.Builder
		labels = make(model.LabelSet)
	)

	obj := event.InvolvedObject
	if obj.Name == "" {
		return nil, "", fmt.Errorf("no involved object for event")
	}

	labels[model.LabelName("source")] = model.LabelValue("eventhandler")
	labels[model.LabelName("cluster")] = model.LabelValue(eh.ClusterLabel)

	kindStr := strings.ToLower(obj.Kind)
	// TODO: match up labels with k8s integration so that log / metric correlation works
	// and users can use the same dashboard for both logs and metrics
	labels[model.LabelName("kind")] = model.LabelValue(obj.Kind)
	// this is used to enable correlation w/ K8s integration
	labels[model.LabelName(kindStr)] = model.LabelValue(obj.Name)
	// this won't increase cardinality but maybe improves UX?
	labels[model.LabelName("name")] = model.LabelValue(obj.Name)
	labels[model.LabelName("namespace")] = model.LabelValue(obj.Namespace)

	// we add these fields to the log line to reduce label bloat and cardinality
	if event.Action != "" {
		msg.WriteString(fmt.Sprintf("action=%s ", event.Type))
	}
	if obj.APIVersion != "" {
		msg.WriteString(fmt.Sprintf("apiversion=%s ", obj.APIVersion))
	}
	if obj.ResourceVersion != "" {
		msg.WriteString(fmt.Sprintf("resourceversion=%s ", obj.ResourceVersion))
	}

	// useful for debugging, can omit when shipping this
	if event.ResourceVersion != "" {
		msg.WriteString(fmt.Sprintf("evRV=%s ", event.ResourceVersion))
	}

	if event.ReportingInstance != "" {
		msg.WriteString(fmt.Sprintf("reportinginstance=%s ", event.ReportingInstance))
	}
	if event.ReportingController != "" {
		msg.WriteString(fmt.Sprintf("reportingcontroller=%s ", event.ReportingController))
	}
	if event.Source.Component != "" {
		msg.WriteString(fmt.Sprintf("sourcecomponent=%s ", event.Source.Component))
	}
	if event.Source.Host != "" {
		msg.WriteString(fmt.Sprintf("sourcehost=%s ", event.Source.Host))
	}
	if event.Reason != "" {
		msg.WriteString(fmt.Sprintf("reason=%s ", event.Reason))
	}
	if event.Type != "" {
		msg.WriteString(fmt.Sprintf("type=%s ", event.Type))
	}
	if event.Count != 0 {
		msg.WriteString(fmt.Sprintf("count=%d ", event.Count))
	}

	msg.WriteString(fmt.Sprintf("msg=%q", event.Message))

	return labels, msg.String(), nil
}

func getTimestamp(event *v1.Event) time.Time {
	if !event.LastTimestamp.IsZero() {
		return event.LastTimestamp.Time
	}
	return event.EventTime.Time
}

func newEntry(msg string, ts time.Time, labels model.LabelSet) api.Entry {
	entry := logproto.Entry{Timestamp: ts, Line: msg}
	return api.Entry{Labels: labels, Entry: entry}
}

// maintain "last event" state and write it out to disk as necessary
func (eh *EventHandler) updateLastEvent(e *v1.Event, eventTs time.Time) error {
	// resource handler fns don't run concurrently so we probably don't need this...
	eh.Lock()
	defer eh.Unlock()

	eventRv := e.ResourceVersion

	if eh.LastEvent == nil {
		// startup
		eh.LastEvent = &ShippedEvents{Timestamp: eventTs, RvMap: make(map[string]struct{})}
		eh.LastEvent.RvMap[eventRv] = struct{}{}
		return nil
	}

	// if timestamp is the same, add to map
	if eh.LastEvent != nil && eventTs.Equal(eh.LastEvent.Timestamp) {
		eh.LastEvent.RvMap[eventRv] = struct{}{}
		return nil
	}

	// if timestamp is different, write out the current last event and create
	// a new lastEvent struct
	err := eh.writeOutLastEvent()
	if err != nil {
		return err
	}
	eh.LastEvent = &ShippedEvents{Timestamp: eventTs, RvMap: make(map[string]struct{})}
	eh.LastEvent.RvMap[eventRv] = struct{}{}
	return nil
}

func (eh *EventHandler) writeOutLastEvent() error {
	if eh.LastEvent == nil {
		return nil
	}
	if _, err := eh.CacheFile.Seek(0, io.SeekEnd); err != nil {
		return err
	}
	buf, _ := json.Marshal(&eh.LastEvent)
	_, err := fmt.Fprintln(eh.CacheFile, string(buf))
	if err != nil {
		return err
	}
	return nil
}

// RunIntegration runs the eventhandler integration
func (eh *EventHandler) RunIntegration(ctx context.Context) error {
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()

	// todo: figure this out on K8s (PVC, etc.)
	cacheDir := filepath.Dir(eh.CachePath)
	// todo: correct perms here? / k8s config
	if err := os.MkdirAll(cacheDir, 0777); err != nil {
		level.Error(eh.Log).Log("msg", "Failed to create cache dir", "err", err)
		cancel()
	}

	// cache file to store events shipped (prevents double shipping on restart)
	cacheFile, err := os.OpenFile(eh.CachePath, os.O_RDWR|os.O_CREATE, cacheFileMode)
	if err != nil {
		level.Error(eh.Log).Log("msg", "Failed to open or create cache file", "err", err)
		cancel()
	}
	eh.CacheFile = cacheFile

	// attempt to read last timestamp from cache file into a ShippedEvents struct
	initEvent, err := readInitEvent(eh.CacheFile, eh.Log)
	if err != nil {
		level.Error(eh.Log).Log("msg", "Failed to read last event from cache file", "err", err)
		cancel()
	}
	eh.InitEvent = initEvent

	// sync & close cache file
	defer func() {
		if err := eh.CacheFile.Close(); err != nil {
			level.Error(eh.Log).Log("msg", "Failed to close cache file", "err", err)
			cancel()
		}
	}()
	// todo: do we need this?
	defer func() {
		if err := eh.CacheFile.Sync(); err != nil {
			level.Error(eh.Log).Log("msg", "Failed to sync cache file", "err", err)
			cancel()
		}
	}()
	// write out .LastEvent struct to cache file on exit
	defer func() {
		if err := eh.writeOutLastEvent(); err != nil {
			level.Error(eh.Log).Log("msg", "Failed to write out last event", "err", err)
			cancel()
		}
	}()

	// todo: is it worth even doing this?
	go func() {
		level.Info(eh.Log).Log("msg", "Waiting for cache to sync (initial List of events)")
		isSynced := cache.WaitForCacheSync(ctx.Done(), eh.EventInformer.HasSynced)
		if !isSynced {
			level.Error(eh.Log).Log("msg", "Failed to sync informer cache")
			// todo: do we want to bail here?
			return
		}
		level.Info(eh.Log).Log("msg", "Informer cache synced")
	}()

	// start the informer
	// technically we should prob use the factory here, but since we
	// only have one informer atm, this likely doesn't matter
	eh.EventInformer.Run(ctx.Done())

	<-ctx.Done()

	return nil
}

func readInitEvent(file *os.File, logger log.Logger) (*ShippedEvents, error) {
	var (
		initEvent = new(ShippedEvents)
		// skip first newline
		cur  int64 = -1
		char       = make([]byte, 1)
		line       = ""
	)

	stat, err := file.Stat()
	if err != nil {
		return nil, err
	}
	if stat.Size() == 0 {
		level.Info(logger).Log("msg", "Cache file empty, setting zero-valued initEvent")
		return initEvent, nil
	}

	size := stat.Size()
	for {
		cur--
		_, err = file.Seek(cur, 2)
		if err != nil {
			return nil, err
		}
		_, err = file.Read(char)
		if err != nil {
			return nil, err
		}
		// newline
		if char[0] == 10 {
			break
		}
		line = string(char) + line
		if cur == -size {
			break
		}
	}
	err = json.Unmarshal([]byte(line), &initEvent)
	if err != nil {
		return nil, err
	}
	level.Info(logger).Log("msg", "Loaded init event from cache file", "initEventTime", initEvent.Timestamp)
	return initEvent, nil
}
