package cmd

import (
	"fmt"
	"github.com/lahabana/github-pm-groomer/internal/issues"
	"github.com/spf13/cobra"
	"time"
)

func decorateWithIssueSelector(cmd *cobra.Command, selector *issues.Selector) {
	cmd.Flags().StringVar(&selector.Repo, "repo", "", "The <org>/<repo> to query")
	cmd.Flags().StringVar(&selector.Labels, "labels", "", "A list of comma separated labels like: https://docs.github.com/en/rest/reference/issues#list-repository-issues")
	cmd.Flags().StringVar(&selector.State, "state", "open", fmt.Sprintf("The state of the issue: open,closed,all"))
	cmd.Flags().DurationVar(&selector.Since, "since", time.Duration(0), "Only apply to issues touched since")
	cmd.Flags().IntVar(&selector.Limit, "limit", -1, "The max number of issues to return (-1 for all)")
	cmd.Flags().StringVar(&selector.IssueList, "issues", "", "A comma separated list of issues to modify")
}
