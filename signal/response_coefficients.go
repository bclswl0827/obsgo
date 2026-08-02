package signal

import (
	"fmt"
	"math"
	"math/cmplx"
	"strings"

	"github.com/bclswl0827/obsgo/inventory"
)

func evalCoefficients(value *inventory.Coefficients, frequency, sampleRate float64) (complex128, error) {
	if value == nil {
		return 1, nil
	}
	if len(value.Numerators) == 0 && len(value.Denominators) == 0 {
		// Some inventories use an empty Coefficients element to describe a
		// pure gain/decimation stage. It contributes unity, not zero.
		return 1, nil
	}
	if sampleRate <= 0 {
		return 0, fmt.Errorf("coefficient stage requires a positive sample rate")
	}
	transfer := strings.ToUpper(strings.TrimSpace(value.TransferFunctionType))
	if transfer != "" && transfer != "D" && transfer != "DIGITAL" && transfer != "DIGITAL (Z-TRANSFORM)" && transfer != "DIGITAL(Z-TRANSFORM)" {
		return 0, fmt.Errorf("unsupported coefficient transfer function type %q", value.TransferFunctionType)
	}
	phase := -2 * math.Pi * frequency / sampleRate
	numerator := evaluateDigitalPolynomial(value.Numerators, phase)
	if len(value.Denominators) == 0 {
		return numerator, nil
	}
	denominator := evaluateDigitalPolynomial(value.Denominators, phase)
	if denominator == 0 {
		return 0, nil
	}
	return numerator / denominator, nil
}

func evaluateDigitalPolynomial(coefficients []float64, phase float64) complex128 {
	value := complex(0, 0)
	for index, coefficient := range coefficients {
		value += complex(coefficient, 0) * cmplx.Exp(complex(0, phase*float64(index)))
	}
	return value
}
