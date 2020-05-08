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
	)

	cmd := &cobra.Command{
		Use:   "config-sync [directory]",
		Short: "Sync config files from a directory to an Agent's config management API",
		Args:  cobra.ExactArgs(1),

		Run: func(_ *cobra.Command, args []string) {
			directory := args[0]
			cli := client.New(agentAddr)

			logger := log.NewLogfmtLogger(log.NewSyncWriter(os.Stdout))

			err := agentctl.ConfigSync(logger, cli.PrometheusClient, directory)
			if err != nil {
				level.Error(logger).Log("msg", "failed to sync config", "err", err)
				return
			}
		},
	}

	cmd.Flags().StringVarP(&agentAddr, "addr", "a", "http://localhost:12345", "address of the agent to connect to")
	must(cmd.MarkFlagRequired("addr"))
	return cmd
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
