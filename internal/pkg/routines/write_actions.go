package routines

import (
	"log/slog"
	"strconv"
	"time"

	"github.com/wobcom/rtbrick-optic-programmer/internal/pkg/rtbrick"
)

func ActionSetPowerModeTo(power rtbrick.PowerMode) I2cAction {
	powerModeToDescription := map[rtbrick.PowerMode]string{
		rtbrick.PowerModeHighPower: "high",
		rtbrick.PowerModeLowPower:  "low",
	}
	return func(args I2cActionArgs) error {
		slog.Info("set_power", slog.String("state", powerModeToDescription[power]))

		wPage, wByte, wValue := rtbrick.GetPowerProgramming(power)
		err := args.Handle.Connection.DoI2CSet(args.Handle.I2cBusId, wPage, wByte, wValue)
		if err != nil {
			return err
		}

		time.Sleep(1 * time.Second)

		return nil
	}
}

func ActionUnconditionallySetPowerClassOverride(args I2cActionArgs) error {
	if args.Page1E.PowerClassOverride != 0x01 {
		slog.Info("Setting Power Class Override...")

		wPage, wByte, wValue := rtbrick.GetPowerClassProgramming()
		err := args.Handle.Connection.DoI2CSet(args.Handle.I2cBusId, wPage, wByte, wValue)
		if err != nil {
			return err
		}

		time.Sleep(1 * time.Second)
	}

	return nil
}

func ActionUnconditionallyDisableFlexTune(args I2cActionArgs) error {
	if args.Page1E.FlexTuneEnabled {
		slog.Info("Disabling Flex Tune...")

		wPage, wByte, wValue := rtbrick.GetFlexTuneProgramming()
		err := args.Handle.Connection.DoI2CSet(args.Handle.I2cBusId, wPage, wByte, wValue)
		if err != nil {
			return err
		}

		time.Sleep(1 * time.Second)
	} else {
		slog.Info("Flex Tune is already disabled...")
	}

	return nil
}

func ActionSetGridProgramming(args I2cActionArgs) error {
	gridSpacing := args.Cmd.Float64("grid-spacing")
	channel := args.Cmd.Int("channel")

	needsGridProgramming := args.Page12.GridDisplay != strconv.FormatFloat(gridSpacing, 'f', 3, 64)
	needsChannelProgramming := args.Page12.Channel == nil || *args.Page12.Channel != channel

	if needsGridProgramming {
		slog.Info(
			"grid spacing mismatch, reprogramming",
			slog.Float64("target", gridSpacing),
			slog.String("current", args.Page12.GridDisplay),
		)
		wPage, wByte, wValue := rtbrick.GetGridProgramming(gridSpacing)
		err := args.Handle.Connection.DoI2CSet(args.Handle.I2cBusId, wPage, wByte, wValue)
		if err != nil {
			return err
		}

	} else {
		slog.Info(
			"grid spacing already matching, will not reprogram",
			slog.String("current", args.Page12.GridDisplay),
		)
	}

	if needsGridProgramming || needsChannelProgramming {
		slog.Info(
			"channel mismatch, reprogramming",
			slog.Int("target", channel),
			slog.Int("current", *args.Page12.Channel),
		)

		wPage, wByte, wValue, wPage2, wByte2, wValue2 := rtbrick.GetChannelProgramming(gridSpacing, channel)
		err := args.Handle.Connection.DoI2CSet(args.Handle.I2cBusId, wPage, wByte, wValue)
		if err != nil {
			return err
		}
		err = args.Handle.Connection.DoI2CSet(args.Handle.I2cBusId, wPage2, wByte2, wValue2)
		if err != nil {
			return err
		}

	} else {
		slog.Info(
			"channel already matching, will not reprogram",
			slog.Int("current", *args.Page12.Channel),
		)
	}

	time.Sleep(1 * time.Second)

	return nil
}

func ActionUnconditionallyEnableNominalWavelengthControl(args I2cActionArgs) error {
	if !args.Page1B.NominalWavelengthControlEnabled {
		slog.Info("Setting Nominal Wavelength Control Programming...")

		wPage, wByte, wValue := rtbrick.GetNominalWavelengthControlProgramming()
		err := args.Handle.Connection.DoI2CSet(args.Handle.I2cBusId, wPage, wByte, wValue)
		if err != nil {
			return err
		}

		time.Sleep(1 * time.Second)
	} else {
		slog.Info("Nominal Wavelength Control is already enabled...")
	}

	return nil
}
