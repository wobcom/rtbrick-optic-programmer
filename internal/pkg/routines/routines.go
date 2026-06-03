package routines

import (
	"context"

	"github.com/urfave/cli/v3"
	"github.com/wobcom/rtbrick-optic-programmer/internal/pkg/rtbrick"
	connection "github.com/wobcom/rtbrick-optic-programmer/internal/pkg/ssh"
)

type I2cRWHandle struct {
	Connection *connection.RouterConnection
	I2cBusId   int
}

func newI2cRWHandle(user string, router string, iface string) (*I2cRWHandle, error) {
	handle := I2cRWHandle{}

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
			handle.I2cBusId = port.I2CBus
		}
	}

	handle.Connection = routerConnection
	return &handle, nil
}

func closeI2CRWHandle(handle *I2cRWHandle) {
	err := handle.Connection.Close()
	if err != nil {
		panic(err)
	}
}

type I2cActionArgs struct {
	Handle *I2cRWHandle
	Cmd    *cli.Command
	Page00 *rtbrick.I2CPage00
	Page12 *rtbrick.I2CPage12
	Page1E *rtbrick.I2CPage1E
	Page1B *rtbrick.I2CPageB0
}

type I2cAction func(args I2cActionArgs) error

func I2cTemplateMethod(actions []I2cAction) cli.ActionFunc {
	return func(_ context.Context, cmd *cli.Command) error {
		user := cmd.String("user")
		router := cmd.StringArg("device")
		iface := cmd.StringArg("interface")

		handle, err := newI2cRWHandle(user, router, iface)
		if err != nil {
			return err
		}
		defer closeI2CRWHandle(handle)

		page00, err := handle.Connection.GetI2CDump(handle.I2cBusId, 0x00)
		if err != nil {
			return err
		}
		page12, err := handle.Connection.GetI2CDump(handle.I2cBusId, 0x12)
		if err != nil {
			return err
		}
		page1E, err := handle.Connection.GetI2CDump(handle.I2cBusId, 0x1E)
		if err != nil {
			return err
		}
		page1B, err := handle.Connection.GetI2CDump(handle.I2cBusId, 0x1B)
		if err != nil {
			return err
		}

		resultPage00 := rtbrick.InterpretPage00(page00)
		resultPage12 := rtbrick.InterpretPage12(page12)
		resultPage1E := rtbrick.InterpretPage1E(page1E)
		resultPage1B := rtbrick.InterpretPageB0(page1B)

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
	ActionSetPowerModeTo(rtbrick.PowerModeLowPower),
	ActionEnablePowerClassOverride,
	ActionDisableFlexTune,
	ActionSetGridProgramming,
	ActionEnableNominalWavelengthControl,
	ActionSetPowerModeTo(rtbrick.PowerModeHighPower),
}

var I2CWriteAll = I2cTemplateMethod(i2cWriteActions[:])
