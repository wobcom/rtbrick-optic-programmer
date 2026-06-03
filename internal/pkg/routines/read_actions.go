package routines

import (
	"fmt"
	"log/slog"
)

func ActionShowBasicAdminInfo(args I2cActionArgs) error {
	slog.Info("module_info", slog.String("vendor_name", args.Page00.VendorName))
	slog.Info("module_info", slog.String("vendor_phy", args.Page00.VendorPN))
	slog.Info("module_info", slog.String("vendor_serial", args.Page00.VendorSN))
	slog.Info("module_info", slog.Bool("low_pwr_mode_enabled", args.Page00.LowPowerMode))

	return nil
}

func ActionShowTunableLaserStatus(args I2cActionArgs) error {
	slog.Info("module_info", slog.String(
		"tuning_status", fmt.Sprintf("%b", args.Page12.Status),
	))
	slog.Info("module_info", slog.String("grid_spacing", args.Page12.GridDisplay))
	slog.Info("module_info", slog.Float64("frequency", float64(args.Page12.Frequency)*1e-12))
	slog.Info("module_info", slog.Int("frequency_offset", args.Page12.FrequencyOffset))
	if args.Page12.Channel != nil {
		slog.Info("module_info", slog.Int("channel", *args.Page12.Channel))
	} else {
		slog.Warn("No Valid Channel found!")
	}

	return nil
}

func ActionShowFlexOptixCustomPages(args I2cActionArgs) error {
	// TODO: check with args.page00.VendorName
	slog.Info("module_info", slog.Bool("flex_tune_enabled", args.Page1E.FlexTuneEnabled))
	slog.Info("module_info", slog.String("power_class_override_status",
		fmt.Sprintf("%x", args.Page1E.PowerClassOverride),
	))
	slog.Info("module_info", slog.Bool(
		"nominal_wavelength_control_enabled",
		args.Page1B.NominalWavelengthControlEnabled,
	))

	return nil
}
