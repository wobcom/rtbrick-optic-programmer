package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/urfave/cli/v3"
	"github.com/wobcom/rtbrick-optic-programmer/internal/pkg/rtbrick"
	connection "github.com/wobcom/rtbrick-optic-programmer/internal/pkg/ssh"
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
				Action: func(ctx context.Context, cmd *cli.Command) error {

					user := cmd.String("user")
					router := cmd.StringArg("device")
					iface := cmd.StringArg("interface")

					routerConnection, err := connection.New(user, router)
					if err != nil {
						return err
					}

					err = routerConnection.Connect()
					if err != nil {
						return err
					}

					defer routerConnection.Close()

					_, ppdConfig, err := routerConnection.GetDeviceInformation()
					if err != nil {
						return err
					}

					var i2cBusId int

					for _, port := range ppdConfig.Ports {
						if port.Name == iface {
							i2cBusId = port.I2CBus
						}
					}

					page00, err := routerConnection.GetI2CDump(i2cBusId, 0x00)
					if err != nil {
						return err
					}
					page12, err := routerConnection.GetI2CDump(i2cBusId, 0x12)
					if err != nil {
						return err
					}
					page1E, err := routerConnection.GetI2CDump(i2cBusId, 0x1E)
					if err != nil {
						return err
					}
					page1B, err := routerConnection.GetI2CDump(i2cBusId, 0x1B)
					if err != nil {
						return err
					}

					resultPage00 := rtbrick.InterpretPage00(page00)
					resultPage12 := rtbrick.InterpretPage12(page12)
					resultPage1E := rtbrick.InterpretPage1E(page1E)
					resultPage1B := rtbrick.InterpretPageB0(page1B)

					slog.Info("module_info", slog.String("vendor_name", resultPage00.VendorName))
					slog.Info("module_info", slog.String("vendor_phy", resultPage00.VendorPN))
					slog.Info("module_info", slog.String("vendor_serial", resultPage00.VendorSN))

					slog.Info("module_info", slog.String(
						"tuning_status", fmt.Sprintf("%b", resultPage12.Status),
					))
					slog.Info("module_info", slog.String("grid_spacing", resultPage12.GridDisplay))
					slog.Info("module_info", slog.Float64("frequency", float64(resultPage12.Frequency)*1e-12))
					slog.Info("module_info", slog.Int("frequency_offset", resultPage12.FrequencyOffset))
					if resultPage12.Channel != nil {
						slog.Info("module_info", slog.Int("channel", *resultPage12.Channel))
					} else {
						slog.Warn("No Valid Channel found!")
					}

					slog.Info("module_info", slog.Bool("flex_tune_enabled", resultPage1E.FlexTuneEnabled))
					slog.Info("module_info", slog.String("power_class_override_status",
						fmt.Sprintf("%x", resultPage1E.PowerClassOverride),
					))
					slog.Info("module_info", slog.Bool("low_pwr_mode_enabled", resultPage00.LowPowerMode))

					slog.Info("module_info", slog.Bool(
						"nominal_wavelength_control_enabled",
						resultPage1B.NominalWavelengthControlEnabled,
					))

					return nil
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
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {

					user := cmd.String("user")
					router := cmd.StringArg("device")
					iface := cmd.StringArg("interface")

					gridSpacing := cmd.Int("grid-spacing")
					channel := cmd.Int("channel")

					routerConnection, err := connection.New(user, router)
					if err != nil {
						return err
					}

					err = routerConnection.Connect()
					if err != nil {
						return err
					}

					defer routerConnection.Close()

					_, ppdConfig, err := routerConnection.GetDeviceInformation()
					if err != nil {
						return err
					}

					var i2cBusId int

					for _, port := range ppdConfig.Ports {
						if port.Name == iface {
							i2cBusId = port.I2CBus
						}
					}

					page00, err := routerConnection.GetI2CDump(i2cBusId, 0x00)
					if err != nil {
						return err
					}
					page12, err := routerConnection.GetI2CDump(i2cBusId, 0x12)
					if err != nil {
						return err
					}
					page1E, err := routerConnection.GetI2CDump(i2cBusId, 0x1E)
					if err != nil {
						return err
					}
					page1B, err := routerConnection.GetI2CDump(i2cBusId, 0x1B)
					if err != nil {
						return err
					}

					resultPage00 := rtbrick.InterpretPage00(page00)
					resultPage12 := rtbrick.InterpretPage12(page12)
					resultPage1E := rtbrick.InterpretPage1E(page1E)
					resultPage1B := rtbrick.InterpretPageB0(page1B)

					slog.Info("module_info", slog.String("vendor_name", resultPage00.VendorName))
					slog.Info("module_info", slog.String("vendor_phy", resultPage00.VendorPN))
					slog.Info("module_info", slog.String("vendor_serial", resultPage00.VendorSN))

					slog.Info("Setting Low Power Mode...")

					wPage, wByte, wValue := rtbrick.GetLowPowerProgramming(true)
					err = routerConnection.DoI2CSet(i2cBusId, wPage, wByte, wValue)
					if err != nil {
						return err
					}

					time.Sleep(1 * time.Second)

					if resultPage1E.PowerClassOverride != 0x01 {
						slog.Info("Setting Power Class Override...")

						wPage, wByte, wValue = rtbrick.GetPowerClassProgramming()
						err = routerConnection.DoI2CSet(i2cBusId, wPage, wByte, wValue)
						if err != nil {
							return err
						}

						time.Sleep(1 * time.Second)
					}

					if resultPage1E.FlexTuneEnabled {
						slog.Info("Disabling Flex Tune...")

						wPage, wByte, wValue = rtbrick.GetFlexTuneProgramming()
						err = routerConnection.DoI2CSet(i2cBusId, wPage, wByte, wValue)
						if err != nil {
							return err
						}

						time.Sleep(1 * time.Second)
					} else {
						slog.Info("Flex Tune is already disabled...")
					}

					needsGridProgramming := resultPage12.GridDisplay != strconv.Itoa(gridSpacing)
					needsChannelProgramming := resultPage12.Channel == nil || *resultPage12.Channel != channel

					if needsGridProgramming {
						slog.Info(
							"grid spacing mismatch, reprogramming",
							slog.Int("target", gridSpacing),
							slog.String("current", resultPage12.GridDisplay),
						)
						wPage, wByte, wValue := rtbrick.GetGridProgramming(gridSpacing)
						err := routerConnection.DoI2CSet(i2cBusId, wPage, wByte, wValue)
						if err != nil {
							return err
						}

					} else {
						slog.Info(
							"grid spacing already matching, will not reprogram",
							slog.String("current", resultPage12.GridDisplay),
						)
					}

					if needsGridProgramming || needsChannelProgramming {
						slog.Info(
							"channel mismatch, reprogramming",
							slog.Int("target", channel),
							slog.Int("current", *resultPage12.Channel),
						)

						wPage, wByte, wValue, wPage2, wByte2, wValue2 := rtbrick.GetChannelProgramming(gridSpacing, channel)
						err := routerConnection.DoI2CSet(i2cBusId, wPage, wByte, wValue)
						if err != nil {
							return err
						}
						err = routerConnection.DoI2CSet(i2cBusId, wPage2, wByte2, wValue2)
						if err != nil {
							return err
						}

					} else {
						slog.Info(
							"channel already matching, will not reprogram",
							slog.Int("current", *resultPage12.Channel),
						)
					}

					time.Sleep(1 * time.Second)

					if !resultPage1B.NominalWavelengthControlEnabled {
						slog.Info("Setting Nominal Wavelength Control Programming...")

						wPage, wByte, wValue = rtbrick.GetNominalWavelengthControlProgramming()
						err = routerConnection.DoI2CSet(i2cBusId, wPage, wByte, wValue)
						if err != nil {
							return err
						}

						time.Sleep(1 * time.Second)
					} else {
						slog.Info("Nominal Wavelength Control is already enabled...")
					}

					slog.Info("Enabling High Power Mode...")

					wPage, wByte, wValue = rtbrick.GetLowPowerProgramming(false)
					err = routerConnection.DoI2CSet(i2cBusId, wPage, wByte, wValue)
					if err != nil {
						return err
					}

					return nil
				},
			},
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		slog.Error("fatal error!", "error", err)
		os.Exit(1)
	}
}
