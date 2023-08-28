package main

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"syscall"
	"time"

	"github.com/gorilla/mux"
)

// main handles creating the benchmark.
func main() {

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

	// Get it warmed up
	startRun(100, true, 1*time.Minute)

	startRun(1_000, true, 60*time.Minute)
	startRun(1_000, false, 60*time.Minute)

	startRun(10_000, true, 60*time.Minute)
	startRun(10_000, false, 60*time.Minute)
	startRun(100_000, true, 60*time.Minute)
	startRun(100_000, true, 60*time.Minute)
}

func startRun(metricCount int, allowWAL bool, run time.Duration) {
	os.RemoveAll("./simple-data")
	os.RemoveAll("./old-data")
	allow = allowWAL
	// Do 1_000 run with WAL for 60 minutes
	avalanche := startAvalanche(metricCount)
	defer func() {
		err := avalanche.Process.Kill()
		if err != nil {
			println(err.Error())
		}
		defer syscall.Kill(-avalanche.Process.Pid, syscall.SIGKILL)

	}()
	_ = os.Setenv("METRIC_COUNT", strconv.Itoa(metricCount))
	_ = os.Setenv("ALLOW_WAL", strconv.FormatBool(allowWAL))

	simple := startSimpleAgent()
	defer syscall.Kill(-simple.Process.Pid, syscall.SIGKILL)
	defer os.RemoveAll("./simple-data")
	old := startOldAgent()
	defer syscall.Kill(-old.Process.Pid, syscall.SIGKILL)
	defer os.RemoveAll("./old-data")

	time.Sleep(run)
}

func buildAgent() {
	cmd := exec.Command("go", "build", "../../../../../cmd/grafana-agent-flow")
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		panic(err.Error())
	}
}

func startAvalanche(metricCount int) *exec.Cmd {
	cmd := exec.Command("/bin/sh", "-c", fmt.Sprintf("docker run -p 9001:9001 quay.io/freshtracks.io/avalanche --metric-count=%d", metricCount))
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Start()
	if err != nil {
		panic(err.Error())
	}
	return cmd
}

func startSimpleAgent() *exec.Cmd {
	cmd := exec.Command("./grafana-agent-flow", "run", "./simple.river", "--storage.path=./simple-data")
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	f, _ := os.OpenFile("./simple-log.txt", os.O_APPEND|os.O_CREATE, 0666)
	cmd.Stdout = f
	err := cmd.Start()
	if err != nil {
		panic(err.Error())
	}
	return cmd
}

func startOldAgent() *exec.Cmd {
	cmd := exec.Command("./grafana-agent-flow", "run", "./rw.river", "--storage.path=./old-data", "--server.http.listen-addr=127.0.0.1:12346")
	f, _ := os.OpenFile("./old-log.txt", os.O_APPEND|os.O_CREATE, 0666)
	cmd.Stdout = f
	err := cmd.Start()
	if err != nil {
		panic(err.Error())
	}
	return cmd
}

var allow = false
var index = 0

func httpServer() {
	r := mux.NewRouter()
	r.HandleFunc("/post", func(w http.ResponseWriter, r *http.Request) {
		handlePost(w, r)
	})
	r.HandleFunc("/allow", func(w http.ResponseWriter, r *http.Request) {
		println("allowing")
		allow = true
	})
	r.HandleFunc("/block", func(w http.ResponseWriter, r *http.Request) {
		println("blocking")
		allow = false
	})
	http.Handle("/", r)
	println("Starting server")
	err := http.ListenAndServe(":8888", nil)
	if err != nil {
		println(err)
	}
}

func handlePost(w http.ResponseWriter, r *http.Request) {
	index++
	println(fmt.Sprintf("index %d", index))
	if allow {
		println(fmt.Sprintf("Body context is %d", r.ContentLength))
		return
	} else {
		println("returning 500")
		w.WriteHeader(500)
	}
}
