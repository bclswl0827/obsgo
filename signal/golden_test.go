package signal

import (
	"math"
	"testing"
)

func TestOregonIIRButterworthGolden(t *testing.T) {
	testOregonIIRGolden(t, "butterworth", func() (*IIRFilter, error) {
		return NewButterworth(1, LowPass, 0, 1, 0.1)
	}, []float64{
		0.24523727525278555,
		0.24523727525278555,
		0,
		-0.5095254494944288,
		0,
	}, []float64{
		0.24523727595806122,
		0.37019190192222595,
		0.18862219154834747,
		0.09610781073570251,
	})
}

func TestOregonIIRChebyshevIGolden(t *testing.T) {
	testOregonIIRGolden(t, "chebyshev I", func() (*IIRFilter, error) {
		return NewChebyshevI(1, 0.5, LowPass, 0, 1, 0.1)
	}, []float64{
		0.3938803955302445,
		0.3938803955302445,
		0,
		-0.21223920893951095,
		0,
	}, []float64{
		0.3938803970813751,
		0.4774772524833679,
		0.10133939236402512,
		0.021508192643523216,
	})
}

func TestOregonIIRChebyshevIIGolden(t *testing.T) {
	testOregonIIRGolden(t, "chebyshev II", func() (*IIRFilter, error) {
		return NewChebyshevII(1, 0.5, LowPass, 0, 1, 0.1)
	}, []float64{
		0.13975523402351372,
		0.13975523402351372,
		0,
		-0.7204895319529725,
		0,
	}, []float64{
		0.1397552341222763,
		0.24044741690158844,
		0.1732398420572281,
		0.12481749802827835,
	})
}

func testOregonIIRGolden(t *testing.T, name string, create func() (*IIRFilter, error), wantSection, wantOutput []float64) {
	t.Helper()
	filter, err := create()
	if err != nil {
		t.Fatalf("create %s filter: %v", name, err)
	}
	if len(filter.sections) != 1 {
		t.Fatalf("%s section count = %d, want 1", name, len(filter.sections))
	}
	section := filter.sections[0]
	gotSection := []float64{section.b0, section.b1, section.b2, section.a1, section.a2}
	assertFloatSliceClose(t, name+" coefficients", gotSection, wantSection, 1e-12)
	got := filter.Apply([]float64{1, 0, 0, 0})
	assertFloatSliceClose(t, name+" impulse response", got, wantOutput, 1e-7)
}

func TestOregonIIRChunkedProcessing(t *testing.T) {
	input := make([]float64, 97)
	for i := range input {
		input[i] = math.Sin(0.13*float64(i)) + 0.25*math.Cos(0.037*float64(i))
	}

	filters := []struct {
		name   string
		create func() (*IIRFilter, error)
	}{
		{"Butterworth", func() (*IIRFilter, error) {
			return NewButterworth(4, BandPass, 0.5, 8, 0.01)
		}},
		{"Chebyshev I", func() (*IIRFilter, error) {
			return NewChebyshevI(4, 0.5, BandPass, 0.5, 8, 0.01)
		}},
		{"Chebyshev II", func() (*IIRFilter, error) {
			return NewChebyshevII(4, 0.5, BandPass, 0.5, 8, 0.01)
		}},
	}
	for _, testCase := range filters {
		oneShot, err := testCase.create()
		if err != nil {
			t.Fatal(err)
		}
		chunked, err := testCase.create()
		if err != nil {
			t.Fatal(err)
		}
		want := oneShot.Apply(input)
		got := make([]float64, 0, len(input))
		for _, end := range []int{7, 31, 64, len(input)} {
			start := len(got)
			got = append(got, chunked.Apply(input[start:end])...)
		}
		assertFloatSliceClose(t, testCase.name+" chunked processing", got, want, 0)
	}
}

func TestCenteredHilbertOregonBehavior(t *testing.T) {
	filter, err := NewCenteredHilbertTransform(10, 0.05, 0.95)
	if err != nil {
		t.Fatal(err)
	}
	coefficients := filter.Coefficients()
	if len(coefficients) != 21 {
		t.Fatalf("coefficient count = %d, want 21", len(coefficients))
	}
	for i := range coefficients {
		if math.IsNaN(coefficients[i]) || math.IsInf(coefficients[i], 0) {
			t.Fatalf("coefficient %d is not finite: %v", i, coefficients[i])
		}
		if math.Abs(coefficients[i]+coefficients[len(coefficients)-1-i]) > 1e-7 {
			t.Fatalf("coefficient %d is not antisymmetric", i)
		}
	}
	if math.Abs(coefficients[10]) > 1e-7 {
		t.Fatalf("center coefficient = %g, want zero", coefficients[10])
	}

	input := []float64{1, 2, 3, 4}
	output := filter.Filter(input)
	if len(output) != len(input)+20 {
		t.Fatalf("Hilbert output length = %d, want %d", len(output), len(input)+20)
	}
	for i, value := range output {
		if math.IsNaN(value) || math.IsInf(value, 0) {
			t.Fatalf("Hilbert output %d is not finite: %v", i, value)
		}
	}
	impulse := filter.Filter([]float64{1})
	assertFloatSliceClose(t, "Hilbert impulse", impulse, coefficients, 1e-7)
}

func TestFFTSeisPlotJSSmallInputGolden(t *testing.T) {
	got := CalcDFT([]float64{1, 2, 3})
	want := []float64{
		6,
		4.969079494476318,
		2.4142136573791504,
		-0.35595348477363586,
		-2,
		-1.886687159538269,
		-0.4142135679721832,
		1.2735612392425537,
		2,
		1.3559534549713135,
		1.5857864618301392,
		0.2735612690448761,
		-2,
		-3.9690794944763184,
		-4.41421365737915,
		-2.8866872787475586,
	}
	assertFloatSliceClose(t, "small FFT", got, want, 1e-7)

	inverse, err := InverseDFT(got, 3)
	if err != nil {
		t.Fatal(err)
	}
	assertFloatSliceClose(t, "small FFT inverse", inverse, []float64{1, 2, 3}, 1e-7)
	if _, err := InverseDFT([]float64{1, 2}, 1); err == nil {
		t.Fatal("inverse FFT accepted a packed transform shorter than OregonDSP's minimum length")
	}
}

func TestTaperSeisPlotJSBoundaries(t *testing.T) {
	got, err := Taper([]float64{1, 1, 1, 1}, 0.5, Hanning)
	if err != nil {
		t.Fatal(err)
	}
	assertFloatSliceClose(t, "Hanning taper", got, []float64{0, 0.5, 0.5, 0}, 1e-15)

	if _, err = Taper([]float64{1, 1, 1, 1}, -0.25, Hanning); err == nil {
		t.Fatal("negative taper width was accepted")
	}

	got, err = Taper([]float64{1, 1, 1, 1}, 0.5, TaperType("unknown"))
	if err != nil {
		t.Fatal(err)
	}
	cosineEdge := 1 - math.Sqrt(0.5)
	assertFloatSliceClose(t, "fallback cosine taper", got, []float64{0, cosineEdge, cosineEdge, 0}, 1e-15)
	if _, err := Taper([]float64{1, 1}, 0.5001, Hanning); err == nil {
		t.Fatal("taper accepted width greater than 0.5")
	}
}

func TestObsPyCosineTaperGolden(t *testing.T) {
	got, err := CosineTaper([]float64{1, 1, 1, 1, 1, 1}, 1, true, false)
	if err != nil {
		t.Fatal(err)
	}
	assertFloatSliceClose(t, "ObsPy cosine taper", got, []float64{0, 0.5, 1, 1, 0.5, 0}, 1e-15)
	for _, percentage := range []float64{-0.1, 1.1, math.NaN(), math.Inf(1)} {
		if _, err := CosineTaper([]float64{1, 1}, percentage, true, false); err == nil {
			t.Fatalf("ObsPy cosine taper accepted percentage %g", percentage)
		}
	}
	for _, percentage := range []float64{-0.1, 0.51, math.NaN(), math.Inf(1)} {
		if _, err := TraceCosineTaper([]float64{1, 1}, percentage); err == nil {
			t.Fatalf("ObsPy trace taper accepted percentage %g", percentage)
		}
	}
}

func TestCalculusBoundaries(t *testing.T) {
	got, err := Differentiate([]float64{1, 3, 6}, 2)
	if err != nil {
		t.Fatal(err)
	}
	assertFloatSliceClose(t, "differentiate", got, []float64{4, 6}, 0)

	got, err = Differentiate([]float64{1}, 2)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 0 {
		t.Fatalf("single-sample differentiate length = %d, want 0", len(got))
	}

	got, err = DifferentiateGradient([]float64{1, 3, 7, 13}, 2)
	if err != nil {
		t.Fatal(err)
	}
	assertFloatSliceClose(t, "gradient differentiate", got, []float64{4, 6, 10, 12}, 0)

	got, err = Integrate([]float64{1, 2}, 2, 3)
	if err != nil {
		t.Fatal(err)
	}
	assertFloatSliceClose(t, "rectangular integrate", got, []float64{3, 3.5, 4.5}, 0)

	got, err = Integrate(nil, 2, 3)
	if err != nil {
		t.Fatal(err)
	}
	assertFloatSliceClose(t, "empty rectangular integrate", got, []float64{3}, 0)

	got, err = CumulativeTrapezoid([]float64{1, 3, 5}, 2)
	if err != nil {
		t.Fatal(err)
	}
	assertFloatSliceClose(t, "cumulative trapezoid", got, []float64{0, 1, 3}, 0)
}

func assertFloatSliceClose(t *testing.T, name string, got, want []float64, tolerance float64) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("%s length = %d, want %d", name, len(got), len(want))
	}
	for i := range want {
		if math.Abs(got[i]-want[i]) > tolerance {
			t.Fatalf("%s[%d] = %.17g, want %.17g (tol %.3g)", name, i, got[i], want[i], tolerance)
		}
	}
}
