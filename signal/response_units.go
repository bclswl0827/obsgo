package signal

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/bclswl0827/obsgo/inventory"
)

// motionUnit describes a physical motion unit in SI terms. Scale is the
// number of SI units represented by one value in the named unit.
type motionUnit struct {
	scale float64
	gamma int
	known bool
}

// parseMotionUnit recognizes the motion units used by StationXML and the
// common aliases emitted by SeisComP and ObsPy. Unknown units are retained as
// dimensionless units; this is the evalresp/"DEF" behavior for non-motion
// quantities such as counts and volts.
func parseMotionUnit(unit string) motionUnit {
	canonical := canonicalMotionUnit(unit)
	switch canonical {
	case "disp", "displacement":
		canonical = "m"
	case "vel", "velocity":
		canonical = "m/s"
	case "acc", "acceleration":
		canonical = "m/s2"
	}
	base := canonical
	gamma := 0
	switch {
	case strings.HasSuffix(base, "/s2"):
		base = strings.TrimSuffix(base, "/s2")
		gamma = 2
	case strings.HasSuffix(base, "/s"):
		base = strings.TrimSuffix(base, "/s")
		gamma = 1
	}
	switch base {
	case "m":
		return motionUnit{scale: 1, gamma: gamma, known: true}
	case "cm":
		return motionUnit{scale: 1e-2, gamma: gamma, known: true}
	case "mm":
		return motionUnit{scale: 1e-3, gamma: gamma, known: true}
	case "um":
		return motionUnit{scale: 1e-6, gamma: gamma, known: true}
	case "nm":
		return motionUnit{scale: 1e-9, gamma: gamma, known: true}
	default:
		return motionUnit{scale: 1, gamma: 0, known: false}
	}
}

func canonicalMotionUnit(unit string) string {
	unit = strings.ToLower(strings.TrimSpace(unit))
	unit = strings.ReplaceAll(unit, "µ", "u")
	unit = strings.ReplaceAll(unit, "μ", "u")
	unit = strings.ReplaceAll(unit, "meters", "m")
	unit = strings.ReplaceAll(unit, "meter", "m")
	unit = strings.ReplaceAll(unit, "metres", "m")
	unit = strings.ReplaceAll(unit, "metre", "m")
	unit = strings.ReplaceAll(unit, "seconds", "s")
	unit = strings.ReplaceAll(unit, "second", "s")
	unit = strings.ReplaceAll(unit, "secs", "s")
	unit = strings.ReplaceAll(unit, "sec", "s")
	unit = strings.ReplaceAll(unit, "per", "/")
	unit = strings.ReplaceAll(unit, "**", "")
	unit = strings.ReplaceAll(unit, "^", "")
	unit = strings.ReplaceAll(unit, "(", "")
	unit = strings.ReplaceAll(unit, ")", "")
	unit = strings.ReplaceAll(unit, "[", "")
	unit = strings.ReplaceAll(unit, "]", "")
	var output strings.Builder
	for _, character := range unit {
		if !unicode.IsSpace(character) {
			output.WriteRune(character)
		}
	}
	unit = output.String()
	unit = strings.ReplaceAll(unit, "/s/s", "/s2")
	return unit
}

func responseUnitName(response *inventory.Response) string {
	if response == nil {
		return ""
	}
	for _, item := range responseStages(response, EvalResponseOptions{}) {
		stage := item.stage
		if stage.PolesZeros != nil && stage.PolesZeros.Input.Name != "" {
			return stage.PolesZeros.Input.Name
		}
		if stage.Coefficients != nil && stage.Coefficients.Input.Name != "" {
			return stage.Coefficients.Input.Name
		}
		if stage.FIR != nil && stage.FIR.Input.Name != "" {
			return stage.FIR.Input.Name
		}
	}
	if response.Sensitivity != nil {
		return response.Sensitivity.Input.Name
	}
	return ""
}

func responseInputUnit(stages []responseStage, response *inventory.Response) string {
	for _, item := range stages {
		stage := item.stage
		if stage.PolesZeros != nil && stage.PolesZeros.Input.Name != "" {
			return stage.PolesZeros.Input.Name
		}
		if stage.Coefficients != nil && stage.Coefficients.Input.Name != "" {
			return stage.Coefficients.Input.Name
		}
		if stage.FIR != nil && stage.FIR.Input.Name != "" {
			return stage.FIR.Input.Name
		}
	}
	return responseUnitName(response)
}

func outputUnitName(output ResponseOutput) (string, error) {
	switch output {
	case Displacement:
		return "m", nil
	case Velocity:
		return "m/s", nil
	case Acceleration:
		return "m/s2", nil
	default:
		return "", fmt.Errorf("unsupported response output %q", output)
	}
}
