package main

import (
	"context"
	"os"

	"github.com/chirag-bruno/nori/internal/cli"
	urfavecli "github.com/urfave/cli/v3"
)

func main() {
	app := &urfavecli.Command{
		Name:  "nori",
		Usage: "deterministic package manager",
		Commands: []*urfavecli.Command{
			{
				Name:   "init",
				Usage:  "add ~/.nori/shims to PATH",
				Action: cli.InitCommand,
			},
			{
				Name:   "update",
				Usage:  "pull latest registry index + manifests",
				Action: cli.UpdateCommand,
			},
			{
				Name:   "search",
				Usage:  "find packages by name/desc",
				Action: cli.SearchCommand,
			},
			{
				Name:   "info",
				Usage:  "show versions, platforms, bins",
				Action: cli.InfoCommand,
			},
			{
				Name:   "install",
				Usage:  "install for current OS/arch",
				Action: cli.InstallCommand,
			},
			{
				Name:   "use",
				Usage:  "set global active version",
				Action: cli.UseCommand,
			},
			{
				Name:   "list",
				Usage:  "list installed versions for current OS/arch",
				Action: cli.ListCommand,
			},
			{
				Name:   "which",
				Usage:  "show path of the active binary target",
				Action: cli.WhichCommand,
			},
		},
	}

	if err := app.Run(context.Background(), os.Args); err != nil {
		os.Exit(1)
	}
}

