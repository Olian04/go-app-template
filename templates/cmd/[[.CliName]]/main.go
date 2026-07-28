// [[ when (modeIs "cli" "cli-library") ]]
package main

import (
	"context"
	"fmt"
	"log/slog"
	"maps"
	"os"
	"os/signal"
	"syscall"

	"github.com/urfave/cli/v3"

	"[[.ModulePath]]/cmd/[[.CliName]]/version"
	"[[.ModulePath]]/internal/config"
	"[[.ModulePath]]/internal/observability/logging"
)

func main() {
	vi := version.Info()

	cli.VersionPrinter = printVersion(vi)

	root := &cli.Command{
		Name:    "[[.CliName]]",
		Usage:   "[[.CliName]] CLI",
		Version: vi.Version,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "config",
				Aliases: []string{"c"},
				Usage:   "path to YAML config file",
				Sources: cli.EnvVars("APP_CONFIG_FILE"),
			},
		},
		Action: runCLI,
	}

	// SIGINT/SIGTERM cancel ctx so long-running commands can stop cleanly and
	// runCLI's deferred logging cleanup still executes.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := root.Run(ctx, os.Args); err != nil {
		// Written straight to stderr, not slog: runCLI's deferred cleanup has
		// already torn the configured logger down by the time Run returns.
		fmt.Fprintf(os.Stderr, "[[.CliName]]: %v\n", err)
		os.Exit(1)
	}
}

func runCLI(ctx context.Context, c *cli.Command) error {
	cfg, err := config.Load(config.Options{
		Path: c.String("config"),
	})
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	logLabels := maps.Clone(map[string]string(cfg.Labels))
	if logLabels == nil {
		logLabels = make(map[string]string, 1)
	}
	logLabels["software_version"] = version.Info().Version

	cleanup, err := logging.Setup(
		logging.WithLevel(cfg.Logging.Level),
		logging.WithFormat(cfg.Logging.Format),
		logging.WithStream(cfg.Logging.Stream),
		logging.WithLabels(logLabels),
	)
	if err != nil {
		return fmt.Errorf("setup logging: %w", err)
	}
	defer cleanup()

	// Replace with real work; ctx carries cancellation from SIGINT/SIGTERM.
	logging.FromContext(ctx).Info("[[.CliName]] ran", "version", version.Info().Version)
	_, err = fmt.Fprintf(c.Root().Writer, "%s %s\n", c.Name, version.Info().Version)
	return err
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
