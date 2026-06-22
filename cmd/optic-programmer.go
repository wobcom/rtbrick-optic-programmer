package main

import (
	"context"
	"log/slog"
	"os"
	"strconv"
	"strings"

	"github.com/urfave/cli/v3"
	"github.com/wobcom/rtbrick-optic-programmer/internal/pkg"
	"github.com/wobcom/rtbrick-optic-programmer/internal/pkg/optic"
	"github.com/wobcom/rtbrick-optic-programmer/internal/pkg/optic/cmis"
	"github.com/wobcom/rtbrick-optic-programmer/internal/pkg/optic/default"
	"github.com/wobcom/rtbrick-optic-programmer/internal/pkg/optic/sff8636"
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
	func(state *pkg.ModuleState) pkg.ConcreteExtensionManagementStrategy {
		return cmis.NewCMISExtension(state)
	},
}

var safeModeConcreteExtensionStrategies = [...]func(state *pkg.ModuleState) pkg.ConcreteExtensionManagementStrategy{
	// no manufacturers enabled, only lower mem and page 00
	func(state *pkg.ModuleState) pkg.ConcreteExtensionManagementStrategy {
		return sff8636.NewSFF8636Extension(state)
	},
	func(state *pkg.ModuleState) pkg.ConcreteExtensionManagementStrategy {
		return cmis.NewCMISExtension(state)
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
	call func(module *pkg.ModuleState, context context.Context, cmd *cli.Command) error) cli.ActionFunc {
	return func(context context.Context, cmd *cli.Command) error {
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
		err = call(module, context, cmd)
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
						Action: ActionTemplateMethod(restrictedFeatureSetFactory, func(
							module *pkg.ModuleState,
							context context.Context,
							command *cli.Command,
						) error {
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
						Action: ActionTemplateMethod(allFeatureSetFactory, func(
							module *pkg.ModuleState,
							context context.Context,
							command *cli.Command,
						) error {
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
				Name:    "set",
				Aliases: []string{"s"},
				Usage:   "sets parameter",
				Commands: []*cli.Command{
					{
						Name:        "dwdmgrid",
						Description: "For host-programmable DWDM modules, sets channel from grid",
						Flags: []cli.Flag{
							&cli.Float64Flag{
								Name:  "grid-spacing",
								Usage: "grid to use, in ghz",
							},
							&cli.IntFlag{
								Name:  "channel",
								Usage: "channel number to use",
							},
							&cli.IntFlag{
								Name:  "lane",
								Usage: "media lane number (0-numbered)",
							},
						},
						Action: ActionTemplateMethod(allFeatureSetFactory, func(
							module *pkg.ModuleState,
							context context.Context,
							command *cli.Command,
						) error {
							_, err := module.Get()
							if err != nil {
								return err
							}

							gridSpacing := command.Float64("grid-spacing")
							gridSpacingStr := strconv.FormatFloat(gridSpacing, 'f', 3, 64)
							channel := command.Int("channel")
							lane := command.Int("lane")

							if module.FlexOptixSFF8636Extension.Active {
								extension := module.FlexOptixSFF8636Extension

								if !extension.LaserCapabilities.SupportedFrequencies[gridSpacingStr] {
									panic("Module does not support this frequency.")
								}

								targetFreq := optic.DWDMGridMap[channel]
								gridMultiplier := pkg.MultiplierMap[gridSpacingStr]
								targetOffset := int16((targetFreq - optic.DWDMCenterFreqHz) / gridMultiplier)

								if targetOffset > extension.LaserCapabilities.GridHighChannel[gridSpacingStr] ||
									targetOffset < extension.LaserCapabilities.GridLowChannel[gridSpacingStr] {
									panic("target offset is above or below maximum frequencies for this grid.")
								}

								// flexoptix only supports n = 0
								extension.TunableLaserCtrlStatus.GridSpacingTx[lane] = pkg.FloatGhzToCMISGridSpacing[gridSpacingStr]
								extension.TunableLaserCtrlStatus.ChannelNumberTx[lane] = targetOffset

								_, err := module.SetExtensionsState(module)
								if err != nil {
									return err
								}

								bytes, err := module.ToJson()
								if err != nil {
									return err
								}
								_, err = os.Stdout.Write(bytes)
								if err != nil {
									return err
								}
							} else if module.CMISOnlyExtension.Active {
								// pass for now
							} else {
								panic("Module does not support grid programming")
							}

							return nil
						}),
					},
				},
			},
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		slog.Error("fatal error!", "error", err)
		os.Exit(1)
	}
}
