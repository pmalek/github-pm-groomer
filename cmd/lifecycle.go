package cmd

import (
	"github.com/lahabana/github-pm-groomer/internal/lifecycle"
	"github.com/spf13/cobra"
	"time"
)

var (
	lifecycleCmd = &cobra.Command{
		Use: "lifecycle",
		Short: "Mark issue as stale if they have been used for some time or rotten or close them.",
		Long: "Mark issue as stale if they have been used for some time or rotten or close them.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := lifeCycleOpts.Validate(); err != nil {
				return err
			}
			return lifecycle.Run(cmd.Context(), ghClient, lifeCycleOpts, time.Now())
		},
	}
	lifeCycleOpts lifecycle.Opts
)

func init() {
	lifecycleCmd.Flags().DurationVar(&lifeCycleOpts.StaleDuration, "stale", time.Duration(0), "How long to wait before marking an issue as staled")
	lifecycleCmd.Flags().StringVar(&lifeCycleOpts.StaleLabel, "stale-label", "triage/stale", "The name of the label for staled issues")
	lifecycleCmd.Flags().DurationVar(&lifeCycleOpts.RottenDuration, "rotten", time.Duration(0), "How long to wait before closing issues marked as rotten")
	lifecycleCmd.Flags().StringVar(&lifeCycleOpts.RottenLabel, "rotten-label", "triage/rotten", "The name of the label for rotten issues")
	decorateWithIssueSelector(lifecycleCmd, &lifeCycleOpts.IssueSelector)

	rootCmd.AddCommand(lifecycleCmd)
}
