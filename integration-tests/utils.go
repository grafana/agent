package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
)

const (
	agentBinaryPath  = "../../../build/grafana-agent-flow"
	dockerComposeCmd = "docker-compose"
	makeCmd          = "make"
)

type TestLog struct {
	TestDir    string
	AgentLog   string
	TestOutput string
}

var logChan chan TestLog

func buildAgent() {
	fmt.Println("Building agent...")
	cmd := exec.Command(makeCmd, "-C", "..", "agent-flow")
	if err := cmd.Run(); err != nil {
		panic(err)
	}
}

func setupEnvironment() {
	fmt.Println("Setting up environment with Docker Compose...")
	cmd := exec.Command(dockerComposeCmd, "up", "-d")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		panic(err)
	}
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

	cmd.Process.Kill()

	agentLog := agentLogBuffer.String()

	if errTest != nil {
		logChan <- TestLog{
			TestDir:    dirName,
			AgentLog:   agentLog,
			TestOutput: string(testOutput),
		}
	}

	os.RemoveAll(filepath.Join(testDir, "data-agent"))
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
}

func cleanUpEnvironment() {
	fmt.Println("Cleaning up Docker environment...")
	exec.Command(dockerComposeCmd, "down").Run()
}

func reportResults() {
	close(logChan)
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
