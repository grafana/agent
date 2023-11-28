package main

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sync"

	"github.com/testcontainers/testcontainers-go"
)

const (
	agentImage                     = "agent-integration-tests"
	prometheusMetricGeneratorImage = "prometheus-metrics-generator"
	otelMetricsGeneratorImage      = "otel-metrics-generator"
	mimirImage                     = "grafana/mimir:2.10.4"
	lokiImage                      = "grafana/loki:2.9.2"
	networkName                    = "integration-tests"
)

var (
	runningContainers []testcontainers.Container
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

func setupContainers(ctx context.Context) {
	executeCommand("make", []string{"-C", "..", "AGENT_IMAGE=" + agentImage, "agent-image"}, "Building agent")
	buildDockerImage("./configs/prom-gen/Dockerfile", "../", prometheusMetricGeneratorImage)
	buildDockerImage("./configs/otel-gen/Dockerfile", "../", otelMetricsGeneratorImage)
	runningContainers = append(runningContainers, startContainer(ctx, createMimirContainer(mimirImage)))
	runningContainers = append(runningContainers, startContainer(ctx, createLokiContainer(lokiImage)))
	runningContainers = append(runningContainers, startContainer(ctx, createPrometheusMetricGeneratorContainer(prometheusMetricGeneratorImage)))
	runningContainers = append(runningContainers, startContainer(ctx, createOTLPMetricsGeneratorContainer(otelMetricsGeneratorImage)))
}

func buildDockerImage(dockerfilePath, contextPath, imageName string) {
	cmd := exec.Command("docker", "build", "-f", dockerfilePath, "-t", imageName, contextPath)
	err := cmd.Run()
	if err != nil {
		log.Fatalf("Failed to build Docker image: %s", err)
	}
}

func runSingleTest(ctx context.Context, testDir string) {
	info, err := os.Stat(testDir)
	if err != nil {
		panic(err)
	}
	if !info.IsDir() {
		return
	}

	dirName := filepath.Base(testDir)

	agentContainer := startContainer(ctx, createAgentContainer(agentImage, dirName))

	err = agentContainer.StartLogProducer(ctx)
	if err != nil {
		panic(err)
	}

	var agentLogBuffer bytes.Buffer
	logConsumer := &logConsumer{buf: &agentLogBuffer}
	agentContainer.FollowOutput(logConsumer)

	testCmd := exec.Command("go", "test")
	testCmd.Dir = testDir
	testOutput, errTest := testCmd.CombinedOutput()

	agentLog := agentLogBuffer.String()

	if errTest != nil {
		logChan <- TestLog{
			TestDir:    dirName,
			AgentLog:   agentLog,
			TestOutput: string(testOutput),
		}
	}

	err = agentContainer.StopLogProducer()
	if err != nil {
		panic(err)
	}

	err = agentContainer.Terminate(ctx)
	if err != nil {
		panic(err)
	}

	err = os.RemoveAll(filepath.Join(testDir, "data-agent"))
	if err != nil {
		panic(err)
	}
}

func runAllTests(ctx context.Context) {
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
			runSingleTest(ctx, td)
		}(testDir)
	}
	wg.Wait()
	close(logChan)
}

func cleanUpEnvironment(ctx context.Context) {
	fmt.Println("Terminate test containers...")
	var errors []error

	for _, container := range runningContainers {
		err := container.Terminate(ctx)
		if err != nil {
			errors = append(errors, err)
		}
	}

	if len(errors) > 0 {
		panic(fmt.Sprintf("Errors occurred while terminating containers: %v", errors))
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

func setupNetwork(ctx context.Context) testcontainers.Network {
	networkRequest := testcontainers.GenericNetworkRequest{
		NetworkRequest: testcontainers.NetworkRequest{
			Name:           networkName,
			CheckDuplicate: true,
		},
	}
	network, err := testcontainers.GenericNetwork(ctx, networkRequest)
	if err != nil {
		panic(err)
	}
	return network
}

func cleanUpImages() {
	fmt.Println("Cleaning up Docker images...")

	images := []string{agentImage, mimirImage, lokiImage, prometheusMetricGeneratorImage, otelMetricsGeneratorImage}

	args := append([]string{"rmi"}, images...)
	cmd := exec.Command("docker", args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		fmt.Printf("Error removing images: %s\n", stderr.String())
	}
}

func cleanUpNetwork(ctx context.Context, network testcontainers.Network) {
	err := network.Remove(ctx)
	if err != nil {
		fmt.Printf("Failed to remove network: %s\n", err)
	}
}

type logConsumer struct {
	buf *bytes.Buffer
}

func (c *logConsumer) Accept(l testcontainers.Log) {
	c.buf.Write(l.Content)
}
