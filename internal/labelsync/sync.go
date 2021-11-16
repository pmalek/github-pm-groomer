package labelsync

import (
	"context"
	"fmt"
	"github.com/lahabana/github-pm-groomer/internal/github/api"
	"github.com/lahabana/github-pm-groomer/internal/utils"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"
)

type Opts struct {
	Repo     string
	FilePath string
}

type Conf struct {
	Default RepoConf `yaml:"default"`
}
type RepoConf struct {
	Labels []LabelDef `yamls:"labels"`
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
	return syncLabels(ctx, client, opts.Repo, labelConf)
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
		fmt.Fprintf(os.Stdout, "doing: %v\n", def)
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
		b, err = ioutil.ReadAll(r.Body)
		if err != nil {
			return out, err
		}
	} else {
		var err error
		b, err = ioutil.ReadFile(path)
		if err != nil {
			return out, err
		}
	}
	if err := yaml.Unmarshal(b, &out); err != nil {
		return out, err
	}
	return out, nil
}
