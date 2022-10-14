package supportbundle

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"runtime"
	"runtime/pprof"
	"time"

	"github.com/grafana/agent/pkg/build"
	"github.com/grafana/agent/pkg/config"
	"github.com/grafana/agent/pkg/server"
	"github.com/mackerelio/go-osstat/uptime"
	"gopkg.in/yaml.v3"
)

// Bundle collects all the data that is exposed as a support bundle.
type Bundle struct {
	meta                  []byte
	config                []byte
	agentMetrics          []byte
	agentMetricsInstances []byte
	agentMetricsTargets   []byte
	agentLogsInstances    []byte
	heapBuf               *bytes.Buffer
	goroutineBuf          *bytes.Buffer
	blockBuf              *bytes.Buffer
	mutexBuf              *bytes.Buffer
	cpuBuf                *bytes.Buffer
}

// Metadata contains general runtime information about the current Agent.
type Metadata struct {
	BuildVersion    string   `yaml:"build_version"`
	OS              string   `yaml:"os"`
	Architecture    string   `yaml:"architecture"`
	Uptime          float64  `yaml:"uptime"`
	EnabledFeatures []string `yaml:"enabled_features"`
}

// Export gathers the information required for the support bundle.
func Export(enabledFeatures []string, cfg config.Config, srvAddress string, duration float64, dialContext server.DialContextFunc) (*Bundle, error) {
	// Gather runtime metadata.
	ut, err := uptime.Get()
	if err != nil {
		return nil, err
	}
	m := Metadata{
		BuildVersion:    build.Version,
		OS:              runtime.GOOS,
		Architecture:    runtime.GOARCH,
		Uptime:          ut.Seconds(),
		EnabledFeatures: enabledFeatures,
	}
	meta, err := yaml.Marshal(m)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal support bundle metadata: %s", err)
	}
	// Gather current configuration.
	config, err := yaml.Marshal(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal config: %s", err)
	}

	var httpClient http.Client
	httpClient.Transport = &http.Transport{DialContext: dialContext}
	// Gather Agent's own metrics.
	resp, err := httpClient.Get("http://" + srvAddress + "/metrics")
	if err != nil {
		return nil, fmt.Errorf("failed to get internal Agent metrics: %s", err)
	}
	agentMetrics, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read internal Agent metrics: %s", err)
	}

	// Collect the Agent metrics instances and target statuses.
	resp, err = httpClient.Get("http://" + srvAddress + "/agent/api/v1/metrics/instances")
	if err != nil {
		return nil, fmt.Errorf("failed to get internal Agent metrics: %s", err)
	}
	agentMetricsInstances, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read internal Agent metrics: %s", err)
	}
	resp, err = httpClient.Get("http://" + srvAddress + "/agent/api/v1/metrics/targets")
	if err != nil {
		return nil, fmt.Errorf("failed to get Agent metrics targets: %s", err)
	}
	agentMetricsTargets, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read Agent metrics targets: %s", err)
	}

	// Collect the Agent's logs instances and target statuses.
	resp, err = httpClient.Get("http://" + srvAddress + "/agent/api/v1/logs/instances")
	if err != nil {
		return nil, fmt.Errorf("failed to get Agent logs instances: %s", err)
	}
	agentLogsInstances, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read Agent logs instances: %s", err)
	}

	// TODO(@tpaschalis) Add back after grafana/agent@2175 is resolved, as it
	// currently results in a panic.
	// resp, err = http.DefaultClient.Get("http://" + srvAddress + "/agent/api/v1/logs/targets")
	// if err != nil {
	// 	return fmt.Errorf("failed to get  Agent logs targets: %s", err)
	// }
	// agentLogsTargets, err := io.ReadAll(resp.Body)
	// if err != nil {
	// 	return fmt.Errorf("failed to read internal Agent metrics: %s", err)
	// }

	// Export pprof data.
	var (
		heapBuf      bytes.Buffer
		goroutineBuf bytes.Buffer
		blockBuf     bytes.Buffer
		mutexBuf     bytes.Buffer
		cpuBuf       bytes.Buffer
	)
	// TODO(@tpaschalis) Since these are the built-in profiles, do we actually
	// need the nil check?
	if p := pprof.Lookup("heap"); p != nil {
		if err := p.WriteTo(&heapBuf, 0); err != nil {
			return nil, err
		}
	}
	if p := pprof.Lookup("goroutine"); p != nil {
		if err := p.WriteTo(&goroutineBuf, 0); err != nil {
			return nil, err
		}
	}
	runtime.SetBlockProfileRate(1)
	if p := pprof.Lookup("block"); p != nil {
		if err := p.WriteTo(&blockBuf, 0); err != nil {
			return nil, err
		}
	}
	runtime.SetBlockProfileRate(0)

	runtime.SetMutexProfileFraction(1)
	if p := pprof.Lookup("mutex"); p != nil {
		if err := p.WriteTo(&mutexBuf, 0); err != nil {
			return nil, err
		}
	}
	runtime.SetMutexProfileFraction(0)

	// TODO(@tpaschalis) Figure out how to better correlate CPU profile
	// duration to server timeout settings. Also, ideally a CPU profile should
	// include at least one scrape or some log collection, but we can't
	// guarantee that.
	err = pprof.StartCPUProfile(&cpuBuf)
	if err != nil {
		return nil, err
	}
	time.Sleep(time.Duration(duration-1) * time.Second)
	pprof.StopCPUProfile()

	// Finally, bundle everything up to be served, either as a zip from
	// memory, or exported to a directory.
	bundle := &Bundle{
		meta:                  meta,
		config:                config,
		agentMetrics:          agentMetrics,
		agentMetricsInstances: agentMetricsInstances,
		agentMetricsTargets:   agentMetricsTargets,
		agentLogsInstances:    agentLogsInstances,
		heapBuf:               &heapBuf,
		goroutineBuf:          &goroutineBuf,
		blockBuf:              &blockBuf,
		mutexBuf:              &mutexBuf,
		cpuBuf:                &cpuBuf,
		// agentLogsTargets:   agentLogsTargets,
	}

	return bundle, nil
}

// Serve the collected data and logs as a zip file over the given
// http.ResponseWriter.
func Serve(rw http.ResponseWriter, b *Bundle, logsBuf *bytes.Buffer) error {
	zw := zip.NewWriter(rw)
	rw.Header().Set("Content-Type", "application/zip")
	rw.Header().Set("Content-Disposition", "attachment; filename=\"agent-support-bundle.zip\"")

	if err := writeByteSlice(zw, b.meta, "agent-support-bundle", "agent-metadata.yaml"); err != nil {
		return err
	}
	if err := writeByteSlice(zw, b.config, "agent-support-bundle", "agent-config.yaml"); err != nil {
		return err
	}
	if err := writeByteSlice(zw, b.agentMetrics, "agent-support-bundle", "agent-metrics.txt"); err != nil {
		return err
	}
	if err := writeByteSlice(zw, b.agentMetricsInstances, "agent-support-bundle", "agent-metrics-instances.json"); err != nil {
		return err
	}
	if err := writeByteSlice(zw, b.agentMetricsTargets, "agent-support-bundle", "agent-metrics-targets.json"); err != nil {
		return err
	}
	if err := writeByteSlice(zw, b.agentLogsInstances, "agent-support-bundle", "agent-logs-instances.json"); err != nil {
		return err
	}

	if err := writeBytesBuff(zw, logsBuf, "agent-support-bundle", "agent-logs.txt"); err != nil {
		return err
	}

	if err := writeBytesBuff(zw, b.cpuBuf, "agent-support-bundle", "pprof", "cpu.pprof"); err != nil {
		return err
	}
	if err := writeBytesBuff(zw, b.heapBuf, "agent-support-bundle", "pprof", "heap.pprof"); err != nil {
		return err
	}
	if err := writeBytesBuff(zw, b.goroutineBuf, "agent-support-bundle", "pprof", "goroutine.pprof"); err != nil {
		return err
	}
	if err := writeBytesBuff(zw, b.mutexBuf, "agent-support-bundle", "pprof", "mutex.pprof"); err != nil {
		return err
	}
	if err := writeBytesBuff(zw, b.blockBuf, "agent-support-bundle", "pprof", "block.pprof"); err != nil {
		return err
	}

	err := zw.Close()
	if err != nil {
		return fmt.Errorf("failed to flush the zip writer: %v", err)
	}
	return nil
}

func writeByteSlice(zw *zip.Writer, b []byte, fn ...string) error {
	f, err := zw.Create(filepath.Join(fn...))
	if err != nil {
		return err
	}
	_, err = f.Write(b)
	if err != nil {
		return err
	}
	return nil
}

func writeBytesBuff(zw *zip.Writer, b *bytes.Buffer, fn ...string) error {
	f, err := zw.Create(filepath.Join(fn...))
	if err != nil {
		return err
	}
	_, err = io.Copy(f, b)
	if err != nil {
		return err
	}
	return nil
}
