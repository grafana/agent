package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var specificTest string

func main() {
	rootCmd := &cobra.Command{
		Use:   "integration-tests",
		Short: "Run integration tests",
		Run:   runIntegrationTests,
	}

	rootCmd.PersistentFlags().StringVar(&specificTest, "test", "", "Specific test directory to run")

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func runIntegrationTests(cmd *cobra.Command, args []string) {
	ctx := context.Background()
	network := setupNetwork(ctx)

	defer reportResults()
	defer cleanUpImages()
	defer cleanUpNetwork(ctx, network)
	defer cleanUpEnvironment(ctx)

	setupContainers(ctx)

	if specificTest != "" {
		fmt.Println("Running", specificTest)
		if !filepath.IsAbs(specificTest) && !strings.HasPrefix(specificTest, "./tests/") {
			specificTest = "./tests/" + specificTest
		}
		logChan = make(chan TestLog, 1)
		runSingleTest(ctx, specificTest)
	} else {
		testDirs, err := filepath.Glob("./tests/*")
		if err != nil {
			panic(err)
		}
		logChan = make(chan TestLog, len(testDirs))
		runAllTests(ctx)
	}
}
