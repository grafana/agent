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

	name := os.Args[1]
	allowWal := os.Args[2]
	duration := os.Args[3]
	discovery := os.Args[4]
	allowWalBool, _ := strconv.ParseBool(allowWal)
	parsedDuration, _ := time.ParseDuration(duration)
	fmt.Println(name, allowWalBool, parsedDuration, discovery)
	startRun(name, allowWalBool, parsedDuration, discovery)

}

func startRun(name string, allowWAL bool, run time.Duration, discovery string) {
	os.RemoveAll("~/bench/queue-data")
	os.RemoveAll("~/bench/old-data")
	os.RemoveAll("~/bench/test-data")
	os.RemoveAll("~/bench/qcache-data")

	allow = allowWAL
	_ = os.Setenv("NAME", name)
	_ = os.Setenv("ALLOW_WAL", strconv.FormatBool(allowWAL))
	_ = os.Setenv("DISCOVERY", discovery)

	metric := startMetricsAgent()
	fmt.Println("starting metric agent")
	defer metric.Process.Kill()
	defer metric.Process.Release()
	defer metric.Wait()
	defer syscall.Kill(-metric.Process.Pid, syscall.SIGKILL)
	defer os.RemoveAll("~/bench/metric-data")

	old := startOldAgent()
	fmt.Println("starting old agent")
	defer old.Process.Kill()
	defer old.Process.Release()
	defer old.Wait()
	defer syscall.Kill(-old.Process.Pid, syscall.SIGKILL)
	defer os.RemoveAll("~/bench/old-data")

	queue := startQueueAgent()
	fmt.Println("start queue agent")
	defer queue.Process.Kill()
	defer queue.Process.Release()
	defer queue.Wait()
	defer syscall.Kill(-queue.Process.Pid, syscall.SIGKILL)
	defer os.RemoveAll("~/bench/queue-data")

	cache := startQCacheAgent()
	fmt.Println("start cache agent")
	defer cache.Process.Kill()
	defer cache.Process.Release()
	defer cache.Wait()
	defer syscall.Kill(-cache.Process.Pid, syscall.SIGKILL)
	defer os.RemoveAll("~/bench/qcache-data")

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

func startQCacheAgent() *exec.Cmd {
	cmd := exec.Command("./grafana-agent-flow", "run", "./qcache.river", "--storage.path=~/bench/qcache-data", "--server.http.listen-addr=127.0.0.1:12348")
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

func startMetricsAgent() *exec.Cmd {
	cmd := exec.Command("./grafana-agent-flow", "run", "./test.river", "--storage.path=~/bench/test-data", "--server.http.listen-addr=127.0.0.1:9001")
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
	//println(fmt.Sprintf("index %d", index))
	if allow {
		//println(fmt.Sprintf("Body context is %d", r.ContentLength))
		return
	} else {
		println("returning 500")
		w.WriteHeader(500)
	}
}
