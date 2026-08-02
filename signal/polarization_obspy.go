package signal

import (
	"fmt"
	"math"
)

// PolarizationMethod selects the sliding-window polarization algorithm.
type PolarizationMethod string

const (
	PolarizationFlinn PolarizationMethod = "flinn"
	PolarizationPM    PolarizationMethod = "pm"
)

// PolarizationOptions controls ObsPy-style three-component window analysis.
// WindowLength and Step are sample counts. TaperFraction follows ObsPy's
// cosine taper percentage; zero selects the default 0.22. SmoothLength, when
// positive, smooths each component before windowing. DerivativeKernel is
// applied to per-window eigenvalues and shape measures using endpoint
// replication, matching eigval's centered lfilter preparation.
type PolarizationOptions struct {
	WindowLength     int
	Step             int
	TaperFraction    float64
	DisableTaper     bool
	SmoothLength     int
	DerivativeKernel []float64
	Method           PolarizationMethod
	NoiseThreshold   float64
	UseHilbert       bool
	SampleRate       float64
}

// PolarizationWindowResult contains one sliding-window result and the
// optional auxiliary quantities produced from the result sequence.
type PolarizationWindowResult struct {
	StartSample              int
	CenterSample             int
	Result                   PolarizationResult
	EigenvalueDerivative     [3]float64
	RectilinearityDerivative float64
	PlanarityDerivative      float64
	HilbertFrequency         [3]float64
}

// AnalyzePolarizationWindows performs ObsPy-style fixed sliding-window
// polarization analysis. Incomplete trailing windows are omitted. The Flinn
// method uses covariance eigenstructure; PM uses through-origin orthogonal
// particle-motion slopes.
func AnalyzePolarizationWindows(vertical, north, east []float64, options PolarizationOptions) ([]PolarizationWindowResult, error) {
	if len(vertical) != len(north) || len(vertical) != len(east) {
		return nil, fmt.Errorf("Z, north, and east components must have the same length")
	}
	if options.WindowLength < 2 {
		return nil, fmt.Errorf("polarization window must contain at least two samples")
	}
	if options.Step <= 0 {
		options.Step = options.WindowLength
	}
	if options.Method == "" {
		options.Method = PolarizationFlinn
	}
	if options.Method != PolarizationFlinn && options.Method != PolarizationPM {
		return nil, fmt.Errorf("unsupported polarization method %q", options.Method)
	}
	if options.TaperFraction == 0 {
		options.TaperFraction = 0.22
	}
	if options.TaperFraction < 0 || options.TaperFraction > 1 {
		return nil, fmt.Errorf("polarization taper fraction must be between 0 and 1")
	}
	if options.SmoothLength < 0 {
		return nil, fmt.Errorf("polarization smooth length must not be negative")
	}
	if options.NoiseThreshold < 0 {
		return nil, fmt.Errorf("polarization noise threshold must not be negative")
	}
	if options.UseHilbert && (options.SampleRate <= 0 || math.IsNaN(options.SampleRate) || math.IsInf(options.SampleRate, 0)) {
		return nil, fmt.Errorf("sample rate must be finite and positive when Hilbert quantities are requested")
	}
	processed := [][]float64{append([]float64(nil), vertical...), append([]float64(nil), north...), append([]float64(nil), east...)}
	if options.SmoothLength > 1 {
		for index := range processed {
			processed[index] = movingAverage(processed[index], options.SmoothLength)
		}
	}
	results := make([]PolarizationWindowResult, 0)
	for start := 0; start+options.WindowLength <= len(vertical); start += options.Step {
		window := make([][]float64, 3)
		for component := range window {
			window[component] = append([]float64(nil), processed[component][start:start+options.WindowLength]...)
			if !options.DisableTaper && options.TaperFraction != 0 {
				var err error
				window[component], err = CosineTaper(window[component], options.TaperFraction, true, false)
				if err != nil {
					return nil, err
				}
			}
		}
		result, err := analyzePolarizationWindow(window[0], window[1], window[2], options)
		if err != nil {
			return nil, err
		}
		windowResult := PolarizationWindowResult{
			StartSample:  start,
			CenterSample: start + options.WindowLength/2,
			Result:       result,
		}
		if options.UseHilbert {
			for component := range window {
				frequency, frequencyErr := InstantaneousFrequency(window[component], options.SampleRate)
				if frequencyErr != nil {
					return nil, frequencyErr
				}
				windowResult.HilbertFrequency[component] = Mean(frequency)
			}
		}
		results = append(results, windowResult)
	}
	if len(options.DerivativeKernel) > 0 {
		applyPolarizationDerivatives(results, options.DerivativeKernel)
	}
	return results, nil
}

// InstantaneousFrequency calculates ObsPy's Hilbert/gradient instantaneous
// frequency estimate in Hz. The gradient uses first differences at the ends
// and centered differences in the interior.
func InstantaneousFrequency(data []float64, sampleRate float64) ([]float64, error) {
	if sampleRate <= 0 || math.IsNaN(sampleRate) || math.IsInf(sampleRate, 0) {
		return nil, fmt.Errorf("sample rate must be finite and positive")
	}
	if len(data) == 0 {
		return []float64{}, nil
	}
	analytic := ObsPyAnalyticSignal(data)
	derivative := make([]complex128, len(data))
	if len(data) == 1 {
		derivative[0] = 0
	} else {
		derivative[0] = analytic[1] - analytic[0]
		derivative[len(data)-1] = analytic[len(data)-1] - analytic[len(data)-2]
		for index := 1; index < len(data)-1; index++ {
			derivative[index] = (analytic[index+1] - analytic[index-1]) / 2
		}
	}
	result := make([]float64, len(data))
	for index := range result {
		denominator := 2 * math.Pi * (real(analytic[index])*real(analytic[index]) + imag(analytic[index])*imag(analytic[index]))
		if denominator != 0 {
			result[index] = (real(analytic[index])*imag(derivative[index]) - imag(analytic[index])*real(derivative[index])) * sampleRate / denominator
		}
	}
	return result, nil
}

func analyzePolarizationWindow(vertical, north, east []float64, options PolarizationOptions) (PolarizationResult, error) {
	if options.Method == PolarizationPM {
		return analyzeParticleMotion(vertical, north, east, options.NoiseThreshold)
	}
	if options.NoiseThreshold > 0 {
		filtered := [3][]float64{}
		for index := range vertical {
			if vertical[index]*vertical[index]+north[index]*north[index]+east[index]*east[index] <= options.NoiseThreshold {
				continue
			}
			filtered[0] = append(filtered[0], vertical[index])
			filtered[1] = append(filtered[1], north[index])
			filtered[2] = append(filtered[2], east[index])
		}
		vertical, north, east = filtered[0], filtered[1], filtered[2]
	}
	return AnalyzePolarization(vertical, north, east)
}

func analyzeParticleMotion(vertical, north, east []float64, noiseThreshold float64) (PolarizationResult, error) {
	if len(vertical) < 2 {
		return PolarizationResult{}, fmt.Errorf("polarization requires at least two samples")
	}
	azimuthX := make([]float64, 0, len(vertical))
	azimuthY := make([]float64, 0, len(vertical))
	incidenceX := make([]float64, 0, len(vertical))
	incidenceY := make([]float64, 0, len(vertical))
	for index := range vertical {
		if vertical[index]*vertical[index]+north[index]*north[index]+east[index]*east[index] <= noiseThreshold {
			continue
		}
		horizontal := math.Hypot(north[index], east[index])
		azimuthX = append(azimuthX, east[index])
		azimuthY = append(azimuthY, north[index])
		incidenceX = append(incidenceX, horizontal)
		incidenceY = append(incidenceY, math.Abs(vertical[index]))
	}
	azimuthSlope, azimuthSlopeError, ok := orthogonalSlopeWithError(azimuthX, azimuthY)
	if !ok {
		return PolarizationResult{}, fmt.Errorf("polarization window has insufficient horizontal motion")
	}
	incidenceSlope, incidenceSlopeError, ok := orthogonalSlopeWithError(incidenceX, incidenceY)
	if !ok {
		return PolarizationResult{}, fmt.Errorf("polarization window has insufficient non-zero motion")
	}
	azimuthRadians := math.Atan2(1, azimuthSlope)
	incidenceRadians := math.Atan2(1, incidenceSlope)
	result := PolarizationResult{
		Azimuth:        azimuthRadians / DegreesToRadians,
		Incidence:      incidenceRadians / DegreesToRadians,
		AzimuthError:   slopeAngleRelativeError(azimuthSlope, azimuthSlopeError, azimuthRadians),
		IncidenceError: slopeAngleRelativeError(incidenceSlope, incidenceSlopeError, incidenceRadians),
	}
	if result.Azimuth < 0 {
		result.Azimuth += 360
	}
	if result.Azimuth > 180 {
		result.Azimuth -= 180
	}
	return result, nil
}

// orthogonalSlope returns the major-axis slope of a line constrained through
// the origin. This is the closed-form equivalent of the ODR fit used by
// ObsPy's particle_motion_odr for a one-parameter linear model.
func orthogonalSlope(x, y []float64) (float64, bool) {
	slope, _, ok := orthogonalSlopeWithError(x, y)
	return slope, ok
}

func orthogonalSlopeWithError(x, y []float64) (float64, float64, bool) {
	if len(x) != len(y) || len(x) == 0 {
		return 0, 0, false
	}
	var xx, xy, yy float64
	for index := range x {
		xx += x[index] * x[index]
		xy += x[index] * y[index]
		yy += y[index] * y[index]
	}
	if xx+yy == 0 {
		return 0, 0, false
	}
	var slope float64
	if xy == 0 {
		if yy > xx {
			return math.Inf(1), 0, true
		}
		slope = 0
	} else {
		delta := math.Hypot(xx-yy, 2*xy)
		slope = (yy - xx + delta) / (2 * xy)
	}
	if len(x) <= 1 || math.IsInf(slope, 0) {
		return slope, 0, true
	}
	denominator := math.Pow(1+slope*slope, 1.5)
	var residualSquares, derivativeSquares float64
	for index := range x {
		residual := (y[index] - slope*x[index]) / math.Sqrt(1+slope*slope)
		derivative := -(x[index] + slope*y[index]) / denominator
		residualSquares += residual * residual
		derivativeSquares += derivative * derivative
	}
	if derivativeSquares == 0 {
		return slope, 0, true
	}
	standardError := math.Sqrt(residualSquares / float64(len(x)-1) / derivativeSquares)
	return slope, standardError, true
}

func slopeAngleRelativeError(slope, standardError, angle float64) float64 {
	if angle == 0 || math.IsInf(slope, 0) {
		return 0
	}
	return standardError / ((1 + slope*slope) * angle)
}

func movingAverage(data []float64, length int) []float64 {
	result := make([]float64, len(data))
	half := length / 2
	for index := range result {
		start, end := index-half, index+half+1
		if start < 0 {
			start = 0
		}
		if end > len(data) {
			end = len(data)
		}
		for sample := start; sample < end; sample++ {
			result[index] += data[sample]
		}
		result[index] /= float64(end - start)
	}
	return result
}

func applyPolarizationDerivatives(results []PolarizationWindowResult, kernel []float64) {
	if len(results) == 0 {
		return
	}
	for component := 0; component < 3; component++ {
		values := make([]float64, len(results))
		for index := range results {
			values[index] = results[index].Result.Eigenvalues[component]
		}
		derivative := endpointFilter(values, kernel)
		for index := range results {
			results[index].EigenvalueDerivative[component] = derivative[index]
		}
	}
	for metric := 0; metric < 2; metric++ {
		values := make([]float64, len(results))
		for index := range results {
			if metric == 0 {
				values[index] = results[index].Result.Rectilinearity
			} else {
				values[index] = results[index].Result.Planarity
			}
		}
		derivative := endpointFilter(values, kernel)
		for index := range results {
			if metric == 0 {
				results[index].RectilinearityDerivative = derivative[index]
			} else {
				results[index].PlanarityDerivative = derivative[index]
			}
		}
	}
}

func endpointFilter(values, kernel []float64) []float64 {
	result := make([]float64, len(values))
	if len(values) == 0 || len(kernel) == 0 {
		return result
	}
	half := len(kernel) / 2
	padded := make([]float64, len(values)+2*half)
	for index := range padded {
		source := index - half
		if source < 0 {
			source = 0
		}
		if source >= len(values) {
			source = len(values) - 1
		}
		padded[index] = values[source]
	}
	for index := range result {
		for coefficient := range kernel {
			result[index] += kernel[coefficient] * padded[index+half-coefficient]
		}
	}
	return result
}
