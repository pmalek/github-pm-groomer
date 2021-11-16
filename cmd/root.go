package cmd

import (
	"context"
	"github.com/lahabana/github-pm-groomer/internal/github/api"
	"github.com/spf13/cobra"
	"os"
)

var (
	rootCmd = &cobra.Command{
		Use:   "github-pm-groomer",
		Short: "A CLI to do common product management stuff on github",
	}
	ghClient api.Client
)

func Execute() error {
	ghClient = api.New(os.Getenv("GITHUB_API_TOKEN"))
	if err := ghClient.Ping(context.Background()); err != nil {
		return err
	}
	return rootCmd.Execute()
}
