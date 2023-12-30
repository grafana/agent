package main

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
	"time"
)

func startLogsRun(run time.Duration) {
	allow = true
	_ = os.MkdirAll("./data/", 0777)
	_ = os.RemoveAll("./data/")
	_ = os.Setenv("NAME", "logs")
	gen := startLogsGenAgent()
	old := startLogsAgent()
	fmt.Println("starting logs agent")
	defer func() {
		_ = old.Process.Kill()
		_ = old.Process.Release()
		_ = old.Wait()
		_ = syscall.Kill(-old.Process.Pid, syscall.SIGKILL)
		_ = gen.Process.Kill()
		_ = gen.Process.Release()
		_ = gen.Wait()
		_ = syscall.Kill(-gen.Process.Pid, syscall.SIGKILL)
		_ = os.RemoveAll("./data/")
	}()

	time.Sleep(run)
}

func startLogsAgent() *exec.Cmd {
	cmd := exec.Command("./grafana-agent-flow", "run", "./logs.river", "--storage.path=./data/logs", "--server.http.listen-addr=127.0.0.1:12346")
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Start()
	if err != nil {
		panic(err.Error())
	}
	return cmd
}

func startLogsGenAgent() *exec.Cmd {
	cmd := exec.Command("./grafana-agent-flow", "run", "./logsgen.river", "--storage.path=./data/logs-gen", "--server.http.listen-addr=127.0.0.1:12349")
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Start()
	if err != nil {
		panic(err.Error())
	}
	return cmd
}
