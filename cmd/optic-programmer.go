package main

import (
	"context"
	"log/slog"
	"os"
	"strings"

	"github.com/urfave/cli/v3"
	"github.com/wobcom/rtbrick-optic-programmer/internal/pkg"
	"github.com/wobcom/rtbrick-optic-programmer/internal/pkg/optic/cmis"
	"github.com/wobcom/rtbrick-optic-programmer/internal/pkg/optic/default"
	"github.com/wobcom/rtbrick-optic-programmer/internal/pkg/optic/sff8636"
	"github.com/wobcom/rtbrick-optic-programmer/internal/pkg/routines"
	connection "github.com/wobcom/rtbrick-optic-programmer/internal/pkg/rtbrick/ssh"
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

var concreteManagementStrategies = [...]func(state *pkg.ModuleState) pkg.ConcreteManagementStrategy{
	func(state *pkg.ModuleState) pkg.ConcreteManagementStrategy { return sff8636.New(state) },
	func(state *pkg.ModuleState) pkg.ConcreteManagementStrategy { return cmis.New(state) },
}

var concreteExtensionStrategies = [...]func(state *pkg.ModuleState) pkg.ConcreteExtensionManagementStrategy{
	func(state *pkg.ModuleState) pkg.ConcreteExtensionManagementStrategy {
		return sff8636.NewSFF8636Extension(state)
	},
	func(state *pkg.ModuleState) pkg.ConcreteExtensionManagementStrategy {
		return sff8636.NewFlexOptixSFF8636Extension(state)
	},
}

var safeModeConcreteExtensionStrategies = [...]func(state *pkg.ModuleState) pkg.ConcreteExtensionManagementStrategy{
	// no manufacturers enabled, only lower mem and page 00
	func(state *pkg.ModuleState) pkg.ConcreteExtensionManagementStrategy {
		return sff8636.NewSFF8636Extension(state)
	},
}

var defaultManagementStrategy = func(state *pkg.ModuleState) pkg.ConcreteManagementStrategy {
	return _default.New(state)
}

var restrictedFeatureSetFactory = func(handle *connection.I2cRWHandle) *pkg.ModuleState {
	return pkg.NewModuleState(
		defaultManagementStrategy,
		concreteManagementStrategies[:],
		safeModeConcreteExtensionStrategies[:],
		handle,
	)
}

var allFeatureSetFactory = func(handle *connection.I2cRWHandle) *pkg.ModuleState {
	return pkg.NewModuleState(
		defaultManagementStrategy,
		concreteManagementStrategies[:],
		concreteExtensionStrategies[:],
		handle,
	)
}

func ActionTemplateMethod(
	moduleFactory func(handle *connection.I2cRWHandle) *pkg.ModuleState,
	call func(module *pkg.ModuleState) error) cli.ActionFunc {
	return func(_ context.Context, cmd *cli.Command) error {
		user := cmd.String("user")
		router := cmd.String("device")
		iface := cmd.String("interface")

		handle, err := connection.NewI2cRWHandle(user, router, iface)
		if err != nil {
			return err
		}
		defer connection.CloseI2CRWHandle(handle)

		module, err := moduleFactory(handle).Get()
		if err != nil {
			return err
		}
		err = call(module)
		if err != nil {
			return err
		}

		return nil
	}
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
			&cli.StringFlag{
				Name:    "device",
				Sources: cli.EnvVars("DEVICE"),
			},
			&cli.StringFlag{
				Name:    "interface",
				Sources: cli.EnvVars("INTERFACE"),
			},
		},
		Commands: []*cli.Command{
			{
				Name:  "show",
				Usage: "Shows information about an optic in a specific device",
				Commands: []*cli.Command{
					{
						Name: "basic",
						Action: ActionTemplateMethod(restrictedFeatureSetFactory, func(module *pkg.ModuleState) error {
							bytes, err := module.ToJson()
							if err != nil {
								return err
							}
							_, err = os.Stdout.Write(bytes)
							if err != nil {
								return err
							}
							return nil
						}),
					},
					{
						Name: "all",
						Action: ActionTemplateMethod(allFeatureSetFactory, func(module *pkg.ModuleState) error {
							bytes, err := module.ToJson()
							if err != nil {
								return err
							}
							_, err = os.Stdout.Write(bytes)
							if err != nil {
								return err
							}
							return nil
						}),
					},
				},
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
					&cli.StringFlag{
						Name:  "power",
						Value: "low",
					},
				},
				Action: routines.I2CWriteAll,
			},
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		slog.Error("fatal error!", "error", err)
		os.Exit(1)
	}
}
