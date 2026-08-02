package waveform

import "strings"

func differentiatedUnit(unit string) string {
	unit = strings.TrimSpace(unit)
	if unit == "" {
		return ""
	}
	switch unit {
	case "m":
		return "m/s"
	case "m/s":
		return "m/s^2"
	case "m/s^2", "m/s/s":
		return "m/s^3"
	case "cm":
		return "cm/s"
	case "cm/s":
		return "cm/s^2"
	case "cm/s^2", "cm/s/s":
		return "cm/s^3"
	default:
		return unit + "/s"
	}
}

func integratedUnit(unit string) string {
	unit = strings.TrimSpace(unit)
	if unit == "" {
		return ""
	}
	switch unit {
	case "m/s^3":
		return "m/s^2"
	case "m/s^2", "m/s/s":
		return "m/s"
	case "m/s":
		return "m"
	case "cm/s^3":
		return "cm/s^2"
	case "cm/s^2", "cm/s/s":
		return "cm/s"
	case "cm/s":
		return "cm"
	}
	if strings.HasSuffix(unit, "/s/s") {
		return strings.TrimSuffix(unit, "/s")
	}
	if strings.HasSuffix(unit, "/s") {
		return strings.TrimSuffix(unit, "/s")
	}
	return unit + "*s"
}

func powerUnit(unit string) string {
	unit = strings.TrimSpace(unit)
	if unit == "" {
		return ""
	}
	return "(" + unit + ")^2"
}

func powerDensityUnit(unit string) string {
	unit = powerUnit(unit)
	if unit == "" {
		return ""
	}
	return unit + "/Hz"
}
