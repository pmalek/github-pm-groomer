package api

import (
	"context"
	"github.com/google/go-github/v40/github"
	"github.com/lahabana/github-pm-groomer/internal/utils"
	"golang.org/x/oauth2"
	"strings"
	"time"
)

const IssuesPerPage = 100

type Client interface {
	GetIssue(ctx context.Context, orgRepo string, issue int) (*Issue, error)
	GetIssues(ctx context.Context, orgRepo string, options IssueListOptions, page int) ([]*Issue, error)
	UpdateLabels(ctx context.Context, orgRepo string, issue int, labels []string) error
	UpdateIssueState(ctx context.Context, orgRepo string, issue int, state string) error
	Ping(ctx context.Context) error
}

type Issue github.Issue

type githubClient struct {
	client *github.Client
}

func New(token string) Client {
	client := github.NewClient(nil)
	if token != "" {
		ctx := context.Background()
		ts := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: token},
		)
		tc := oauth2.NewClient(ctx, ts)
		client = github.NewClient(tc)
	}

	return &githubClient{
		client: client,
	}
}

func (gc *githubClient) Ping(ctx context.Context) error {
	_, _, err := gc.client.Licenses.Get(ctx, "MIT")
	return err
}

type IssueListOptions struct {
	Labels string
	State string
	Since time.Time
}

func (gc *githubClient) UpdateLabels(ctx context.Context, orgRepo string, issue int, labels []string) error {
	org, repo := utils.MustOrgRepo(orgRepo)
	if len(labels) == 0 {
		_, err := gc.client.Issues.RemoveLabelsForIssue(ctx, org, repo, issue)
		return err
	}
	_, _, err := gc.client.Issues.ReplaceLabelsForIssue(ctx, org, repo, issue, labels)
	return err
}

func (gc *githubClient) UpdateIssueState(ctx context.Context, orgRepo string, issue int, state string) error {
	org, repo := utils.MustOrgRepo(orgRepo)
	_, _, err := gc.client.Issues.Edit(ctx, org, repo, issue, &github.IssueRequest{State: &state})
	return err

}

func (gc *githubClient) GetIssues(ctx context.Context, orgRepo string, options IssueListOptions, page int) ([]*Issue, error) {
	org, repo := utils.MustOrgRepo(orgRepo)
	issues, _, err := gc.client.Issues.ListByRepo(ctx, org, repo, &github.IssueListByRepoOptions{
		Labels: strings.Split(options.Labels, ","),
		Since: options.Since,
		State: options.State,
		ListOptions: github.ListOptions{
			Page:    page,
			PerPage: IssuesPerPage,
		},
	})
	res := make([]*Issue, len(issues))
	for i := range issues {
		res[i] = (*Issue)(issues[i])
	}
	return res, err
}

func (gc *githubClient) GetIssue(ctx context.Context, orgRepo string, issueNumber int) (*Issue, error) {
	org, repo := utils.MustOrgRepo(orgRepo)
	issue, _, err := gc.client.Issues.Get(ctx, org, repo, issueNumber)
	if err != nil {
		return nil, err
	}
	return (*Issue)(issue), nil
}
