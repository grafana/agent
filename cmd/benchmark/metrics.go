package main

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/spf13/cobra"
)

type metric struct {
	name        string
	config      string
	port        int
	description string
}

var metricsList = map[string]metric{
	"normal": {
		name:        "normal",
		config:      "./configs/normal.river",
		port:        12346,
		description: "This is the normal configuration for the agent. It is used as a baseline for comparison.",
	},
	"relabel_large_cache": {
		name:        "relabel_large_cache",
		config:      "./configs/relabel_large_cache.river",
		port:        12347,
		description: "This configuration has a large relabel cache. It is used to test the performance of relabeling.",
	},
	"relabel_normal_cache": {
		name:        "relabel_normal_cache",
		config:      "./configs/relabel_normal_cache.river",
		port:        12348,
		description: "This configuration has the default label cache. It is used to test the performance of relabeling.",
	},
}

type metrics struct {
	name         string
	duration     time.Duration
	benchmark    string
	metricSource string
	networkDown  bool
}

func metricsCommand() *cobra.Command {
	f := &metrics{}
	cmd := &cobra.Command{
		Use:   "metrics [flags]",
		Short: "Run a set of benchmarks.",
		RunE: func(_ *cobra.Command, args []string) error {

			username := os.Getenv("PROM_USERNAME")
			if username == "" {
				panic("PROM_USERNAME env must be set")
			}
			password := os.Getenv("PROM_PASSWORD")
			if password == "" {
				panic("PROM_PASSWORD env must be set")
			}

			// Start the HTTP server, that can swallow requests.
			go httpServer()
			// Build the agent
			buildAgent()

			running := make(map[string]*exec.Cmd)
			test := startMetricsAgent()
			defer cleanupPid(test, "./data/test-data")
			networkdown = f.networkDown
			benchmarks := strings.Split(f.benchmark, ",")
			for _, b := range benchmarks {
				met, found := metricsList[b]
				if !found {
					return fmt.Errorf("unknown benchmark %q", b)
				}
				_ = os.RemoveAll("./data/" + met.name)

				_ = os.Setenv("NAME", f.name)
				_ = os.Setenv("HOST", fmt.Sprintf("localhost:%d", met.port))
				_ = os.Setenv("RUNTYPE", met.name)
				_ = os.Setenv("NETWORK_DOWN", strconv.FormatBool(f.networkDown))
				_ = os.Setenv("DISCOVERY", fmt.Sprintf("http://127.0.0.1:9001/api/v0/component/prometheus.test.metrics.%s/discovery", f.metricSource))
				agent := startNormalAgent(met)
				running[met.name] = agent
			}
			signalChannel := make(chan os.Signal, 1)
			signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM)
			t := time.NewTimer(f.duration)
			select {
			case <-t.C:
			case <-signalChannel:
			}
			for k, p := range running {
				cleanupPid(p, fmt.Sprintf("./data/%s", k))
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&f.name, "name", "n", f.name, "The name of the benchmark to run, this will be added to the exported metrics.")
	cmd.Flags().DurationVarP(&f.duration, "duration", "d", f.duration, "The duration to run the test for.")
	cmd.Flags().StringVarP(&f.metricSource, "type", "t", f.metricSource, "The type of metrics to use; single,man,churn,large or if you have added any to test.river they can be referenced.")
	cmd.Flags().StringVarP(&f.benchmark, "benchmarks", "b", f.benchmark, "List of benchmarks to run. Run `benchmark list` to list all possible benchmarks.")
	cmd.Flags().BoolVarP(&f.networkDown, "network-down", "a", f.networkDown, "If set to true, the network will be down for the duration of the test.")
	return cmd
}

func startNormalAgent(met metric) *exec.Cmd {
	cmd := exec.Command("./grafana-agent-flow", "run", met.config, fmt.Sprintf("--storage.path=./data/%s", met.name), fmt.Sprintf("--server.http.listen-addr=127.0.0.1:%d", met.port))
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Start()

	if err != nil {
		panic(err.Error())
	}
	return cmd
}

func startMetricsAgent() *exec.Cmd {
	cmd := exec.Command("./grafana-agent-flow", "run", "./configs/test.river", "--storage.path=./data/test-data", "--server.http.listen-addr=127.0.0.1:9001")
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	err := cmd.Start()
	if err != nil {
		panic(err.Error())
	}
	return cmd
}
