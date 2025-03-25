package labels

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/pmalek/github-pm-groomer/internal/github/api"
	"github.com/pmalek/github-pm-groomer/internal/issues"
)

type Action string

const (
	AddAction     = "add"
	RemoveAction  = "remove"
	ReplaceAction = "replace"
)

var AllOptions = []string{AddAction, RemoveAction, ReplaceAction}

type Opts struct {
	Action        string
	Label         string
	NewLabel      string
	IssueSelector issues.Selector
}

func (l Opts) Validate() error {
	if l.Action != AddAction && l.Action != RemoveAction && l.Action != ReplaceAction {
		return fmt.Errorf("invalid action type '%s' valid options: %s", l.Action, strings.Join(AllOptions, ","))
	}
	if l.Action == ReplaceAction && l.NewLabel == "" {
		return errors.New("when using replace you must set a new label")
	}
	if l.Label == "" {
		return errors.New("must set a label to act on")
	}
	if err := l.IssueSelector.Validate(); err != nil {
		return err
	}
	return nil
}

func Run(ctx context.Context, client api.Client, opts Opts, now time.Time) error {
	if opts.Action == RemoveAction || opts.Action == ReplaceAction {
		// We're removing so let's only select issues with the label in the first place
		opts.IssueSelector.Labels = strings.Join(append(strings.Split(opts.IssueSelector.Labels, ","), opts.Label), ",")
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
			newLabels = issue.RemoveLabel(opts.Label)
		case AddAction:
			newLabels = issue.AddLabel(opts.Label)
		case ReplaceAction:
			newLabels = issue.ReplaceLabel(opts.Label, opts.NewLabel)

		}
		if err := client.UpdateLabels(ctx, opts.IssueSelector.Repo, *issue.Number, newLabels); err != nil {
			return err
		}
	}
}
