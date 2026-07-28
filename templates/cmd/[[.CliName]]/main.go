// [[ when (modeIs "cli" "cli-library") ]]
package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log/slog"
	"maps"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/urfave/cli/v3"

	"[[.ModulePath]]/cmd/[[.CliName]]/version"
	"[[.ModulePath]]/internal/config"
	"[[.ModulePath]]/internal/domain/echo"
	"[[.ModulePath]]/internal/observability/logging"
)

func main() {
	vi := version.Info()

	cli.VersionPrinter = printVersion(vi)

	root := &cli.Command{
		Name:      "[[.CliName]]",
		Usage:     "Echo a message through the domain model",
		ArgsUsage: "[message...]  (reads stdin when no arguments are given)",
		Version:   vi.Version,
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

	// This is the CLI's IO adapter: read args or stdin, hand the value to the
	// same domain the other modes use, write the result to stdout.
	message, err := readMessage(c)
	if err != nil {
		return err
	}

	res := echo.NewService().Echo(echo.Request{Message: message})

	logging.FromContext(ctx).Info("echoed message", "bytes_in", len(message), "bytes_out", len(res.Message))
	_, err = fmt.Fprintln(c.Root().Writer, res.Message)
	return err
}

// readMessage joins positional args, falling back to stdin so the command
// composes in a pipeline.
//
// Note: urfave/cli trims surrounding whitespace from positional args, so the
// domain's own trimming is only observable on the stdin path.
func readMessage(c *cli.Command) (string, error) {
	if c.Args().Len() > 0 {
		return strings.Join(c.Args().Slice(), " "), nil
	}
	data, err := io.ReadAll(bufio.NewReader(c.Root().Reader))
	if err != nil {
		return "", fmt.Errorf("read stdin: %w", err)
	}
	return string(data), nil
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
