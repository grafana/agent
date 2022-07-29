package main

import (
	"os"

	"github.com/prometheus/common/version"
	"github.com/spf13/cobra"
)

func isFlowEnabled() bool {
	key, found := os.LookupEnv("EXPERIMENTAL_ENABLE_FLOW")
	if !found {
		return false
	}
	return key == "true" || key == "1"
}

func runFlow() {
	var cmd = &cobra.Command{
		Use:     "agent [global options] <subcommand>",
		Short:   "Grafana Agent Flow",
		Version: version.Print("agent"),

		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Usage()
		},
	}
	cmd.SetVersionTemplate("{{ .Version }}\n")

	cmd.AddCommand(runCommand())

	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
