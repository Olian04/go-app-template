// [[ when (modeIs "cli" "cli-library") ]]
package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/urfave/cli/v3"

	"[[.ModulePath]]/cmd/[[.CliName]]/version"
)

func main() {
	vi := version.Info()

	cli.VersionPrinter = printVersion(vi)

	root := &cli.Command{
		Name:    "[[.CliName]]",
		Usage:   "[[.CliName]] CLI",
		Version: vi.Version,
		Action: func(ctx context.Context, c *cli.Command) error {
			_, err := fmt.Fprintf(c.Root().Writer, "%s %s\n", c.Name, vi.Version)
			return err
		},
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := root.Run(ctx, os.Args); err != nil {
		slog.Error("[[.CliName]] exited with error", "error_message", err)
		os.Exit(1)
	}
}

func printVersion(vi version.VersionInfo) func(cmd *cli.Command) {
	return func(cmd *cli.Command) {
		_, err := fmt.Fprintf(cmd.Root().Writer, "%s version %s\nrevision %s\nbuild_time %s\n",
			cmd.Name, vi.Version, vi.Revision, vi.BuildTime)
		if err != nil {
			slog.Error("write version", "error_message", err.Error())
		}
	}
}
