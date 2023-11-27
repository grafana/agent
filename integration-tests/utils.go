package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
)

const (
	agentBinaryPath = "../../../build/grafana-agent-flow"
)

type TestLog struct {
	TestDir    string
	AgentLog   string
	TestOutput string
}

var logChan chan TestLog

func executeCommand(command string, args []string, taskDescription string) {
	fmt.Printf("%s...\n", taskDescription)
	cmd := exec.Command(command, args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		log.Fatalf("Error: %s\n", stderr.String())
	}
}

func buildAgent() {
	executeCommand("make", []string{"-C", "..", "agent-flow"}, "Building agent")
}

func setupEnvironment() {
	executeCommand("docker-compose", []string{"up", "-d"}, "Setting up environment with Docker Compose")
}

func runSingleTest(testDir string) {
	info, err := os.Stat(testDir)
	if err != nil {
		panic(err)
	}
	if !info.IsDir() {
		return
	}

	dirName := filepath.Base(testDir)

	var agentLogBuffer bytes.Buffer
	cmd := exec.Command(agentBinaryPath, "run", "config.river")
	cmd.Dir = testDir
	cmd.Stdout = &agentLogBuffer
	cmd.Stderr = &agentLogBuffer

	if err := cmd.Start(); err != nil {
		logChan <- TestLog{
			TestDir:  dirName,
			AgentLog: fmt.Sprintf("Failed to start agent: %v", err),
		}
		return
	}

	testCmd := exec.Command("go", "test")
	testCmd.Dir = testDir
	testOutput, errTest := testCmd.CombinedOutput()

	err = cmd.Process.Kill()
	if err != nil {
		panic(err)
	}

	agentLog := agentLogBuffer.String()

	if errTest != nil {
		logChan <- TestLog{
			TestDir:    dirName,
			AgentLog:   agentLog,
			TestOutput: string(testOutput),
		}
	}

	err = os.RemoveAll(filepath.Join(testDir, "data-agent"))
	if err != nil {
		panic(err)
	}
}

func runAllTests() {
	testDirs, err := filepath.Glob("./tests/*")
	if err != nil {
		panic(err)
	}
	var wg sync.WaitGroup

	for _, testDir := range testDirs {
		fmt.Println("Running", testDir)
		wg.Add(1)
		go func(td string) {
			defer wg.Done()
			runSingleTest(td)
		}(testDir)
	}
	wg.Wait()
	close(logChan)
}

func cleanUpEnvironment() {
	fmt.Println("Cleaning up Docker environment...")
	err := exec.Command("docker-compose", "down", "--volumes", "--rmi", "all").Run()
	if err != nil {
		panic(err)
	}
}

func reportResults() {
	testsFailed := 0
	for log := range logChan {
		fmt.Printf("Failure detected in %s:\n", log.TestDir)
		fmt.Println("Test output:", log.TestOutput)
		fmt.Println("Agent logs:", log.AgentLog)
		testsFailed++
	}

	if testsFailed > 0 {
		fmt.Printf("%d tests failed!\n", testsFailed)
		os.Exit(1)
	} else {
		fmt.Println("All integration tests passed!")
	}
}
