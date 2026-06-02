package main

import (
	"context"
	"log/slog"
	"os"
	"strings"

	"github.com/urfave/cli/v3"
	"github.com/wobcom/rtbrick-optic-programmer/internal/pkg/routines"
)

func getLoglevel(level string) slog.Level {
	switch strings.ToLower(level) {
	case "error":
		return slog.LevelError
	case "warn":
		return slog.LevelWarn
	case "info":
		return slog.LevelInfo
	case "debug":
		return slog.LevelDebug
	}
	panic("invalid log level provided")
}

func main() {
	slog.SetLogLoggerLevel(slog.LevelInfo)
	if level, ok := os.LookupEnv("LOG_LEVEL"); ok {
		slog.SetLogLoggerLevel(getLoglevel(level))
	}

	cmd := &cli.Command{
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "user",
				Sources: cli.EnvVars("USER"),
			},
		},
		Commands: []*cli.Command{
			{
				Name:    "show",
				Aliases: []string{"s"},
				Usage:   "Shows information about an optic in a specific device",
				Arguments: []cli.Argument{
					&cli.StringArg{
						Name: "device",
					},
					&cli.StringArg{
						Name: "interface",
					},
				},
				Action: routines.I2cRead,
			},
			{
				Name:    "program",
				Aliases: []string{"s"},
				Usage:   "Programs an optic in a specific device",
				Arguments: []cli.Argument{
					&cli.StringArg{
						Name: "device",
					},
					&cli.StringArg{
						Name: "interface",
					},
				},
				Flags: []cli.Flag{
					&cli.IntFlag{
						Name: "grid-spacing",
					},
					&cli.IntFlag{
						Name: "channel",
					},
				},
				Action: routines.I2cWrite,
			},
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		slog.Error("fatal error!", "error", err)
		os.Exit(1)
	}
}
