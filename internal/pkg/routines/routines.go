package routines

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"github.com/urfave/cli/v3"
	"github.com/wobcom/rtbrick-optic-programmer/internal/pkg/rtbrick"
	connection "github.com/wobcom/rtbrick-optic-programmer/internal/pkg/ssh"
)

type i2cRWHandle struct {
	connection *connection.RouterConnection
	i2cBusId   int
}

func newI2cRWHandle(user string, router string, iface string) (*i2cRWHandle, error) {
	handle := i2cRWHandle{}

	routerConnection, err := connection.New(user, router)
	if err != nil {
		return nil, err
	}
	err = routerConnection.Connect()
	if err != nil {
		return nil, err
	}
	_, ppdConfig, err := routerConnection.GetDeviceInformation()
	if err != nil {
		return nil, err
	}
	for _, port := range ppdConfig.Ports {
		if port.Name == iface {
			handle.i2cBusId = port.I2CBus
		}
	}

	handle.connection = routerConnection
	return &handle, nil
}

func closeI2CRWHandle(handle *i2cRWHandle) {
	err := handle.connection.Close()
	if err != nil {
		panic(err)
	}
}

// TODO: extract into commmon.go, read.go, write.go
type i2cActionArgs struct {
	handle *i2cRWHandle
	cmd    *cli.Command
	page00 *rtbrick.I2CPage00
	page12 *rtbrick.I2CPage12
	page1E *rtbrick.I2CPage1E
	page1B *rtbrick.I2CPageB0
}

type i2cAction func(args i2cActionArgs) error

func i2cTemplateMethod(actions []i2cAction) cli.ActionFunc {
	return func(_ context.Context, cmd *cli.Command) error {
		user := cmd.String("user")
		router := cmd.StringArg("device")
		iface := cmd.StringArg("interface")

		handle, err := newI2cRWHandle(user, router, iface)
		if err != nil {
			return err
		}
		defer closeI2CRWHandle(handle)

		page00, err := handle.connection.GetI2CDump(handle.i2cBusId, 0x00)
		if err != nil {
			return err
		}
		page12, err := handle.connection.GetI2CDump(handle.i2cBusId, 0x12)
		if err != nil {
			return err
		}
		page1E, err := handle.connection.GetI2CDump(handle.i2cBusId, 0x1E)
		if err != nil {
			return err
		}
		page1B, err := handle.connection.GetI2CDump(handle.i2cBusId, 0x1B)
		if err != nil {
			return err
		}

		resultPage00 := rtbrick.InterpretPage00(page00)
		resultPage12 := rtbrick.InterpretPage12(page12)
		resultPage1E := rtbrick.InterpretPage1E(page1E)
		resultPage1B := rtbrick.InterpretPageB0(page1B)

		actionArgs := i2cActionArgs{
			handle: handle,
			cmd:    cmd,
			page00: &resultPage00,
			page12: &resultPage12,
			page1E: &resultPage1E,
			page1B: &resultPage1B,
		}

		for _, action := range actions {
			err := action(actionArgs)
			if err != nil {
				return err
			}
		}

		return nil
	}
}

func actionShowBasicAdminInfo(args i2cActionArgs) error {
	slog.Info("module_info", slog.String("vendor_name", args.page00.VendorName))
	slog.Info("module_info", slog.String("vendor_phy", args.page00.VendorPN))
	slog.Info("module_info", slog.String("vendor_serial", args.page00.VendorSN))
	slog.Info("module_info", slog.Bool("low_pwr_mode_enabled", args.page00.LowPowerMode))

	return nil
}

func actionShowTunableLaserStatus(args i2cActionArgs) error {
	slog.Info("module_info", slog.String(
		"tuning_status", fmt.Sprintf("%b", args.page12.Status),
	))
	slog.Info("module_info", slog.String("grid_spacing", args.page12.GridDisplay))
	slog.Info("module_info", slog.Float64("frequency", float64(args.page12.Frequency)*1e-12))
	slog.Info("module_info", slog.Int("frequency_offset", args.page12.FrequencyOffset))
	if args.page12.Channel != nil {
		slog.Info("module_info", slog.Int("channel", *args.page12.Channel))
	} else {
		slog.Warn("No Valid Channel found!")
	}

	return nil
}

func actionShowFlexOptixCustomPages(args i2cActionArgs) error {
	// TODO: check with args.page00.VendorName
	slog.Info("module_info", slog.Bool("flex_tune_enabled", args.page1E.FlexTuneEnabled))
	slog.Info("module_info", slog.String("power_class_override_status",
		fmt.Sprintf("%x", args.page1E.PowerClassOverride),
	))
	slog.Info("module_info", slog.Bool(
		"nominal_wavelength_control_enabled",
		args.page1B.NominalWavelengthControlEnabled,
	))

	return nil
}

var i2cReadActions = [...]i2cAction{
	actionShowBasicAdminInfo,
	actionShowTunableLaserStatus,
	actionShowFlexOptixCustomPages,
}

var I2CRead = i2cTemplateMethod(i2cReadActions[:])

var i2cWriteActions = [...]i2cAction{
	actionShowBasicAdminInfo,
	func(
		args i2cActionArgs,
	) error {
		gridSpacing := args.cmd.Float64("grid-spacing")
		channel := args.cmd.Int("channel")

		slog.Info("Setting Low Power Mode...")

		wPage, wByte, wValue := rtbrick.GetLowPowerProgramming(true)
		err := args.handle.connection.DoI2CSet(args.handle.i2cBusId, wPage, wByte, wValue)
		if err != nil {
			return err
		}

		time.Sleep(1 * time.Second)

		if args.page1E.PowerClassOverride != 0x01 {
			slog.Info("Setting Power Class Override...")

			wPage, wByte, wValue = rtbrick.GetPowerClassProgramming()
			err = args.handle.connection.DoI2CSet(args.handle.i2cBusId, wPage, wByte, wValue)
			if err != nil {
				return err
			}

			time.Sleep(1 * time.Second)
		}

		if args.page1E.FlexTuneEnabled {
			slog.Info("Disabling Flex Tune...")

			wPage, wByte, wValue = rtbrick.GetFlexTuneProgramming()
			err = args.handle.connection.DoI2CSet(args.handle.i2cBusId, wPage, wByte, wValue)
			if err != nil {
				return err
			}

			time.Sleep(1 * time.Second)
		} else {
			slog.Info("Flex Tune is already disabled...")
		}

		needsGridProgramming := args.page12.GridDisplay != strconv.FormatFloat(gridSpacing, 'f', 3, 64)
		needsChannelProgramming := args.page12.Channel == nil || *args.page12.Channel != channel

		if needsGridProgramming {
			slog.Info(
				"grid spacing mismatch, reprogramming",
				slog.Float64("target", gridSpacing),
				slog.String("current", args.page12.GridDisplay),
			)
			wPage, wByte, wValue := rtbrick.GetGridProgramming(gridSpacing)
			err := args.handle.connection.DoI2CSet(args.handle.i2cBusId, wPage, wByte, wValue)
			if err != nil {
				return err
			}

		} else {
			slog.Info(
				"grid spacing already matching, will not reprogram",
				slog.String("current", args.page12.GridDisplay),
			)
		}

		if needsGridProgramming || needsChannelProgramming {
			slog.Info(
				"channel mismatch, reprogramming",
				slog.Int("target", channel),
				slog.Int("current", *args.page12.Channel),
			)

			wPage, wByte, wValue, wPage2, wByte2, wValue2 := rtbrick.GetChannelProgramming(gridSpacing, channel)
			err := args.handle.connection.DoI2CSet(args.handle.i2cBusId, wPage, wByte, wValue)
			if err != nil {
				return err
			}
			err = args.handle.connection.DoI2CSet(args.handle.i2cBusId, wPage2, wByte2, wValue2)
			if err != nil {
				return err
			}

		} else {
			slog.Info(
				"channel already matching, will not reprogram",
				slog.Int("current", *args.page12.Channel),
			)
		}

		time.Sleep(1 * time.Second)

		if !args.page1B.NominalWavelengthControlEnabled {
			slog.Info("Setting Nominal Wavelength Control Programming...")

			wPage, wByte, wValue = rtbrick.GetNominalWavelengthControlProgramming()
			err = args.handle.connection.DoI2CSet(args.handle.i2cBusId, wPage, wByte, wValue)
			if err != nil {
				return err
			}

			time.Sleep(1 * time.Second)
		} else {
			slog.Info("Nominal Wavelength Control is already enabled...")
		}

		slog.Info("Enabling High Power Mode...")

		wPage, wByte, wValue = rtbrick.GetLowPowerProgramming(false)
		err = args.handle.connection.DoI2CSet(args.handle.i2cBusId, wPage, wByte, wValue)
		if err != nil {
			return err
		}

		return nil
	},
}

var I2CWrite = i2cTemplateMethod(i2cWriteActions[:])
