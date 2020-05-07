// Command agentctl provides utilies for interacting with Grafana Cloud Agent
package main

import (
	"github.com/spf13/cobra"
)

func main() {
	var cmd = &cobra.Command{
		Use:   "agentctl",
		Short: "Tools for interacting with the Grafana Cloud Agent",
	}

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
			_ = directory
			// TODO(rfratto): impl
		},
	}

	cmd.Flags().StringVarP(&agentAddr, "addr", "a", "", "address of the agent to connect to")
	must(cmd.MarkFlagRequired("addr"))
	return cmd
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
