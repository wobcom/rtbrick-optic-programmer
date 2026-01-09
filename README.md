# rtbrick-optic-programmer

A project to programm Finisar DWDM optics plugged into an RtBrick device.


# Building

```shell
go mod download
go build ./cmd/optic-programmer.go
```

# Usage

```shell
./optic-programmer show --user horst lrmd0001.infra.lab.wobcom.de ifp-0/0/2
2026/01/09 16:54:56 Running command=sudo i2cset -y 33 0x50 127 0
2026/01/09 16:54:56 Running command=sudo i2cdump -y 33 0x50 b
2026/01/09 16:54:57 Running command=sudo i2cset -y 33 0x50 127 0
2026/01/09 16:54:57 Running command=sudo i2cset -y 33 0x50 127 18
2026/01/09 16:54:57 Running command=sudo i2cdump -y 33 0x50 b
2026/01/09 16:54:58 Running command=sudo i2cset -y 33 0x50 127 0
2026/01/09 16:54:58 Running command=sudo i2cset -y 33 0x50 127 30
2026/01/09 16:54:58 Running command=sudo i2cdump -y 33 0x50 b
2026/01/09 16:54:59 Running command=sudo i2cset -y 33 0x50 127 0
2026/01/09 16:54:59 Running command=sudo i2cset -y 33 0x50 127 27
2026/01/09 16:54:59 Running command=sudo i2cdump -y 33 0x50 b
2026/01/09 16:55:00 Running command=sudo i2cset -y 33 0x50 127 0
2026/01/09 16:55:00 Vendor Name: FINISAR CORP.
2026/01/09 16:55:00 Vendor PN: FTLC3351R3PL1
2026/01/09 16:55:00 Vendor SN: 2511W1653
2026/01/09 16:55:00 Grid Spacing: 100
2026/01/09 16:55:00 Frequency Offset: -9
2026/01/09 16:55:00 Frequency: 192.2 THz
2026/01/09 16:55:00 Channel: 22
2026/01/09 16:55:00 Flex Tune Enabled: false
2026/01/09 16:55:00 Low Power Mode: false
2026/01/09 16:55:00 Nominal Wavelength Control Enabled: false
```

```shell
./optic-programmer program --user jwagner lrma0002.infra.lab.wobcom.de ifp-0/0/2 --grid-spacing=100 --channel=22
2026/01/09 16:53:22 Running command=sudo i2cset -y 11 0x50 127 0
2026/01/09 16:53:22 Running command=sudo i2cdump -y 11 0x50 b
2026/01/09 16:53:23 Running command=sudo i2cset -y 11 0x50 127 0
2026/01/09 16:53:23 Running command=sudo i2cset -y 11 0x50 127 18
2026/01/09 16:53:23 Running command=sudo i2cdump -y 11 0x50 b
2026/01/09 16:53:23 Running command=sudo i2cset -y 11 0x50 127 0
2026/01/09 16:53:24 Running command=sudo i2cset -y 11 0x50 127 30
2026/01/09 16:53:24 Running command=sudo i2cdump -y 11 0x50 b
2026/01/09 16:53:24 Running command=sudo i2cset -y 11 0x50 127 0
2026/01/09 16:53:24 Running command=sudo i2cset -y 11 0x50 127 27
2026/01/09 16:53:25 Running command=sudo i2cdump -y 11 0x50 b
2026/01/09 16:53:25 Running command=sudo i2cset -y 11 0x50 127 0
2026/01/09 16:53:25 Vendor Name: FINISAR CORP.
2026/01/09 16:53:25 Vendor PN: FTLC3351R3PL1
2026/01/09 16:53:25 Vendor SN: 2511W1695
2026/01/09 16:53:25 Setting Low Power Mode...
2026/01/09 16:53:25 Running command=sudo i2cset -y 11 0x50 127 0
2026/01/09 16:53:25 Running command=sudo i2cset -y 11 0x50 93 2
2026/01/09 16:53:26 Running command=sudo i2cset -y 11 0x50 127 0
2026/01/09 16:53:27 Flex Tune is already disabled...
2026/01/09 16:53:27 Grid Spacing is already programmed at 100 GHz, no programming needed...
2026/01/09 16:53:27 Channel must be programmed to 22, currently 37
2026/01/09 16:53:27 Running command=sudo i2cset -y 11 0x50 127 18
2026/01/09 16:53:27 Running command=sudo i2cset -y 11 0x50 137 247
2026/01/09 16:53:27 Running command=sudo i2cset -y 11 0x50 127 0
2026/01/09 16:53:27 Running command=sudo i2cset -y 11 0x50 127 18
2026/01/09 16:53:27 Running command=sudo i2cset -y 11 0x50 136 255
2026/01/09 16:53:27 Running command=sudo i2cset -y 11 0x50 127 0
2026/01/09 16:53:28 Setting Nominal Wavelength Control Programming...
2026/01/09 16:53:28 Running command=sudo i2cset -y 11 0x50 127 176
2026/01/09 16:53:28 Running command=sudo i2cset -y 11 0x50 129 1
2026/01/09 16:53:28 Running command=sudo i2cset -y 11 0x50 127 0
2026/01/09 16:53:29 Enabling High Power Mode...
2026/01/09 16:53:29 Running command=sudo i2cset -y 11 0x50 127 0
2026/01/09 16:53:30 Running command=sudo i2cset -y 11 0x50 93 4
2026/01/09 16:53:30 Running command=sudo i2cset -y 11 0x50 127 0


```

