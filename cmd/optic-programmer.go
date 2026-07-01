package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"

	"github.com/urfave/cli/v3"
	"github.com/wobcom/rtbrick-optic-programmer/internal/optic"
	cmis2 "github.com/wobcom/rtbrick-optic-programmer/internal/optic/cmis"
	"github.com/wobcom/rtbrick-optic-programmer/internal/optic/default"
	sff8637 "github.com/wobcom/rtbrick-optic-programmer/internal/optic/sff8636"
	"github.com/wobcom/rtbrick-optic-programmer/internal/rtbrick/ssh"
)

const OK = "module has processed command. check module status."

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

var concreteManagementStrategies = [...]func(state *optic.ModuleState) optic.ConcreteManagementStrategy{
	func(state *optic.ModuleState) optic.ConcreteManagementStrategy { return sff8637.New(state) },
	func(state *optic.ModuleState) optic.ConcreteManagementStrategy { return cmis2.New(state) },
}

var concreteExtensionStrategies = [...]func(state *optic.ModuleState) optic.ConcreteExtensionManagementStrategy{
	func(state *optic.ModuleState) optic.ConcreteExtensionManagementStrategy {
		return sff8637.NewSFF8636Extension(state)
	},
	func(state *optic.ModuleState) optic.ConcreteExtensionManagementStrategy {
		return sff8637.NewFlexOptixSFF8636Extension(state)
	},
	func(state *optic.ModuleState) optic.ConcreteExtensionManagementStrategy {
		return cmis2.NewCMISExtension(state)
	},
}

var safeModeConcreteExtensionStrategies = [...]func(state *optic.ModuleState) optic.ConcreteExtensionManagementStrategy{
	// no manufacturers enabled, only lower mem and page 00
	func(state *optic.ModuleState) optic.ConcreteExtensionManagementStrategy {
		return sff8637.NewSFF8636Extension(state)
	},
	func(state *optic.ModuleState) optic.ConcreteExtensionManagementStrategy {
		return cmis2.NewCMISExtension(state)
	},
}

var defaultManagementStrategy = func(state *optic.ModuleState) optic.ConcreteManagementStrategy {
	return _default.New(state)
}

var restrictedFeatureSetFactory = func(handle *connection.I2cRWHandle) *optic.ModuleState {
	return optic.NewModuleState(
		defaultManagementStrategy,
		concreteManagementStrategies[:],
		safeModeConcreteExtensionStrategies[:],
		handle,
	)
}

var allFeatureSetFactory = func(handle *connection.I2cRWHandle) *optic.ModuleState {
	return optic.NewModuleState(
		defaultManagementStrategy,
		concreteManagementStrategies[:],
		concreteExtensionStrategies[:],
		handle,
	)
}

// ActionTemplateMethod allows to factorize common action code and let caller plug callback
// module will always have basic administrative data (safe lower memory and safe page 00) fetched
// before being passed to callback with the rest of the other arguments.
func ActionTemplateMethod(
	moduleFactory func(handle *connection.I2cRWHandle) *optic.ModuleState,
	call func(module *optic.ModuleState, context context.Context, cmd *cli.Command) error) cli.ActionFunc {
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
	slog.SetLogLoggerLevel(slog.LevelWarn)
	if level, ok := os.LookupEnv("LOG_LEVEL"); ok {
		slog.SetLogLoggerLevel(getLoglevel(level))
	}

	cmd := &cli.Command{
		Copyright: "WDZ GmbH 2026",
		Usage:     "in-field optical module programming for rtbrick",
		Description: "CMIS & SFF8636 optics programmer for rtbrick remote devices. Only works with RBFS,\n" +
			"uses SMBus & i2c utils to issue direct i2c commands to optical modules.",
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
				Usage: "show information about an optic in a specific device. show basic (failsafe) or show all.",
				Commands: []*cli.Command{
					{
						Name: "basic",
						Action: ActionTemplateMethod(restrictedFeatureSetFactory, func(
							module *optic.ModuleState,
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
							module *optic.ModuleState,
							context context.Context,
							command *cli.Command,
						) error {
							_, err2 := module.GetExtensionsState()
							if err2 != nil {
								return err2
							}
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
				Name:  "reset",
				Usage: "WARNING, YOU MAY LOSE CONFIG!!!! request module software reset",
				Action: ActionTemplateMethod(restrictedFeatureSetFactory, func(
					module *optic.ModuleState,
					context context.Context,
					command *cli.Command,
				) error {
					module.SoftwareReset = true

					_, err := module.SetAdministrativeInformation()
					if err != nil {
						return err
					}

					println(OK)
					return nil
				}),
			},
			{
				Name:    "set",
				Aliases: []string{"s"},
				Usage:   "sets parameter",
				Commands: []*cli.Command{
					{
						Name:  "low-power",
						Usage: "toggle low power mode on/off",
						Action: ActionTemplateMethod(restrictedFeatureSetFactory, func(
							module *optic.ModuleState,
							context context.Context,
							command *cli.Command,
						) error {
							if command.Args().Len() != 1 {
								return errors.New("please provide on/off argument")
							}
							toggle := command.Args().Get(0)

							if toggle == "on" {
								module.LowPwrRequestSW = true
							} else if toggle == "off" {
								module.LowPwrRequestSW = false
							} else {
								return errors.New(fmt.Sprintf("cannot parse on/off argument %s", toggle))
							}

							_, err := module.SetAdministrativeInformation()
							if err != nil {
								return err
							}

							println(OK)
							return nil
						}),
					},
					{
						Name:  "dwdmgrid",
						Usage: "sets dwdm grid mode and channel",
						Flags: []cli.Flag{
							&cli.Float64Flag{
								Name:  "grid-spacing",
								Usage: "grid to use, in ghz",
								Value: 100.000,
							},
							&cli.IntFlag{
								Name:  "channel",
								Usage: "channel number to use",
							},
							&cli.IntFlag{
								Name:  "lane",
								Usage: "media lane number",
								Value: 1,
							},
							&cli.IntFlag{
								Name:  "bank",
								Usage: "lane bank number to use",
								Value: 1,
							},
						},
						Action: ActionTemplateMethod(allFeatureSetFactory, func(
							module *optic.ModuleState,
							context context.Context,
							command *cli.Command,
						) error {
							_, err := module.GetExtensionsState() // non-basic, fetch extension states first.
							if err != nil {
								return err
							}

							gridSpacing := command.Float64("grid-spacing")
							gridSpacingStr := strconv.FormatFloat(gridSpacing, 'f', 3, 64)
							channel := int16(command.Int("channel"))
							lane := command.Int("lane") - 1
							bank := command.Int("bank") - 1

							// putting it here cos I need to capture args.
							var setChannelAndGrid = func(
								extension *optic.CommonTunableLaserFields,
							) error {
								if !extension.Capabilities.SupportedGridSpacings[gridSpacingStr] {
									panic("Module does not support this frequency.")
								}

								if channel > extension.Capabilities.GridHighChannel[gridSpacingStr] ||
									channel < extension.Capabilities.GridLowChannel[gridSpacingStr] {
									panic("target offset is above or below maximum frequencies for this grid.")
								}

								extension.CtrlStatus[bank].GridSpacingTx[lane] = optic.FloatGhzToCMISGridSpacing[gridSpacingStr]
								extension.CtrlStatus[bank].ChannelNumberTx[lane] = channel

								_, err := module.SetExtensionsState()
								if err != nil {
									return err
								}

								println(OK)

								return nil
							}

							if module.FlexOptixSFF8636.Active {
								err := setChannelAndGrid(&module.FlexOptixSFF8636.TunableLaser)
								if err != nil {
									return err
								}
							} else if module.CMIS.Active {
								err := setChannelAndGrid(&module.CMIS.TunableLaser)
								if err != nil {
									return err
								}
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
