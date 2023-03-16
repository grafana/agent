package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/grafana/agent/pkg/autodiscovery/runner"
	"github.com/grafana/agent/pkg/river/diag"
	"github.com/spf13/cobra"
)

func autodiscoverCommand() *cobra.Command {
	f := &flowAd{}

	cmd := &cobra.Command{
		Use:   "autodiscover",
		Short: "Run autodiscovery",
		Long: `The autodiscover subcommand detects things you can monitor in the host machine and generates
a River configuration file to visualize them in Grafana Cloud.`,
		Args:         cobra.RangeArgs(0, 0),
		SilenceUsage: true,

		RunE: func(_ *cobra.Command, args []string) error {
			var err error

			f.Run("-")

			var diags diag.Diagnostics
			if errors.As(err, &diags) {
				for _, diag := range diags {
					fmt.Fprintln(os.Stderr, diag)
				}
				return fmt.Errorf("encountered errors during autodiscovery")
			}

			return err
		},
	}

	return cmd
}

type flowAd struct {
	write bool
}

func (fa *flowAd) Run(configFile string) error {
	ad := runner.Autodiscovery{
		Disabled: map[runner.AutodiscT]struct{}{
			"redis": {},
		},
	}
	detected := ad.Do(os.Stdout)
	_ = detected

	var integrations []string
	for _, d := range detected {
		integrations = append(integrations, string(d))
	}
	fmt.Fprintf(os.Stderr, "Installing Grafana Cloud integrations:\n")
	runner.InstallIntegrations(os.Getenv("GCLOUD_ADMIN_API_KEY"), integrations...)

	return nil
}
