package signal

import (
	"math"
	"testing"
)

func TestRotateZNEToLQTAndBack(t *testing.T) {
	z := []float64{1, 2, 3}
	n := []float64{4, 5, 6}
	e := []float64{7, 8, 9}
	l, q, transverse, err := RotateZNEToLQT(z, n, e, 37, 52)
	if err != nil {
		t.Fatal(err)
	}
	zBack, nBack, eBack, err := RotateLQTToZNE(l, q, transverse, 37, 52)
	if err != nil {
		t.Fatal(err)
	}
	for index := range z {
		if math.Abs(zBack[index]-z[index]) > 1e-12 || math.Abs(nBack[index]-n[index]) > 1e-12 || math.Abs(eBack[index]-e[index]) > 1e-12 {
			t.Fatalf("round trip at %d = (%g,%g,%g), want (%g,%g,%g)", index, zBack[index], nBack[index], eBack[index], z[index], n[index], e[index])
		}
	}
}

func TestResampleInterpolation(t *testing.T) {
	data, err := Resample([]float64{0, 1, 2}, 1, 2, LinearInterpolation)
	if err != nil {
		t.Fatal(err)
	}
	want := []float64{0, 0.5, 1, 1.5, 2}
	if len(data) != len(want) {
		t.Fatalf("resampled length = %d, want %d", len(data), len(want))
	}
	for index := range want {
		if math.Abs(data[index]-want[index]) > 1e-12 {
			t.Fatalf("resampled[%d] = %g, want %g", index, data[index], want[index])
		}
	}
}

func TestObsPyInterpolationKernels(t *testing.T) {
	data := []float64{0.92961609, 0.31637555, 0.18391881, 0.20456028, 0.56772503, 0.5955447, 0.96451452, 0.6531771, 0.74890664, 0.65356987}
	got, err := Interpolate1DWithOptions(data, 0, 1, 0, 0.45, 21, LanczosInterpolation, InterpolationOptions{LanczosWidth: 1})
	if err != nil {
		t.Fatal(err)
	}
	want := []float64{0.92961609, 0.55712768, 0.31720733, 0.24275977, 0.17825931, 0.16750234, 0.17561933, 0.20626905, 0.37726064, 0.5647072, 0.47145546, 0.59222238, 0.58665834, 0.91241347, 0.79909224, 0.61631275, 0.61258393, 0.61611633, 0.73239733, 0.56371682, 0.65356987}
	assertFloatSliceClose(t, "ObsPy Lanczos interpolation", got, want, 1e-8)

	monotonic, err := Interpolate1D([]float64{0, 1, 2, 3}, 0, 1, 0, 0.25, 13, WeightedAverageSlopesInterpolation)
	if err != nil {
		t.Fatal(err)
	}
	for index, value := range monotonic {
		if value < -1e-12 || value > 3+1e-12 {
			t.Fatalf("weighted slope interpolation[%d] = %g escaped data range", index, value)
		}
	}
}

func TestSpectrumAndWelch(t *testing.T) {
	data := make([]float64, 32)
	for index := range data {
		data[index] = math.Sin(2 * math.Pi * float64(index) / 8)
	}
	spectrum, err := ComputeSpectrum(data, 32, 32, RectangularWindow)
	if err != nil {
		t.Fatal(err)
	}
	if got := spectrum.Amplitude[4]; math.Abs(got-1) > 1e-12 {
		t.Fatalf("sine amplitude = %g, want 1", got)
	}
	overlap := 0.5
	psd, err := Welch(data, 32, WelchOptions{SegmentLength: 16, OverlapFraction: &overlap, Window: HannWindow})
	if err != nil {
		t.Fatal(err)
	}
	if len(psd.Frequencies) != len(psd.Power) || len(psd.Power) == 0 {
		t.Fatalf("invalid Welch result: %#v", psd)
	}
}

func TestObsPySpectrumGolden(t *testing.T) {
	result, err := ObsPySpectrum([]float64{1, 2, 3, 4}, 4, 4, RectangularWindow)
	if err != nil {
		t.Fatal(err)
	}
	assertFloatSliceClose(t, "ObsPy spectrum frequencies", result.Frequencies, []float64{0, 1, 2, 3}, 0)
	assertFloatSliceClose(t, "ObsPy spectrum power", result.Power, []float64{2, 2, 1, 2}, 1e-12)
}

func TestCrossCorrelation(t *testing.T) {
	lags, values, err := NormalizedCrossCorrelation([]float64{0, 1, 0}, []float64{0, 0, 1}, 2)
	if err != nil {
		t.Fatal(err)
	}
	lag, value, err := CorrelationPeak(lags, values)
	if err != nil {
		t.Fatal(err)
	}
	if lag != -1 || math.Abs(value-5.0/6.0) > 1e-12 {
		t.Fatalf("correlation peak = (%d,%g), want (-1,%g)", lag, value, 5.0/6.0)
	}
}

func TestPolarization(t *testing.T) {
	vertical := make([]float64, 32)
	north := make([]float64, 32)
	east := make([]float64, 32)
	for index := range vertical {
		value := math.Sin(float64(index))
		vertical[index] = value
		north[index] = 2 * value
		east[index] = 0
	}
	result, err := AnalyzePolarization(vertical, north, east)
	if err != nil {
		t.Fatal(err)
	}
	if result.Eigenvalues[0] <= 0 || result.Eigenvalues[1] > 1e-12 || result.Eigenvalues[2] > 1e-12 || result.Rectilinearity < 0.999999 {
		t.Fatalf("unexpected polarization result: %#v", result)
	}
}
