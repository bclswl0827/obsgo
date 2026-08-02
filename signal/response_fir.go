package signal

import (
	"fmt"
	"math"
	"math/cmplx"
	"strings"

	"github.com/bclswl0827/obsgo/inventory"
)

func evalFIR(value *inventory.FIR, frequency, sampleRate float64) (complex128, error) {
	if value == nil {
		return 1, nil
	}
	if len(value.Numerators) == 0 {
		return 1, nil
	}
	if sampleRate <= 0 {
		return 0, fmt.Errorf("FIR stage requires a positive sample rate")
	}
	coefficients := value.Numerators
	phase := 2 * math.Pi * frequency / sampleRate
	symmetry := strings.ToUpper(strings.TrimSpace(value.Symmetry))
	switch symmetry {
	case "ODD":
		// Evalresp evaluates symmetric FIR filters about their center and
		// therefore removes their nominal linear phase.
		result := coefficients[len(coefficients)-1]
		for index := 0; index < len(coefficients)-1; index++ {
			result += 2 * coefficients[index] * math.Cos(phase*float64(len(coefficients)-index-1))
		}
		return complex(result, 0), nil
	case "EVEN":
		result := 0.0
		for index, coefficient := range coefficients {
			result += 2 * coefficient * math.Cos(phase*(float64(len(coefficients)-index)-0.5))
		}
		return complex(result, 0), nil
	}
	result := complex(0, 0)
	for index, coefficient := range coefficients {
		result += complex(coefficient, 0) * cmplx.Exp(complex(0, -phase*float64(index)))
	}
	return result, nil
}

func isSymmetricFIR(value *inventory.FIR) bool {
	if value == nil {
		return false
	}
	symmetry := strings.ToUpper(strings.TrimSpace(value.Symmetry))
	return symmetry == "ODD" || symmetry == "EVEN"
}
