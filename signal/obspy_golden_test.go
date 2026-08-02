package signal

import "testing"

// Golden values in this file were generated with ObsPy 1.4.2, SciPy 1.17.0,
// and NumPy 2.4.1. Windowed Lanczos values come directly from ObsPy's C
// lanczos_resample implementation because ObsPy 1.4.2's Python wrapper always
// selects the default Lanczos window.

func TestObsPyFilterGolden(t *testing.T) {
	input := []float64{0.2, -1, 0.5, 2, -0.75, 0.1, 1.2, -0.4, 0, 0.3, -0.2, 0.8}
	lowpass, err := ObsPyLowpass(input, 3, 20, 4, false)
	if err != nil {
		t.Fatal(err)
	}
	assertFloatSliceClose(t, "ObsPy low-pass", lowpass, []float64{
		0.0037126021253794344, 0.0021176639873836996, -0.044105192090284115,
		-0.09244127916937872, 0.03156582484561585, 0.3334506065049757,
		0.5459193412887228, 0.5285434580501585, 0.39021723279250253,
		0.22959325629188512, 0.08835527058919519, 0.020618095183320563,
	}, 1e-12)

	highpass, err := ObsPyHighpass(input, 3, 20, 4, false)
	if err != nil {
		t.Fatal(err)
	}
	assertFloatSliceClose(t, "ObsPy high-pass", highpass, []float64{
		0.055082657614460864, -0.4092421762898037, 0.8569181264640119,
		0.021609880609735288, -1.6884860708452862, 1.0996154274421848,
		0.7412345617658146, -0.948039821659234, 0.333897422134355,
		0.13614304905387212, -0.342260090435831, 0.34906801140268234,
	}, 1e-12)

	bandpass, err := ObsPyBandpass(input, 1, 4, 20, 2, true)
	if err != nil {
		t.Fatal(err)
	}
	assertFloatSliceClose(t, "ObsPy zero-phase band-pass", bandpass, []float64{
		-0.3534849989837794, -0.19560377782224894, 0.22035606408917904,
		0.435828257797809, 0.2777880330430232, 0.08041766343949269,
		0.0030897750028784365, -0.07269542466710022, -0.12687562680916067,
		-0.08776880049038949, -0.022489111350040815, 0.00031535298466300903,
	}, 1e-12)
}

func TestObsPyHilbertGolden(t *testing.T) {
	input := []float64{1, -0.5, 0.25, 2, -1, 0.125}
	assertFloatSliceClose(t, "ObsPy Hilbert", ObsPyHilbert(input), []float64{
		0.3608439182435162, 0.43301270189221924, -1.443375672974064,
		0.721687836487032, 1.0825317547305482, -1.1547005383792515,
	}, 1e-12)
	assertFloatSliceClose(t, "ObsPy envelope", ObsPyEnvelope(input), []float64{
		1.063112568514423, 0.6614378277661476, 1.4648663192705784,
		2.126225137028845, 1.4737282653189494, 1.161446655397196,
	}, 1e-12)
}

func TestObsPyInterpolationGolden(t *testing.T) {
	input := []float64{0, 1, 0, 2, -1, 0.5}
	tests := []struct {
		name   string
		method InterpolationMethod
		want   []float64
	}{
		{"linear", LinearInterpolation, []float64{0.5, 1, 0.5, 0, 1, 2, 0.5, -1, -0.25, 0.5}},
		{"nearest", NearestInterpolation, []float64{0, 1, 1, 0, 0, 2, 2, -1, -1, 0.5}},
		{"zero", ZeroInterpolation, []float64{0, 1, 1, 0, 0, 2, 2, -1, -1, 0.5}},
		{"quadratic", QuadraticInterpolation, []float64{0.8924632352941176, 1, 0.32261029411764697, 0, 1.171875, 2, 0.6461397058823533, -1, -1.0487132352941178, 0.5}},
		{"cubic", CubicInterpolation, []float64{1.3520833333333326, 1, 0.14791666666666736, 0, 1.1812499999999997, 2, 0.8770833333333319, -1, -1.7520833333333334, 0.5}},
		{"weighted-average-slopes", WeightedAverageSlopesInterpolation, []float64{0.625, 1, 0.5, 0, 1, 2, 0.5, -1, -0.4375, 0.5}},
	}
	for _, test := range tests {
		got, err := Interpolate1D(input, 2, 0.5, 2.25, 0.25, 10, test.method)
		if err != nil {
			t.Fatalf("%s interpolation: %v", test.name, err)
		}
		assertFloatSliceClose(t, "ObsPy "+test.name+" interpolation", got, test.want, 2e-12)
	}
}

func TestObsPyLanczosGolden(t *testing.T) {
	tests := []struct {
		name   string
		window LanczosKernelWindow
		want   []float64
	}{
		{"lanczos", LanczosWindow, []float64{
			0.6565612700023488, 1, 0.31342019473363153, 1.1283167545583065e-16,
			1.2280127457451337, 2, 0.5646967301666292, -0.9999999999999999,
			-0.5741533739732474, 0.4999999999999999,
		}},
		{"hanning", HanningLanczosWindow, []float64{
			0.6110325092838447, 1, 0.37323865541091406, 9.74542958129844e-17,
			1.192213211636663, 2, 0.5494517738919673, -0.9999999999999999,
			-0.5091937577365372, 0.49999999999999994,
		}},
		{"blackman", BlackmanLanczosWindow, []float64{
			0.57538180203126, 0.9999999999999999, 0.42077293174769337,
			7.87430710168914e-17, 1.1387371507577861, 1.9999999999999998,
			0.5358705520814588, -0.9999999999999999, -0.42855525323664356,
			0.49999999999999994,
		}},
	}
	for _, test := range tests {
		got, err := Interpolate1DWithOptions(
			[]float64{0, 1, 0, 2, -1, 0.5}, 2, 0.5, 2.25, 0.25, 10,
			LanczosInterpolation, InterpolationOptions{LanczosWidth: 3, LanczosWindow: test.window},
		)
		if err != nil {
			t.Fatalf("%s Lanczos interpolation: %v", test.name, err)
		}
		assertFloatSliceClose(t, "ObsPy "+test.name+" Lanczos", got, test.want, 1e-12)
	}
}

func TestObsPyWelchGolden(t *testing.T) {
	overlap := 0.5
	result, err := Welch(
		[]float64{0.2, -0.1, 0.5, 1.2, -0.7, 0.3, 0.9, -0.4, 0.1, 0.8, -0.2, 0.6},
		20,
		WelchOptions{SegmentLength: 8, OverlapFraction: &overlap, NFFT: 16, Window: HannWindow, Detrend: DetrendConstant},
	)
	if err != nil {
		t.Fatal(err)
	}
	assertFloatSliceClose(t, "ObsPy Welch frequencies", result.Frequencies, []float64{0, 1.25, 2.5, 3.75, 5, 6.25, 7.5, 8.75, 10}, 0)
	assertFloatSliceClose(t, "ObsPy Welch power", result.Power, []float64{
		0.0009661045057899799, 0.0012312309971431745, 0.004001166463512613,
		0.021647115038593342, 0.05846842530880126, 0.09085524246526865,
		0.08419499754331107, 0.046044156179308876, 0.012147050858899104,
	}, 2e-14)

	noOverlap := 0.0
	result, err = Welch(
		[]float64{0.2, -0.1, 0.5, 1.2, -0.7, 0.3, 0.9, -0.4, 0.1, 0.8, -0.2, 0.6},
		20,
		WelchOptions{SegmentLength: 8, OverlapFraction: &noOverlap, NFFT: 16, Window: HannWindow, Detrend: DetrendConstant},
	)
	if err != nil {
		t.Fatal(err)
	}
	assertFloatSliceClose(t, "ObsPy Welch without overlap", result.Power, []float64{
		0.0011017316155351184, 0.0019269485050182082, 0.006251701597789807,
		0.03135486137813008, 0.08281861470757187, 0.12311233637736187,
		0.11151625774509298, 0.06957927242525543, 0.02428511301977579,
	}, 2e-14)

	quarterOverlap := 0.25
	result, err = Welch(
		[]float64{0.2, -0.1, 0.5, 1.2, -0.7, 0.3, 0.9, -0.4, 0.1, 0.8, -0.2, 0.6},
		20,
		WelchOptions{SegmentLength: 7, OverlapFraction: &quarterOverlap, NFFT: 16, Window: HannWindow, Detrend: DetrendConstant},
	)
	if err != nil {
		t.Fatal(err)
	}
	assertFloatSliceClose(t, "ObsPy Welch fractional overlap truncation", result.Power, []float64{
		2.3527509215965235e-05, 0.0027815912855921497, 0.015628261764578058,
		0.04604020319768161, 0.08664265531389938, 0.11241844901437392,
		0.1097685245390314, 0.09341816819218968, 0.04259544256311257,
	}, 2e-14)

	defaultResult, err := Welch(make([]float64, 300), 20, WelchOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if defaultResult.NFFT != 256 || len(defaultResult.Frequencies) != 129 {
		t.Fatalf("default Welch shape = (nfft %d, bins %d), want (256, 129)", defaultResult.NFFT, len(defaultResult.Frequencies))
	}
}

func TestObsPyCorrelationGolden(t *testing.T) {
	a := []float64{0.1, 1.2, -0.3, 0.5, 2, -1, 0.25}
	b := []float64{0.2, -0.4, 1.1, 0.3, -0.7}
	want := []float64{0.12636426351850397, 0.2179356639398523, -0.7142142326569499, 0.5887037817297871, 0.3944614171996544, -0.6617047582894634, 0.2134531478353108}
	for _, method := range []CorrelationMethod{CorrelationDirect, CorrelationFFT} {
		options := DefaultCorrelationOptions()
		options.Method = method
		got, err := Correlate(a, b, 3, options)
		if err != nil {
			t.Fatal(err)
		}
		assertFloatSliceClose(t, "ObsPy correlation "+string(method), got, want, 1e-14)
	}
}

func TestObsPyTemplateCorrelationGolden(t *testing.T) {
	data := []float64{0.1, 0.2, 0.5, 1, 0.7, -0.2, -0.5, 0.3, 0.9, 0.1}
	template := []float64{0.5, 1, 0.7, -0.2}
	tests := []struct {
		mode TemplateCorrelationMode
		want []float64
	}{
		{TemplateValid, []float64{-0.8087693101032828, -0.0776735637380617, 1, 0.6041585400191987, -0.5042741452666882, -0.8743962285879124, 0.2943920288775949}},
		{TemplateSame, []float64{-0.8193464903987026, -0.7867957924694433, -0.8087693101032828, -0.0776735637380617, 1, 0.6041585400191987, -0.5042741452666882, -0.8743962285879124, 0.2943920288775949, 0.7621900001431481}},
		{TemplateFull, []float64{-0.9152086306448587, -0.8193464903987026, -0.7867957924694433, -0.8087693101032828, -0.0776735637380617, 1, 0.6041585400191987, -0.5042741452666882, -0.8743962285879124, 0.2943920288775949, 0.7621900001431481, 0.07498687844430794, 0}},
	}
	for _, test := range tests {
		options := DefaultTemplateCorrelationOptions()
		options.Mode = test.mode
		options.Method = CorrelationDirect
		got, err := CorrelateTemplate(data, template, options)
		if err != nil {
			t.Fatal(err)
		}
		assertFloatSliceClose(t, "ObsPy template correlation "+string(test.mode), got, test.want, 2e-14)
	}
}
