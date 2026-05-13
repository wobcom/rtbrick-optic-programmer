package main

import (
	"context"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/urfave/cli/v3"
	"github.com/wobcom/rtbrick-optic-programmer/internal/pkg/rtbrick"
	connection "github.com/wobcom/rtbrick-optic-programmer/internal/pkg/ssh"
)

func main() {
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

					log.Printf("Vendor Name: %v", resultPage00.VendorName)
					log.Printf("Vendor PN: %v", resultPage00.VendorPN)
					log.Printf("Vendor SN: %v", resultPage00.VendorSN)

					log.Printf("Tuning Status: %b", resultPage12.Status)
					log.Printf("Grid Spacing: %v", resultPage12.GridDisplay)
					log.Printf("Frequency Offset: %v", resultPage12.FrequencyOffset)
					log.Printf("Frequency: %v THz", float64(resultPage12.Frequency)*1e-12)
					if resultPage12.Channel != nil {
						log.Printf("Channel: %v", *resultPage12.Channel)
					} else {
						log.Printf("No Valid Channel!")
					}

					log.Printf("Flex Tune Enabled: %v", resultPage1E.FlexTuneEnabled)
					log.Printf("Power Class Override: %x", resultPage1E.PowerClassOverride)
					log.Printf("Low Power Mode: %v", resultPage00.LowPowerMode)

					log.Printf("Nominal Wavelength Control Enabled: %v", resultPage1B.NominalWavelengthControlEnabled)

					return nil
				},
			},
			{
				Name:    "program",
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

					gridSpacing := cmd.Float64("grid-spacing")
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

					log.Printf("Vendor Name: %v", resultPage00.VendorName)
					log.Printf("Vendor PN: %v", resultPage00.VendorPN)
					log.Printf("Vendor SN: %v", resultPage00.VendorSN)

					log.Printf("Setting Low Power Mode...")

					wPage, wByte, wValue := rtbrick.GetLowPowerProgramming(true)
					err = routerConnection.DoI2CSet(i2cBusId, wPage, wByte, wValue)
					if err != nil {
						return err
					}

					time.Sleep(1 * time.Second)

					if resultPage1E.PowerClassOverride != 0x01 {
						log.Printf("Setting Power Class Override...")

						wPage, wByte, wValue = rtbrick.GetPowerClassProgramming()
						err = routerConnection.DoI2CSet(i2cBusId, wPage, wByte, wValue)
						if err != nil {
							return err
						}

						time.Sleep(1 * time.Second)
					}

					if resultPage1E.FlexTuneEnabled {
						log.Printf("Disabling Flex Tune...")

						wPage, wByte, wValue = rtbrick.GetFlexTuneProgramming()
						err = routerConnection.DoI2CSet(i2cBusId, wPage, wByte, wValue)
						if err != nil {
							return err
						}

						time.Sleep(1 * time.Second)
					} else {
						log.Printf("Flex Tune is already disabled...")
					}

					needsGridProgramming := resultPage12.GridDisplay != strconv.FormatFloat(gridSpacing, 'f', 3, 64)
					needsChannelProgramming := resultPage12.Channel == nil || *resultPage12.Channel != channel

					if needsGridProgramming {
						log.Printf("Grid Spacing must be programmed to %v GHz, currently %v GHz", gridSpacing, resultPage12.GridDisplay)
						wPage, wByte, wValue := rtbrick.GetGridProgramming(gridSpacing)
						err := routerConnection.DoI2CSet(i2cBusId, wPage, wByte, wValue)
						if err != nil {
							return err
						}

					} else {
						log.Printf("Grid Spacing is already programmed at %v GHz, no programming needed...", resultPage12.GridDisplay)
					}

					if needsGridProgramming || needsChannelProgramming {
						log.Printf("Channel must be programmed to %v, currently %v", channel, *resultPage12.Channel)

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
						log.Printf("Channel is already programmed at %v, no programming needed...", *resultPage12.Channel)
					}

					time.Sleep(1 * time.Second)

					if !resultPage1B.NominalWavelengthControlEnabled {
						log.Printf("Setting Nominal Wavelength Control Programming...")

						wPage, wByte, wValue = rtbrick.GetNominalWavelengthControlProgramming()
						err = routerConnection.DoI2CSet(i2cBusId, wPage, wByte, wValue)
						if err != nil {
							return err
						}

						time.Sleep(1 * time.Second)
					} else {
						log.Printf("Nominal Wavelength Control is already enabled...")
					}

					log.Printf("Enabling High Power Mode...")

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
		log.Fatal(err)
	}
}
