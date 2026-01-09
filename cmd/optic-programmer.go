package main

import (
	"context"
	"log"
	"os"

	"github.com/urfave/cli/v3"
	"github.com/wobcom/rtbrick-optic-programmer/internal/pkg/rtbrick"
	connection "github.com/wobcom/rtbrick-optic-programmer/internal/pkg/ssh"
)

func main() {
	cmd := &cli.Command{
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "user",
				Usage:   "language for the greeting",
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
					resultPage1B := rtbrick.InterpretPage1B(page1B)

					log.Printf("Vendor Name: %v", resultPage00.VendorName)
					log.Printf("Vendor PN: %v", resultPage00.VendorPN)
					log.Printf("Vendor SN: %v", resultPage00.VendorSN)

					log.Printf("Grid Spacing: %v", resultPage12.GridDisplay)
					log.Printf("Frequency: %v THz", resultPage12.FrequencyDisplay)

					log.Printf("Flex Tune Enabled: %v", resultPage1E.FlexTuneEnabled)
					log.Printf("Low Power Mode: %v", resultPage1E.LowPowerMode)

					log.Printf("Nominal Wavelength Control Enabled: %v", resultPage1B.NominalWavelengthControlEnabled)

					return nil
				},
			},
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}
