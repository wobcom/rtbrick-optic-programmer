# rtbrick-optic-programmer

CMIS & SFF8636 optics programmer for rtbrick remote devices. Features a JSON output for machine parsing.

Tested with Finisar CMIS optics, Finisar SFF8636, FlexOptix SFF8636 C-band programmable.

You are expected to have already set up your ssh key on the device, password entry is currently not supported.

# Building

```shell
go mod download
go build ./cmd/optic-programmer.go
```
# Debug mode
Set the `LOG_LEVEL` env var to "debug" to view i2c dumps as they are read.

# Usage

## Manual
```
NAME:
   optic-programmer - in-field optical module programming for rtbrick

USAGE:
   optic-programmer [global options] [command [command options]]

DESCRIPTION:
   CMIS & SFF8636 optics programmer for rtbrick remote devices. Only works with RBFS,
   uses SMBus & i2c utils to issue direct i2c commands to optical modules.

COMMANDS:
   show           show information about an optic in a specific device. show basic (failsafe) or show all.
   reset          WARNING, YOU MAY LOSE CONFIG!!!! request module software reset
   set, s         sets parameter
   set low-power  toggle low power mode on/off
   set dwdmgrid   sets dwdm grid mode and channel
   help, h        Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --user string        [$USER]
   --device string      [$DEVICE]
   --interface string   [$INTERFACE]
   --help, -h          show help

COPYRIGHT:
   WDZ GmbH 2026
```
## Examples

### Showing basic information
Show basic only shows standard protocol information. Manufacturer protocol extensions are not used with
this command. You can get a detailed output, including protocol and manufacturer extensions,
using the command `show all`.

```shell
./optic-programmer --user llecrivain --device lrma0001.infra.lab.wobcom.de --interface ifp-0/0/2 show basic
{
	"sff8636": {
		"active": true,
		"enable_high_power_class_8": false,
		"enable_high_power_class_57": true,
		"low_pwr_override": true
	},
	"management_protocol": "sff8636",
	"sff_8024_identifier": 17,
	"sff_8024_revision": 8,
	"low_pwr_request_sw": false,
	"software_reset": false,
	"vendor_name": "FLEXOPTIX",
	"vendor_part_number": "Q.16S1HG.14.O2D",
	"vendor_part_revision": "Q.16S1HG.14.O2D",
	"vendor_serial_number": "FQM0023"
}
```

### Programming the tunable laser

Lane and bank are optional, by default the tool will always select first bank, first lane or 1st lane (when
lanes are not banked). Be aware that not specifying grid spacing will reset it to default (100Ghz).

```shell
./optic-programmer --user llecrivain --device lrma0001.infra.lab.wobcom.de --interface ifp-0/0/0 set dwdmgrid \
                   --grid-spacing 50 --channel -20 --lane 1 --bank 1
module has processed command. check module status.
```

## Some jq examples for filtering output
### Checking tuning progress status for lane 0

Tuning complete flag will clear when being read.
```shell
./optic-programmer --user llecrivain --device lrma0001.infra.lab.wobcom.de --interface ifp-0/0/0 show all | \
jq "\
.cmis.tunable_laser.control_status[0] |
{
   grid_spacing_tx: .grid_spacing_tx[0],
   channel_number_tx: .channel_number_tx[0],
   current_laser_frequency_mhz_tx: .current_laser_frequency_mhz_tx[0],
   current_laser_frequency_mhz_tx: .current_laser_frequency_mhz_tx[0],
   tuning_in_progress_tx: .tuning_in_progress_tx[0],
   tuning_complete_flag_tx: .tuning_complete_flag_tx[0]
}"

```
### Only show CMIS tunable laser capabilities
```shell
./optic-programmer --user llecrivain --device lrma0001.infra.lab.wobcom.de --interface ifp-0/0/0 show all | \
jq ".cmis.tunable_laser.capabilities"
```
