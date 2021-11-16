package cmd

import (
	"fmt"
	"github.com/lahabana/github-pm-groomer/internal/labels"
	"github.com/spf13/cobra"
	"strings"
	"time"
)

var (
	labelsCmd = &cobra.Command{
		Use: "labels",
		Short: "do things to labels",
		Long: "Add or Remove labels, the will also apply to PRs",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := opts.Validate(); err != nil {
				return err
			}
			return labels.Run(cmd.Context(), ghClient, opts, time.Now())
		},
	}
	opts labels.LabelOpts
)

func init() {
	labelsCmd.Flags().StringVarP(&opts.Action, "action","a", "add", fmt.Sprintf("what to do on the issues (%s)", strings.Join(labels.AllOptions, ",")))
	labelsCmd.Flags().StringVarP(&opts.Label, "label","l", "", "The label to add/remove")
	decorateWithIssueSelector(labelsCmd, &opts.IssueSelector)

	rootCmd.AddCommand(labelsCmd)
}
