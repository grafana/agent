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

func runTests(logChan chan TestLog) {
	testDirs, err := filepath.Glob("./tests/*")
	if err != nil {
		panic(err)
	}
	var wg sync.WaitGroup

	for _, testDir := range testDirs {
		info, _ := os.Stat(testDir)
		if !info.IsDir() {
			continue
		}
		wg.Add(1)
		go func(testDir string) {
			defer wg.Done()

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
		}(testDir)
	}
	wg.Wait()
}

func main() {
	fmt.Println("Build agent...")
	cmd := exec.Command(makeCmd, "-C", "..", "agent-flow")
	if err := cmd.Run(); err != nil {
		panic(err)
	}

	fmt.Println("Docker compose up...")
	cmd = exec.Command(dockerComposeCmd, "up", "-d")
	if err := cmd.Run(); err != nil {
		panic(err)
	}

	testDirs, err := filepath.Glob("./tests/*")
	if err != nil {
		panic(err)
	}
	logChan := make(chan TestLog, len(testDirs))
	fmt.Println("Start the tests...")
	runTests(logChan)

	fmt.Println("Docker compose down...")
	exec.Command(dockerComposeCmd, "down").Run()

	close(logChan)
	testsFailed := 0
	for log := range logChan {
		fmt.Printf("Failure detected in %s:\n", log.TestDir)
		fmt.Println("Test output:", log.TestOutput)
		fmt.Println("Agent logs:", log.AgentLog)
		testsFailed += 1
	}

	if testsFailed > 0 {
		fmt.Printf("%d tests failed!\n", testsFailed)
		os.Exit(1)
	} else {
		fmt.Println("All integration tests passed!")
		os.Exit(0)
	}
}
