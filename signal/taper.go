package signal

import (
	"fmt"
	"math"
)

// Taper applies a symmetric time-domain taper to a copy of data.
func Taper(data []float64, width float64, taperType TaperType) ([]float64, error) {
	if math.IsNaN(width) || math.IsInf(width, 0) || width < 0 || width > 0.5 {
		return nil, fmt.Errorf("taper width must be between 0 and 0.5")
	}
	output := append([]float64(nil), data...)
	taperLength := int(math.Floor(float64(len(data)) * width))
	if taperLength == 0 {
		return output, nil
	}
	omega, f0, f1, err := TaperCoefficients(taperType, taperLength)
	if err != nil {
		return nil, err
	}
	for index := 0; index < taperLength; index++ {
		factor := f0 - f1*math.Cos(omega*float64(index))
		output[index] *= factor
		output[len(output)-index-1] *= factor
	}
	return output, nil
}

func TaperCoefficients(taperType TaperType, length int) (omega, f0, f1 float64, err error) {
	if length <= 0 {
		return 0, 0, 0, fmt.Errorf("taper length must be positive")
	}
	switch taperType {
	case Hanning:
		return math.Pi / float64(length), 0.5, 0.5, nil
	case Hamming:
		return math.Pi / float64(length), 0.54, 0.46, nil
	default:
		// SeisPlotJS treats every non-HANNING/non-HAMMING type as COSINE.
		return math.Pi / (2 * float64(length)), 1, 1, nil
	}
}
