package lifecycle

import (
	"context"
	"fmt"
	"github.com/lahabana/github-pm-groomer/internal/github/api"
	"github.com/lahabana/github-pm-groomer/internal/issues"
	"time"
)

type Opts struct {
	StaleDuration time.Duration
	StaleLabel string
	RottenDuration time.Duration
	RottenLabel string
	IssueSelector issues.Selector
}

func (o Opts) Validate() error {
	if err := o.IssueSelector.Validate(); err != nil {
		return err
	}
	return nil
}


func Run(ctx context.Context, client api.Client, opts Opts, now time.Time) error {
	iterator := opts.IssueSelector.Iterator(ctx, client, now)
	for {
		n, err := iterator.Next()
		if err != nil {
			return err
		}
		if n == nil {
			return nil
		}
		if n.HasLabel(opts.RottenLabel) {
			if n.UpdatedAt != nil && n.UpdatedAt.Add(opts.RottenDuration).Before(now) {
				err = client.Comment(ctx, opts.IssueSelector.Repo, *n.Number, fmt.Sprintf("Issue rotten for %s. Closing it!", opts.RottenDuration.String()))
				if err != nil {
					return err
				}
				err = client.UpdateIssueState(ctx, opts.IssueSelector.Repo, *n.Number, "closed")
				if err != nil {
					return err
				}
			}

		} else if !n.HasLabel(opts.StaleLabel) {
			if n.UpdatedAt != nil && n.UpdatedAt.Add(opts.StaleDuration).Before(now) {
				err = client.Comment(ctx, opts.IssueSelector.Repo, *n.Number,
					fmt.Sprintf("Issue marked as staled after being inactive for %s, It will be reviewed at the next triage meeting", opts.StaleDuration.String()))
				if err != nil {
					return err
				}
				labels := n.AddLabel(opts.StaleLabel)
				err = client.UpdateLabels(ctx, opts.IssueSelector.Repo, *n.Number, labels)
				if err != nil {
					return err
				}
			}
		}
	}
}
