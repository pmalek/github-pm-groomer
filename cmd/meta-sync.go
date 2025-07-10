package cmd

import (
	"runtime"
	"time"

	"github.com/pmalek/github-pm-groomer/internal/metasync"
	"github.com/spf13/cobra"
)

var (
	metaSyncCmd = &cobra.Command{
		Use:     "meta-sync",
		Aliases: []string{"label-sync"},
		Short:   "do things to repo metadata",
		Long:    "Inspired by https://github.com/kubernetes/test-infra/tree/master/label_sync but with less options",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := metaSyncOpts.Validate(); err != nil {
				return err
			}
			return metasync.Run(cmd.Context(), ghClient, metaSyncOpts, time.Now())
		},
	}
	metaSyncOpts metasync.Opts
)

func init() {
	metaSyncCmd.Flags().StringVarP(&metaSyncOpts.FilePath, "path", "p", "", "The path or url to the labels to sync")
	metaSyncCmd.Flags().IntVarP(&metaSyncOpts.Concurrency, "concurrency", "c", runtime.NumCPU(), "The number of concurrent goroutines to use for syncing metadata")
	rootCmd.AddCommand(metaSyncCmd)
}
