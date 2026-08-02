package signal

import (
	"fmt"
	"math"
	"math/cmplx"

	"github.com/bclswl0827/obsgo/inventory"
)

func evalDecimation(value *inventory.Decimation, frequency float64, asymmetricFIR bool) (complex128, float64, error) {
	if value == nil {
		return 1, 0, nil
	}
	if value.Factor < 0 {
		return 0, 0, fmt.Errorf("decimation factor must not be negative")
	}
	factor := value.Factor
	if factor == 0 {
		factor = 1
	}
	if value.InputSampleRate < 0 {
		return 0, 0, fmt.Errorf("decimation input sample rate must not be negative")
	}
	phase := 0.0
	if asymmetricFIR {
		// Evalresp uses the correction-applied field for an asymmetric FIR;
		// the positive sign compensates the FIR's causal phase delay.
		phase = 2 * math.Pi * frequency * value.Correction
	}
	return cmplx.Exp(complex(0, phase)), value.InputSampleRate / float64(factor), nil
}
