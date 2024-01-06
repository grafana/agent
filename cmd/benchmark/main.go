package main

import (
	"net/http"
	"os"
	"os/exec"
	"syscall"

	"github.com/gorilla/mux"
)

// main handles creating the benchmark.
func main() {
	flags()
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

func cleanupPid(pid *exec.Cmd, dir string) {
	_ = pid.Process.Kill()
	_ = pid.Process.Release()
	_ = pid.Wait()
	_ = syscall.Kill(-pid.Process.Pid, syscall.SIGKILL)
	_ = os.RemoveAll(dir)
}

var networkdown = true

func httpServer() {
	r := mux.NewRouter()
	r.HandleFunc("/post", func(w http.ResponseWriter, r *http.Request) {
		handlePost(w, r)
	})
	r.HandleFunc("/allow", func(w http.ResponseWriter, r *http.Request) {
		println("allowing")
		networkdown = true
	})
	r.HandleFunc("/block", func(w http.ResponseWriter, r *http.Request) {
		println("blocking")
		networkdown = false
	})
	http.Handle("/", r)
	println("Starting server")
	err := http.ListenAndServe(":8888", nil)
	if err != nil {
		println(err)
	}
}

func handlePost(w http.ResponseWriter, _ *http.Request) {
	if networkdown {
		println("returning 500")
		w.WriteHeader(500)
		return
	} else {
		return
	}
}
