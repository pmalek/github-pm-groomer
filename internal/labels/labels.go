package labels

import (
	"context"
	"errors"
	"fmt"
	"github.com/lahabana/github-pm-groomer/internal/github/api"
	"github.com/lahabana/github-pm-groomer/internal/issues"
	"strings"
	"time"
)

type Action string

const (
	AddAction    = "add"
	RemoveAction = "remove"
)

var AllOptions = []string{AddAction, RemoveAction}

type LabelOpts struct {
	Action        string
	Label         string
	IssueSelector issues.Selector
}

func (l LabelOpts) Validate() error {
	if l.Action != AddAction && l.Action != RemoveAction {
		return fmt.Errorf("invalid action type '%s' valid options: %s", l.Action, strings.Join(AllOptions, ","))
	}
	if l.Label == "" {
		return errors.New("must set a label to act on")
	}
	if err := l.IssueSelector.Validate(); err != nil {
		return err
	}
	return nil
}

func Run(ctx context.Context, client api.Client, opts LabelOpts, now time.Time) error {
	if opts.Action == RemoveAction {
		// We're removing so let's only select issues with the label in the first place
		strings.Join(append(strings.Split(opts.IssueSelector.Labels, ","), opts.Label), ",")
	}
	iterator := opts.IssueSelector.Iterator(ctx, client, now)
	for {
		issue, err := iterator.Next()
		if err != nil {
			return err
		}
		if issue == nil {
			return nil
		}
		var newLabels []string
		switch opts.Action {
		case RemoveAction:
			for _, v := range issue.Labels {
				if *v.Name != opts.Label {
					newLabels = append(newLabels, *v.Name)
				}
			}
		case AddAction:
			seen := false
			for _, v := range issue.Labels {
				if *v.Name == opts.Label {
					seen = true
				}
				newLabels = append(newLabels, *v.Name)
			}
			if !seen {
				newLabels = append(newLabels, opts.Label)
			}
		}
		if err := client.UpdateLabels(ctx, opts.IssueSelector.Repo, *issue.Number, newLabels); err != nil {
			return err
		}
	}
}
