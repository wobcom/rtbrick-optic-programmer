package routines

import (
	"context"

	"github.com/urfave/cli/v3"
	"github.com/wobcom/rtbrick-optic-programmer/internal/pkg"
	"github.com/wobcom/rtbrick-optic-programmer/internal/pkg/optic"
	"github.com/wobcom/rtbrick-optic-programmer/internal/pkg/rtbrick/ssh"
)

var CmdStringToPowerMode = map[string]optic.PowerMode{
	"high": optic.PowerModeHighPower,
	"low":  optic.PowerModeLowPower,
}

type I2cActionArgs struct {
	Handle *connection.I2cRWHandle
	Cmd    *cli.Command
	Page00 *optic.I2CPage00
	Page12 *optic.I2CPage12
	Page1E *optic.I2CPage1E
	Page1B *optic.I2CPageB0
}

type I2cAction func(args I2cActionArgs) error

func I2cTemplateMethod(actions []I2cAction) cli.ActionFunc {
	return func(_ context.Context, cmd *cli.Command) error {
		user := cmd.String("user")
		router := cmd.StringArg("device")
		iface := cmd.StringArg("interface")

		handle, err := connection.NewI2cRWHandle(user, router, iface)
		if err != nil {
			return err
		}
		defer connection.CloseI2CRWHandle(handle)

		page00, err := handle.Connection.GetI2CDump(handle.I2cBusId)
		if err != nil {
			return err
		}
		page12, err := handle.Connection.GetI2CDump(handle.I2cBusId)
		if err != nil {
			return err
		}
		page1E, err := handle.Connection.GetI2CDump(handle.I2cBusId)
		if err != nil {
			return err
		}
		page1B, err := handle.Connection.GetI2CDump(handle.I2cBusId)
		if err != nil {
			return err
		}

		resultPage00 := optic.InterpretPage00(pkg.ParseI2CDump(*page00))
		resultPage12 := optic.InterpretPage12(pkg.ParseI2CDump(*page12))
		resultPage1E := optic.InterpretPage1E(pkg.ParseI2CDump(*page1E))
		resultPage1B := optic.InterpretPageB0(pkg.ParseI2CDump(*page1B))

		actionArgs := I2cActionArgs{
			Handle: handle,
			Cmd:    cmd,
			Page00: &resultPage00,
			Page12: &resultPage12,
			Page1E: &resultPage1E,
			Page1B: &resultPage1B,
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

var i2cReadActions = [...]I2cAction{
	ActionShowBasicAdminInfo,
	ActionShowTunableLaserStatus,
	ActionShowFlexOptixCustomPages,
}

var I2CReadAll = I2cTemplateMethod(i2cReadActions[:])

var i2cWriteActions = [...]I2cAction{
	ActionShowBasicAdminInfo,
	// some optics require disabling high power before programming
	// ActionSetPowerModeTo(optic.PowerModeLowPower),
	ActionShowFlexOptixCustomPages, // custom
	// ActionDisableFlexTune,                // custom
	// ActionSetGridProgramming,             // tunable laser
	// ActionEnableNominalWavelengthControl, // custom
	// ActionSetPowerClassOverride,          // custom
	// ActionSetPowerMode,
}

var I2CWriteAll = I2cTemplateMethod(i2cWriteActions[:])
