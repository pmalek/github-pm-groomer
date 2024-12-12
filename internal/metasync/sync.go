package metasync

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/lahabana/github-pm-groomer/internal/github/api"
	"github.com/lahabana/github-pm-groomer/internal/utils"
	"gopkg.in/yaml.v3"
)

type Opts struct {
	Repo     string
	FilePath string
}

type Conf struct {
	Default RepoConf `yaml:"default"`
}
type RepoConf struct {
	Labels     []LabelDef     `yaml:"labels"`
	Milestones []MilestoneDef `yaml:"milestones"`
}

type MilestoneDef struct {
	Title       string    `yaml:"title"`
	Closed      bool      `yaml:"closed"`
	DueDate     time.Time `yaml:"dueDate"`
	Description string    `yaml:"description"`
	Delete      bool      `yaml:"delete"`
}

type LabelDef struct {
	Color       string `yaml:"color"`
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	Delete      bool   `yaml:"delete"`
}

func (o Opts) Validate() error {
	if _, _, err := utils.OrgRepo(o.Repo); err != nil {
		return err
	}

	return nil
}

func Run(ctx context.Context, client api.Client, opts Opts, now time.Time) error {
	labelConf, err := parseConf(opts.FilePath)
	if err != nil {
		return err
	}

	err = syncLabels(ctx, client, opts.Repo, labelConf)
	if err != nil {
		return err
	}
	return syncMilestones(ctx, client, opts.Repo, labelConf)
}

func syncLabels(ctx context.Context, client api.Client, repo string, labelConf Conf) error {
	labels, err := client.ListLabels(ctx, repo)
	if err != nil {
		return err
	}
	byName := map[string]*api.Label{}
	for _, l := range labels {
		byName[*l.Name] = l
	}
	for _, def := range labelConf.Default.Labels {
		cur := byName[def.Name]
		if def.Delete {
			if cur != nil {
				if err := client.DeleteLabel(ctx, repo, def.Name); err != nil {
					return err
				}
			}
		} else {
			label := &api.Label{Color: &def.Color, Name: &def.Name, Description: &def.Description}
			if cur == nil {
				if err := client.CreateLabel(ctx, repo, label); err != nil {
					return err
				}
			} else {
				if cur.Color != label.Color || cur.Description != label.Description {
					if err := client.UpdateLabel(ctx, repo, *cur.Name, label); err != nil {
						return err
					}
				}
			}
		}
	}

	return nil
}

func syncMilestones(ctx context.Context, client api.Client, repo string, labelConf Conf) error {
	milestones, err := client.ListMilestones(ctx, repo)
	if err != nil {
		return err
	}
	byTitle := map[string]*api.Milestone{}
	for _, l := range milestones {
		byTitle[*l.Title] = l
	}
	for _, def := range labelConf.Default.Milestones {
		cur := byTitle[def.Title]
		if def.Delete {
			if cur != nil {
				if err := client.DeleteMilestone(ctx, repo, *cur.Number); err != nil {
					return err
				}
			}
		} else {
			c := "open"
			if def.Closed {
				c = "closed"
			}
			milestone := &api.Milestone{Title: &def.Title, Description: &def.Description, State: &c}
			if cur == nil {
				if err := client.CreateMilestone(ctx, repo, milestone); err != nil {
					return err
				}
			} else {
				if cur.State != milestone.State || cur.Description != milestone.Description {
					if err := client.UpdateMilestone(ctx, repo, *cur.Number, milestone); err != nil {
						return err
					}
				}
			}
		}
	}

	return nil
}

func parseConf(path string) (Conf, error) {
	out := Conf{}
	var b []byte
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		r, err := http.Get(path)
		if err != nil {
			return out, err
		}
		if r.StatusCode != 200 {
			return out, fmt.Errorf("invalid status code: %d", r.StatusCode)
		}
		b, err = io.ReadAll(r.Body)
		if err != nil {
			return out, err
		}
	} else {
		var err error
		b, err = os.ReadFile(path)
		if err != nil {
			return out, err
		}
	}
	if err := yaml.Unmarshal(b, &out); err != nil {
		return out, err
	}
	return out, nil
}
