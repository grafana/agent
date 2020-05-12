// Command agentctl provides utilities for interacting with Grafana Cloud Agent
package main

import (
	"os"

	// Adds version information
	_ "github.com/grafana/agent/pkg/build"
	"github.com/prometheus/common/version"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/grafana/agent/pkg/agentctl"
	"github.com/grafana/agent/pkg/client"
	"github.com/spf13/cobra"
)

func main() {
	var cmd = &cobra.Command{
		Use:     "agentctl",
		Short:   "Tools for interacting with the Grafana Cloud Agent",
		Version: version.Print("agentctl"),
	}
	cmd.SetVersionTemplate("{{ .Version }}\n")

	cmd.AddCommand(
		configSyncCmd(),
	)

	_ = cmd.Execute()
}

func configSyncCmd() *cobra.Command {
	var (
		agentAddr string
		dryRun    bool
	)

	cmd := &cobra.Command{
		Use:   "config-sync [directory]",
		Short: "Sync config files from a directory to an Agent's config management API",
		Long: `config-sync loads all files ending with .yml or .yaml from the specified
directory and uploads them the the config management API. The name of the config
uploaded will be the base name of the file (e.g., the name of the file without
its extension).

The directory is used as the source-of-truth for the entire set of configs that
should be present in the API. config-sync will delete all existing configs from the API 
that do not match any of the names of the configs that were uploaded from the 
source-of-truth directory.`,
		Args: cobra.ExactArgs(1),

		Run: func(_ *cobra.Command, args []string) {
			directory := args[0]
			cli := client.New(agentAddr)

			logger := log.NewLogfmtLogger(log.NewSyncWriter(os.Stdout))

			err := agentctl.ConfigSync(logger, cli.PrometheusClient, directory, dryRun)
			if err != nil {
				level.Error(logger).Log("msg", "failed to sync config", "err", err)
				os.Exit(1)
			}
		},
	}

	cmd.Flags().StringVarP(&agentAddr, "addr", "a", "http://localhost:12345", "address of the agent to connect to")
	cmd.Flags().BoolVarP(&dryRun, "dry-run", "d", false, "use the dry run option to validate config files without attempting to upload")
	must(cmd.MarkFlagRequired("addr"))
	return cmd
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
