package signal

import (
	"fmt"
	"math"
	"math/cmplx"
	"strings"

	"github.com/bclswl0827/obsgo/inventory"
)

func evalPolesZeros(value *inventory.PolesZeros, frequency, sampleRate float64) (complex128, error) {
	if value == nil {
		return 1, nil
	}
	transfer := strings.ToUpper(strings.TrimSpace(value.TransferFunctionType))
	normalization := value.NormalizationFactor
	if normalization == 0 {
		normalization = 1
	}
	var variable complex128
	switch transfer {
	case "", "A", "LAPLACE (RADIANS/SECOND)", "LAPLACE(RADIANS/SECOND)":
		variable = complex(0, 2*math.Pi*frequency)
	case "B", "LAPLACE (HERTZ)", "LAPLACE(HERTZ)":
		variable = complex(0, frequency)
	case "D", "DIGITAL", "DIGITAL (Z-TRANSFORM)", "DIGITAL(Z-TRANSFORM)":
		if sampleRate <= 0 {
			return 0, fmt.Errorf("digital pole-zero stage requires a positive sample rate")
		}
		variable = cmplx.Exp(complex(0, 2*math.Pi*frequency/sampleRate))
	default:
		return 0, fmt.Errorf("unsupported pole-zero transfer function type %q", value.TransferFunctionType)
	}
	numerator := complex(normalization, 0)
	denominator := complex(1, 0)
	for _, zero := range value.Zeros {
		numerator *= variable - complex(zero.Real, zero.Imaginary)
	}
	for _, pole := range value.Poles {
		denominator *= variable - complex(pole.Real, pole.Imaginary)
	}
	if denominator == 0 {
		return 0, nil
	}
	return numerator / denominator, nil
}
