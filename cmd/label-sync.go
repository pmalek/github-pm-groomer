package cmd

import (
	"github.com/lahabana/github-pm-groomer/internal/labelsync"
	"github.com/spf13/cobra"
	"time"
)

var (
	labelSyncCmd = &cobra.Command{
		Use: "label-sync",
		Short: "do things to labels",
		Long: "Inspired by https://github.com/kubernetes/test-infra/tree/master/label_sync but with less options",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := labelSyncOpts.Validate(); err != nil {
				return err
			}
			return labelsync.Run(cmd.Context(), ghClient, labelSyncOpts, time.Now())
		},
	}
	labelSyncOpts labelsync.Opts
)

func init() {
	labelSyncCmd.Flags().StringVarP(&labelSyncOpts.Repo, "repo","r", "", "The repo to use")
	labelSyncCmd.Flags().StringVarP(&labelSyncOpts.FilePath, "path","p", "", "The path or url to the labels to sync")
	rootCmd.AddCommand(labelSyncCmd)
}
