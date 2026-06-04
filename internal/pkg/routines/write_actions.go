package routines

import (
	"log/slog"
	"strconv"
	"time"

	"github.com/pkg/errors"
	"github.com/wobcom/rtbrick-optic-programmer/internal/pkg/optic"
)

func ActionSetPowerModeTo(power optic.PowerMode) I2cAction {
	powerModeToDescription := map[optic.PowerMode]string{
		optic.PowerModeHighPower: "high",
		optic.PowerModeLowPower:  "low",
	}
	return func(args I2cActionArgs) error {
		slog.Info("set_power", slog.String("state", powerModeToDescription[power]))

		wPage, wByte, wValue := optic.GetPowerProgramming(power)
		err := args.Handle.Connection.DoI2CSet(args.Handle.I2cBusId, wPage, wByte, wValue)
		if err != nil {
			return err
		}

		time.Sleep(1 * time.Second)

		return nil
	}
}

func ActionSetPowerMode(args I2cActionArgs) error {
	powerMode, ok := CmdStringToPowerMode[args.Cmd.String("power")]
	if !ok {
		return errors.New("power mode does not exist")
	}
	err := ActionSetPowerModeTo(powerMode)(args)
	if err != nil {
		return err
	}

	return nil
}

func ActionSetPowerClassOverride(args I2cActionArgs) error {
	modeStr := args.Cmd.String("power")
	powerMode, ok := CmdStringToPowerMode[modeStr]
	if !ok {
		return errors.New("power mode does not exist")
	}

	wPage, wByte, wValue := optic.GetPowerClassProgramming(false) // default low
	if powerMode == optic.PowerModeHighPower && args.Page1E.PowerClassOverride != 0x01 {
		wPage, wByte, wValue = optic.GetPowerClassProgramming(true)
	} else if powerMode == optic.PowerModeLowPower && args.Page1E.PowerClassOverride != 0x00 {
		wPage, wByte, wValue = optic.GetPowerClassProgramming(false)
	} else {
		slog.Info("power_class", slog.String("mode", "already_set"))
		return nil
	}

	slog.Info("power_class_set", slog.String("mode", modeStr))
	err := args.Handle.Connection.DoI2CSet(args.Handle.I2cBusId, wPage, wByte, wValue)
	if err != nil {
		return err
	}

	time.Sleep(1 * time.Second)

	return nil
}

func ActionDisableFlexTune(args I2cActionArgs) error {
	if args.Page1E.FlexTuneEnabled {
		slog.Info("Disabling Flex Tune...")

		wPage, wByte, wValue := optic.GetFlexTuneProgramming()
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
		wPage, wByte, wValue := optic.GetGridProgramming(gridSpacing)
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

		wPage, wByte, wValue, wPage2, wByte2, wValue2 := optic.GetChannelProgramming(gridSpacing, channel)
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

func ActionEnableNominalWavelengthControl(args I2cActionArgs) error {
	if !args.Page1B.NominalWavelengthControlEnabled {
		slog.Info("Setting Nominal Wavelength Control Programming...")

		wPage, wByte, wValue := optic.GetNominalWavelengthControlProgramming()
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
