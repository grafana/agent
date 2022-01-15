// Package eventhandler watches for Kubernetes Event objects and ships them to
// a Loki sink
package eventhandler

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/signal"
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

type EventHandler struct {
	LokiClient    *logs.Instance
	Log           log.Logger
	CacheFile     *os.File
	LastEvent     *ShippedEvents
	InitEvent     *ShippedEvents
	EventInformer cache.SharedIndexInformer
	sync.Mutex
}

type ShippedEvents struct {
	// last event's timestamp
	Timestamp time.Time `json:"ts"`
	// map of event RVs already "shipped" (handed off) for last timestamp
	RvMap map[string]struct{} `json:"resourceVersion"`
}

// Initialize informer by setting event handler fns
func (eh *EventHandler) initInformer(eventsInformer cache.SharedIndexInformer) {
	eventsInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    eh.addEvent,
		UpdateFunc: eh.updateEvent,
		DeleteFunc: eh.deleteEvent,
	})
}

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

	labels, msg, err := handleEvent(e)
	if err != nil {
		level.Error(eh.Log).Log("msg", "Error processing event", "err", err, "event", e)
	}

	entry := newEntry(msg, eventTs, labels)
	// todo: config duration properly...
	if ok := eh.LokiClient.SendEntry(entry, time.Second); !ok {
		level.Error(eh.Log).Log("msg", "Error shipping event", "event", e)
		//todo: retry logic??
		return
	}
	level.Info(eh.Log).Log("msg", "Shipped entry", "eventRV", e.ResourceVersion, "eventMsg", e.Message)

	err = eh.updateLastEvent(e, eventTs)
	if err != nil {
		level.Error(eh.Log).Log("msg", "Could not write latest event to cache", "err", err)
	}
}

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
		level.Info(eh.Log).Log("msg", "Event RV didn't change, ignoring", "eRV", eNew.ResourceVersion)
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

	labels, msg, err := handleEvent(eNew)
	if err != nil {
		level.Error(eh.Log).Log("msg", "Error processing event", "err", err, "event", eNew)
	}

	entry := newEntry(msg, eventTs, labels)
	// todo: config duration properly...
	if ok := eh.LokiClient.SendEntry(entry, time.Second); !ok {
		level.Error(eh.Log).Log("msg", "Error shipping event", "event", eNew)
		//todo: retry logic??
		return
	}
	level.Info(eh.Log).Log("msg", "Shipped entry", "event", eNew.ResourceVersion, "eventMsg", eNew.Message)

	err = eh.updateLastEvent(eNew, eventTs)
	if err != nil {
		level.Error(eh.Log).Log("msg", "Could not write latest event to cache", "err", err)
	}
}

func (eh *EventHandler) updateLastEvent(e *v1.Event, eventTs time.Time) error {
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

	// if timestamp is different, write out the line and create
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
	eh.CacheFile.Seek(0, io.SeekEnd)
	buf, _ := json.Marshal(&eh.LastEvent)
	_, err := fmt.Fprintln(eh.CacheFile, string(buf))
	if err != nil {
		return err
	}
	return nil
}

// TODO: deprecate this in favor of just using promtail config on a JSON blob...TBD
func handleEvent(event *v1.Event) (model.LabelSet, string, error) {
	var msg strings.Builder
	labels := make(model.LabelSet)
	obj := event.InvolvedObject

	//todo: config
	labels[model.LabelName("eventhandler")] = model.LabelValue("ehtest0")

	if obj.Name == "" {
		return nil, "", fmt.Errorf("no involved object")
	}

	kindStr := strings.ToLower(obj.Kind)
	// TODO: match up labels with k8s integration so that log / metric correlation works
	// and users can use the same dashboard for both logs and metrics
	labels[model.LabelName("kind")] = model.LabelValue(obj.Kind)
	// this is used to enable correlation w/ K8s integration
	labels[model.LabelName(kindStr)] = model.LabelValue(obj.Name)
	// this won't increase cardinality but maybe improves UX?
	labels[model.LabelName("name")] = model.LabelValue(obj.Name)
	labels[model.LabelName("namespace")] = model.LabelValue(obj.Namespace)

	// we add these fields to the log line to reduce cardinality
	// TODO: is there a better way to do this?
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

	msg.WriteString(fmt.Sprintf("msg=%q", event.Message))

	return labels, msg.String(), nil
}

func (eh *EventHandler) deleteEvent(obj interface{}) {
	// do nothing...
	// todo: log this maybe??
}

func newEntry(msg string, ts time.Time, labels model.LabelSet) api.Entry {
	entry := logproto.Entry{Timestamp: ts, Line: msg}
	return api.Entry{Labels: labels, Entry: entry}
}

func getTimestamp(event *v1.Event) time.Time {
	if !event.LastTimestamp.IsZero() {
		return event.LastTimestamp.Time
	}
	return event.EventTime.Time
}

func readInitEvent(file *os.File, logger log.Logger) (*ShippedEvents, error) {
	var initEvent = new(ShippedEvents)

	stat, err := file.Stat()
	if err != nil {
		return nil, err
	}
	if stat.Size() == 0 {
		level.Info(logger).Log("msg", "Cache file empty, setting zero-valued initEvent")
		return initEvent, nil
	}

	// skip first newline
	var cur int64 = -1
	line := ""
	char := make([]byte, 1)
	size := stat.Size()
	for {
		cur -= 1
		file.Seek(cur, 2)
		file.Read(char)
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

func loadConfig(logger log.Logger) (*kubernetes.Clientset, error) {
	//todo: config
	kubeconfig := "/Users/coachjetha/.kube/config"
	var config *rest.Config
	var err error

	if len(kubeconfig) > 0 {
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
	} else {
		config, err = rest.InClusterConfig()
	}
	if err != nil {
		return nil, err
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return clientset, nil
}

func newEventHandler(l log.Logger, c integrations.Config, globals integrations.Globals) (integrations.Integration, error) {
	// cache file to store events shipped (prevents double shipping on restart)
	// todo: config
	cachePath := "latest.out"
	cacheFile, err := os.OpenFile(cachePath, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		level.Error(l).Log("msg", "Failed to open or create cache file", "err", err)
		os.Exit(1)
	}

	// attempt to read last event from cache file into the lastEvent struct
	initEvent, err := readInitEvent(cacheFile, l)
	if err != nil {
		level.Error(l).Log("msg", "Failed to read last event from cache file", "err", err)
		os.Exit(1)
	}

	clientset, err := loadConfig(l)
	if err != nil {
		level.Error(l).Log("msg", "Failed to load clientset", "err", err)
		os.Exit(1)
	}

	factory := informers.NewSharedInformerFactory(clientset, 2*time.Minute)
	eventInformer := factory.Core().V1().Events().Informer()

	eh := &EventHandler{
		LokiClient:    globals.Logs.Instance("default"),
		Log:           l,
		CacheFile:     cacheFile,
		InitEvent:     initEvent,
		EventInformer: eventInformer,
	}
	eh.initInformer(eventInformer)

	return eh, nil
}

func (eh *EventHandler) RunIntegration(ctx context.Context) error {
	// last op: sync & close cache file
	defer eh.CacheFile.Close()
	//todo do we need this?
	defer eh.CacheFile.Sync()
	// write out LastEvent struct to cache file
	defer eh.writeOutLastEvent()

	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	// do we need this?
	defer cancel()

	// start the informer
	eh.EventInformer.Run(ctx.Done())

	// die if caches don't sync properly
	isSynced := cache.WaitForCacheSync(ctx.Done(), eh.EventInformer.HasSynced)
	if !isSynced {
		level.Error(eh.Log).Log("msg", "Failed to sync informer cache")
		cancel()
	}

	<-ctx.Done()
	//todo: double check shutdown order of ops...
	// esp wrt:
	// loki client and draining channel
	// writing out file
	// shutting down informer, etc...
	return nil
}
