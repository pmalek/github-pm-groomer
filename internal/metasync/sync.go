package metasync

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/avast/retry-go"
	"golang.org/x/sync/errgroup"

	"github.com/google/go-github/v67/github"
	"github.com/pmalek/github-pm-groomer/internal/github/api"
	"github.com/pmalek/github-pm-groomer/internal/utils"
	"gopkg.in/yaml.v3"
)

type Opts struct {
	FilePath    string
	Concurrency int
}

type ConfRoot struct {
	Repos  []string `yaml:"repos"`
	Config Conf     `yaml:"config"`
}

type Conf struct {
	Labels     []LabelDef     `yaml:"labels"`
	Milestones []MilestoneDef `yaml:"milestones"`
}

type Time struct {
	time.Time
}

func (t *Time) UnmarshalYAML(value *yaml.Node) error {
	var s string
	if err := value.Decode(&s); err != nil {
		return err
	}
	parsed, err := time.Parse("2006-01-02", s)
	if err != nil {
		return err
	}
	*t = Time{
		Time: parsed.Add(time.Hour * 24),
	}
	return nil
}

type MilestoneDef struct {
	Title       string `yaml:"title"`
	Closed      bool   `yaml:"closed"`
	DueDate     Time   `yaml:"dueDate"`
	Description string `yaml:"description"`
	Delete      bool   `yaml:"delete"`
}

type LabelDef struct {
	Color       string `yaml:"color"`
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	Delete      bool   `yaml:"delete"`
}

func (o Opts) Validate() error {
	if _, err := os.Stat(o.FilePath); err != nil {
		return err
	}
	if o.Concurrency <= 0 {
		return fmt.Errorf("concurrency must be greater than 0, got %d", o.Concurrency)
	}

	return nil
}

func Run(ctx context.Context, client api.Client, opts Opts, now time.Time) error {
	conf, err := parseConf(opts.FilePath)
	if err != nil {
		return err
	}

	// Check if each repo is valid.
	for _, repo := range conf.Repos {
		if _, _, err := utils.OrgRepo(repo); err != nil {
			return err
		}
	}

	for _, repo := range conf.Repos {
		if err := syncLabels(ctx, client, repo, conf, opts.Concurrency); err != nil {
			return err
		}
	}

	for _, repo := range conf.Repos {
		if err := syncMilestones(ctx, client, repo, conf, opts.Concurrency); err != nil {
			return err
		}
	}

	return nil
}

func syncLabels(
	ctx context.Context,
	client api.Client,
	repo string,
	labelConf ConfRoot,
	concurrency int,
) error {
	logger := slog.With(slog.String("repo", repo))
	logger.LogAttrs(ctx, slog.LevelInfo, "sync labels")

	labels, err := client.ListLabels(ctx, repo)
	if err != nil {
		return err
	}
	byName := map[string]*api.Label{}
	for _, l := range labels {
		byName[*l.Name] = l
	}
	errGroup := errgroup.Group{}
	for _, def := range labelConf.Config.Labels {
		errGroup.Go(func() error {
			return retry.Do(func() error {
				logger := logger.With(slog.String("label", def.Name))
				cur := byName[def.Name]

				if def.Delete {
					if cur != nil {
						logger.LogAttrs(ctx, slog.LevelInfo, "deleting label")
						if err := client.DeleteLabel(ctx, repo, def.Name); err != nil {
							return err
						}
					}
					return nil
				}

				label := &api.Label{Color: &def.Color, Name: &def.Name, Description: &def.Description}
				if cur == nil {
					logger.LogAttrs(ctx, slog.LevelInfo, "creating label")
					if err := client.CreateLabel(ctx, repo, label); err != nil {
						return err
					}
					return nil
				}

				if cur.Color != label.Color || cur.Description != label.Description {
					logger.LogAttrs(ctx, slog.LevelInfo, "updating label")
					if err := client.UpdateLabel(ctx, repo, *cur.Name, label); err != nil {
						return err
					}
				}

				return nil
			},
				retry.Context(ctx),
				retry.OnRetry(retryOnRateLimit(ctx)),
				retry.MaxDelay(30*time.Second),
				retry.MaxJitter(3*time.Second),
				retry.DelayType(retry.RandomDelay),
			)
		})
	}
	errGroup.SetLimit(concurrency)
	if err := errGroup.Wait(); err != nil {
		return err
	}

	return nil
}

func syncMilestones(
	ctx context.Context,
	client api.Client,
	repo string,
	labelConf ConfRoot,
	concurrency int,
) error {
	logger := slog.With(slog.String("repo", repo))
	logger.LogAttrs(ctx, slog.LevelInfo, "sync milestones")

	milestones, err := client.ListMilestones(ctx, repo)
	if err != nil {
		return err
	}
	byTitle := map[string]*api.Milestone{}
	for _, l := range milestones {
		byTitle[*l.Title] = l
	}
	errGroup := errgroup.Group{}
	for _, def := range labelConf.Config.Milestones {
		errGroup.Go(func() error {
			retry.Do(func() error {
				logger := logger.With(slog.String("milestone", def.Title))
				cur := byTitle[def.Title]
				if def.Delete {
					if cur != nil {
						logger.LogAttrs(ctx, slog.LevelInfo, "deleting milestone")
						if err := client.DeleteMilestone(ctx, repo, *cur.Number); err != nil {
							return err
						}
					}
					return nil
				}

				c := "open"
				if def.Closed {
					c = "closed"
				}
				milestone := &api.Milestone{
					Title:       &def.Title,
					Description: &def.Description,
					State:       &c,
				}
				if !def.DueDate.IsZero() {
					milestone.DueOn = &github.Timestamp{Time: def.DueDate.Time}
				}
				if cur == nil {
					logger.LogAttrs(ctx, slog.LevelInfo, "creating milestone")
					if err := client.CreateMilestone(ctx, repo, milestone); err != nil {
						return err
					}
					return nil
				}

				if cur.State != milestone.State || cur.Description != milestone.Description {
					logger.LogAttrs(ctx, slog.LevelInfo, "updating milestone")
					if err := client.UpdateMilestone(ctx, repo, *cur.Number, milestone); err != nil {
						return err
					}
				}

				return nil
			},
				retry.Context(ctx),
				retry.OnRetry(retryOnRateLimit(ctx)),
			)
			return nil
		})
	}
	errGroup.SetLimit(concurrency)
	if err := errGroup.Wait(); err != nil {
		return err
	}

	return nil
}

func retryOnRateLimit(ctx context.Context) func(_ uint, err error) {
	return func(n uint, err error) {
		if errRL, ok := err.(*github.RateLimitError); ok {
			resetTS := errRL.Rate.Reset.Time
			slog.Log(ctx, slog.LevelWarn, "hit rate limit", slog.String("reset", resetTS.String()))
			timer := time.NewTimer(time.Until(resetTS))
			defer timer.Stop()

			select {
			case <-timer.C:
				return
			case <-ctx.Done():
				return
			}
		}

		slog.Log(ctx, slog.LevelWarn, "hit rate limit", slog.String("err", err.Error()))
	}
}

func parseConf(path string) (ConfRoot, error) {
	out := ConfRoot{}
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
