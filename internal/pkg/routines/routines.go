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

// TODO: refactor into array of actions or something more generic to be able to chain action & err handling to reduce code repetition
// TODO: extract into commmon.go, read.go, write.go
type i2cAction func(*i2cRWHandle, *cli.Command, rtbrick.I2CPage00, rtbrick.I2CPage12, rtbrick.I2CPage1E, rtbrick.I2CPageB0) error

func i2cTemplateMethod(action i2cAction) cli.ActionFunc {
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

		slog.Info("module_info", slog.String("vendor_name", resultPage00.VendorName))
		slog.Info("module_info", slog.String("vendor_phy", resultPage00.VendorPN))
		slog.Info("module_info", slog.String("vendor_serial", resultPage00.VendorSN))

		return action(handle, cmd, resultPage00, resultPage12, resultPage1E, resultPage1B)
	}
}

var I2CRead = i2cTemplateMethod(
	func(
		_ *i2cRWHandle,
		_ *cli.Command,
		resultPage00 rtbrick.I2CPage00,
		resultPage12 rtbrick.I2CPage12,
		resultPage1E rtbrick.I2CPage1E,
		resultPage1B rtbrick.I2CPageB0,
	) error {
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
	})

var I2CWrite = i2cTemplateMethod(
	func(
		handle *i2cRWHandle,
		cmd *cli.Command,
		resultPage00 rtbrick.I2CPage00,
		resultPage12 rtbrick.I2CPage12,
		resultPage1E rtbrick.I2CPage1E,
		resultPage1B rtbrick.I2CPageB0,
	) error {
		gridSpacing := cmd.Float64("grid-spacing")
		channel := cmd.Int("channel")

		slog.Info("Setting Low Power Mode...")

		wPage, wByte, wValue := rtbrick.GetLowPowerProgramming(true)
		err := handle.connection.DoI2CSet(handle.i2cBusId, wPage, wByte, wValue)
		if err != nil {
			return err
		}

		time.Sleep(1 * time.Second)

		if resultPage1E.PowerClassOverride != 0x01 {
			slog.Info("Setting Power Class Override...")

			wPage, wByte, wValue = rtbrick.GetPowerClassProgramming()
			err = handle.connection.DoI2CSet(handle.i2cBusId, wPage, wByte, wValue)
			if err != nil {
				return err
			}

			time.Sleep(1 * time.Second)
		}

		if resultPage1E.FlexTuneEnabled {
			slog.Info("Disabling Flex Tune...")

			wPage, wByte, wValue = rtbrick.GetFlexTuneProgramming()
			err = handle.connection.DoI2CSet(handle.i2cBusId, wPage, wByte, wValue)
			if err != nil {
				return err
			}

			time.Sleep(1 * time.Second)
		} else {
			slog.Info("Flex Tune is already disabled...")
		}

		needsGridProgramming := resultPage12.GridDisplay != strconv.FormatFloat(gridSpacing, 'f', 3, 64)
		needsChannelProgramming := resultPage12.Channel == nil || *resultPage12.Channel != channel

		if needsGridProgramming {
			slog.Info(
				"grid spacing mismatch, reprogramming",
				slog.Float64("target", gridSpacing),
				slog.String("current", resultPage12.GridDisplay),
			)
			wPage, wByte, wValue := rtbrick.GetGridProgramming(gridSpacing)
			err := handle.connection.DoI2CSet(handle.i2cBusId, wPage, wByte, wValue)
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
			err := handle.connection.DoI2CSet(handle.i2cBusId, wPage, wByte, wValue)
			if err != nil {
				return err
			}
			err = handle.connection.DoI2CSet(handle.i2cBusId, wPage2, wByte2, wValue2)
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
			err = handle.connection.DoI2CSet(handle.i2cBusId, wPage, wByte, wValue)
			if err != nil {
				return err
			}

			time.Sleep(1 * time.Second)
		} else {
			slog.Info("Nominal Wavelength Control is already enabled...")
		}

		slog.Info("Enabling High Power Mode...")

		wPage, wByte, wValue = rtbrick.GetLowPowerProgramming(false)
		err = handle.connection.DoI2CSet(handle.i2cBusId, wPage, wByte, wValue)
		if err != nil {
			return err
		}

		return nil
	})
