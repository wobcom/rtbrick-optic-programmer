package connection

type I2cRWHandle struct {
	Connection *RouterConnection
	I2cBusId   int
}

func NewI2cRWHandle(user string, router string, iface string) (*I2cRWHandle, error) {
	handle := I2cRWHandle{}

	routerConnection, err := New(user, router)
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

func CloseI2CRWHandle(handle *I2cRWHandle) {
	err := handle.Connection.Close()
	if err != nil {
		panic(err)
	}
}
