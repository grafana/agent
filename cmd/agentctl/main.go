// Command agentctl provides utilities for interacting with Grafana Cloud Agent
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	// Adds version information
	_ "github.com/grafana/agent/pkg/build"
	"github.com/olekukonko/tablewriter"
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
		walStatsCmd(),
		targetStatsCmd(),
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

func targetStatsCmd() *cobra.Command {
	var (
		jobLabel      string
		instanceLabel string
	)

	cmd := &cobra.Command{
		Use:   "target-stats [WAL directory]",
		Short: "Discover statitics on a specific target within the WAL.",
		Long: `target-stats computes statitics on a specific target within the WAL at 
greater detail than the general wal-stats. The statistics computed is the 
cardinality of all series within that target.

The cardinality for a series is defined as the total number of unique
combinations of label names and values that a given metric has. The result of
this operation can be used to define metric_relabel_rules and drop
high-cardinality series that you do not want to send.`,
		Args: cobra.ExactArgs(1),

		Run: func(_ *cobra.Command, args []string) {
			directory := args[0]
			if _, err := os.Stat(directory); os.IsNotExist(err) {
				fmt.Printf("%s does not exist\n", directory)
				os.Exit(1)
			} else if err != nil {
				fmt.Printf("error getting wal: %v\n", err)
				os.Exit(1)
			}

			// Check if ./wal is a subdirectory, use that instead.
			if _, err := os.Stat(filepath.Join(directory, "wal")); err == nil {
				directory = filepath.Join(directory, "wal")
			}

			cardinality, err := agentctl.FindCardinality(directory, jobLabel, instanceLabel)
			if err != nil {
				fmt.Printf("failed to get cardinality: %v\n", err)
				os.Exit(1)
			}

			sort.Slice(cardinality, func(i, j int) bool {
				return cardinality[i].Instances > cardinality[j].Instances
			})

			fmt.Printf("Metric cardinality:\n\n")

			for _, metric := range cardinality {
				fmt.Printf("%s: %d\n", metric.Metric, metric.Instances)
			}
		},
	}

	cmd.Flags().StringVarP(&jobLabel, "job", "j", "", "job label to search for")
	cmd.Flags().StringVarP(&instanceLabel, "instance", "i", "", "instance label to search for")
	must(cmd.MarkFlagRequired("job"))
	must(cmd.MarkFlagRequired("instance"))
	return cmd
}

func walStatsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "wal-stats [WAL directory]",
		Short: "Collect stats on the WAL",
		Long: `wal-stats reads a WAL directory and collects information on the series and 
samples within it.

The "Hash Collisions" value refers to the number of ref IDs a label's hash was
assigned to. A non-zero amount of collisions has no negative effect on the data 
sent to the Remote Write endpoint, but may have an impact on memory usage. Labels 
may collide with multiple ref IDs normally if a series flaps (i.e., gets marked for 
deletion but then comes back at some point).`,
		Args: cobra.ExactArgs(1),

		Run: func(_ *cobra.Command, args []string) {
			directory := args[0]
			if _, err := os.Stat(directory); os.IsNotExist(err) {
				fmt.Printf("%s does not exist\n", directory)
				os.Exit(1)
			} else if err != nil {
				fmt.Printf("error getting wal: %v\n", err)
				os.Exit(1)
			}

			// Check if ./wal is a subdirectory, use that instead.
			if _, err := os.Stat(filepath.Join(directory, "wal")); err == nil {
				directory = filepath.Join(directory, "wal")
			}

			stats, err := agentctl.CalculateStats(directory)
			if err != nil {
				fmt.Printf("failed to get WAL stats: %v\n", err)
				os.Exit(1)
			}

			fmt.Printf("Oldest Sample:      %s\n", stats.From)
			fmt.Printf("Newest Sample:      %s\n", stats.To)
			fmt.Printf("Total Series:       %d\n", stats.Series())
			fmt.Printf("Total Samples:      %d\n", stats.Samples())
			fmt.Printf("Hash Collisions:    %d\n", stats.HashCollisions)
			fmt.Printf("Invalid Refs:       %d\n", stats.InvalidRefs)
			fmt.Printf("Checkpoint Segment: %d\n", stats.CheckpointNumber)
			fmt.Printf("First Segment:      %d\n", stats.FirstSegment)
			fmt.Printf("Latest Segment:     %d\n", stats.LastSegment)

			fmt.Printf("\nPer-target stats:\n")

			table := tablewriter.NewWriter(os.Stdout)
			defer table.Render()

			table.SetHeader([]string{"Job", "Instance", "Series", "Samples"})

			sort.Sort(agentctl.BySeriesCount(stats.Targets))

			for _, t := range stats.Targets {
				seriesStr := fmt.Sprintf("%d", t.Series)
				samplesStr := fmt.Sprintf("%d", t.Samples)
				table.Append([]string{t.Job, t.Instance, seriesStr, samplesStr})
			}
		},
	}
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
