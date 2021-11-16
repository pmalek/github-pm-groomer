package issues

import (
	"context"
	"github.com/lahabana/github-pm-groomer/internal/github/api"
	"github.com/lahabana/github-pm-groomer/internal/utils"
	"strconv"
	"strings"
	"time"
)

type Selector struct {
	Repo      string
	Labels    string
	State     string
	Since     time.Duration
	Limit     int
	IssueList string
}

func (l Selector) Validate() error {
	if _, _, err := utils.OrgRepo(l.Repo); err != nil {
		return err
	}
	return nil
}

func (l Selector) listOpts(now time.Time) api.IssueListOptions {
	r := api.IssueListOptions{
		State:  l.State,
		Labels: l.Labels,
	}
	if l.Since != 0 {
		r.Since = now.Add(-l.Since)
	}
	return r
}

func (l Selector) Iterator(ctx context.Context, client api.Client, now time.Time) SelectorIterator {
	var err error
	i := 0
	if l.IssueList != "" {
		issues := strings.Split(l.IssueList, ",")
		return SelectorIteratorFunc(func() (*api.Issue, error) {
			if err != nil {
				return nil, err
			}
			if i == len(issues) {
				return nil, nil
			}
			var issue *api.Issue
			var num int
			num, err = strconv.Atoi(issues[i])
			if err != nil {
				return nil, err
			}
			issue, err = client.GetIssue(ctx, l.Repo, num)
			i+=1
			return issue, err


		})
	} else {
		opts := l.listOpts(now)
		page := 0
		var currentItems []*api.Issue
		left := l.Limit
		if left == 0 {
			left = -1
		}
		return SelectorIteratorFunc(func() (*api.Issue, error) {
			if err != nil {
				return nil, err
			}
			if i == len(currentItems) {
				i = 0
				currentItems, err = client.GetIssues(ctx, l.Repo, opts, page)
				if err != nil {
					return nil, err
				}
				if len(currentItems) == 0 {
					return nil, nil
				}
				page += 1
			}
			itm := currentItems[i]
			i += 1
			left -= 1
			if left == 0 {
				// We're at the end
				return nil, nil
			}
			return itm, nil
		})
	}

}

type SelectorIterator interface {
	Next() (*api.Issue, error)
}

type SelectorIteratorFunc func() (*api.Issue, error)

func (f SelectorIteratorFunc) Next() (*api.Issue, error) {
	return f()
}
