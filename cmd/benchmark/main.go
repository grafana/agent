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
	os.RemoveAll("./data/normal-data")
	os.RemoveAll("./data/test-data")

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
	defer os.RemoveAll("./data/test-data")

	old := startNormalAgent()
	fmt.Println("starting normal agent")
	defer old.Process.Kill()
	defer old.Process.Release()
	defer old.Wait()
	defer syscall.Kill(-old.Process.Pid, syscall.SIGKILL)
	defer os.RemoveAll("./data/normal-data")

	time.Sleep(run)
}

func buildAgent() {
	cmd := exec.Command("go", "build", "../grafana-agent-flow")
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		panic(err.Error())
	}
}

func startNormalAgent() *exec.Cmd {
	cmd := exec.Command("./grafana-agent-flow", "run", "./normal.river", "--storage.path=./data/normal-data", "--server.http.listen-addr=127.0.0.1:12346")
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
	cmd := exec.Command("./grafana-agent-flow", "run", "./test.river", "--storage.path=./data/test-data", "--server.http.listen-addr=127.0.0.1:9001")
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
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
	if allow {
		return
	} else {
		println("returning 500")
		w.WriteHeader(500)
	}
}
