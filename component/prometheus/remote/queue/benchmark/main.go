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

	metricCount := os.Args[1]
	allowWal := os.Args[2]
	duration := os.Args[3]
	metrics, _ := strconv.Atoi(metricCount)
	allowWalBool, _ := strconv.ParseBool(allowWal)
	parsedDuration, _ := time.ParseDuration(duration)
	fmt.Println(metrics, allowWalBool, parsedDuration)
	startRun(metrics, allowWalBool, parsedDuration)

}

func startRun(metricCount int, allowWAL bool, run time.Duration) {
	os.RemoveAll("./queue-data")
	os.RemoveAll("./old-data")
	os.RemoveAll("./queue-data")
	allow = allowWAL
	_ = os.Setenv("METRIC_COUNT", strconv.Itoa(metricCount))
	_ = os.Setenv("ALLOW_WAL", strconv.FormatBool(allowWAL))

	old := startOldAgent()
	fmt.Println("starting old agent")
	defer old.Process.Kill()
	defer old.Process.Release()
	defer old.Wait()
	defer syscall.Kill(-old.Process.Pid, syscall.SIGKILL)
	defer os.RemoveAll("./old-data")

	queue := startQueueAgent()
	fmt.Println("starting queue agent")
	defer queue.Process.Kill()
	defer queue.Process.Release()
	defer queue.Wait()
	defer syscall.Kill(-queue.Process.Pid, syscall.SIGKILL)
	defer os.RemoveAll("./queue-data")

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

func startQueueAgent() *exec.Cmd {
	cmd := exec.Command("./grafana-agent-flow", "run", "./queue.river", "--storage.path=~/bench/queue-data", "--server.http.listen-addr=127.0.0.1:12347")
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Start()

	if err != nil {
		panic(err.Error())
	}
	return cmd
}

func startOldAgent() *exec.Cmd {
	cmd := exec.Command("./grafana-agent-flow", "run", "./rw.river", "--storage.path=~/bench/old-data", "--server.http.listen-addr=127.0.0.1:12346")
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	//cmd.Stdout = os.Stdout
	//cmd.Stderr = os.Stderr
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
	//println(fmt.Sprintf("index %d", index))
	if allow {
		//println(fmt.Sprintf("Body context is %d", r.ContentLength))
		return
	} else {
		println("returning 500")
		w.WriteHeader(500)
	}
}
