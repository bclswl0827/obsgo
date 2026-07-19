package eewgo

import (
	"math"
	"reflect"
	"testing"
)

var triggerTestData = []float64{
	0.2, -0.1, 0.4, 1.2, -0.7, 0.3, 2.1, -1.4,
	0.8, 0.05, -0.9, 1.7, -0.2, 0.6, -1.1, 0.15,
}

func assertFloatSlicesClose(t *testing.T, got, want []float64) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("length = %d, want %d", len(got), len(want))
	}
	for i := range want {
		tolerance := 1e-14 * math.Max(1, math.Abs(want[i]))
		if math.IsNaN(want[i]) {
			if !math.IsNaN(got[i]) {
				t.Errorf("result[%d] = %v, want NaN", i, got[i])
			}
			continue
		}
		if math.IsNaN(got[i]) || math.IsInf(got[i], 0) || math.Abs(got[i]-want[i]) > tolerance {
			t.Errorf("result[%d] = %.17g, want %.17g", i, got[i], want[i])
		}
	}
}

func TestClassicStaLtaMatchesObsPy(t *testing.T) {
	want := []float64{
		0, 0, 0, 0, 0, 0, 0, 0, 1.4785714285714286,
		1.3892420537897312, 1.4076980754811301, 0.8909876300805027,
		0.6960953011250827, 0.6591832180316891, 0.7756542909312234,
		1.0258979206049148,
	}
	assertFloatSlicesClose(t, ClassicStaLta(triggerTestData, 5, 9), want)

	// This window ordering is intentional: it catches using nsta instead of
	// nlta to form the LTA subtraction slice.
	want = []float64{
		0, 0, 0, 0, 0.5555555555555556, 0.5657026889903602,
		0.5597706963412578, 0.5694609985432394, 0.6763285024154589,
		0.7198169658570925, 0.7103796029970525, 1.1223500374630877,
		1.4365849020726373, 1.5170289119100824, 1.289234149403641,
		0.974755850377741,
	}
	assertFloatSlicesClose(t, ClassicStaLta(triggerTestData, 9, 5), want)

	// NumPy clamps the zero-padding slice when nlta is longer than the data.
	assertFloatSlicesClose(t, ClassicStaLta(triggerTestData, 5, 20), make([]float64, len(triggerTestData)))
}

func TestRecursiveStaLtaMatchesObsPyAndPreservesInput(t *testing.T) {
	input := append([]float64(nil), triggerTestData...)
	wantInput := append([]float64(nil), input...)
	want := []float64{
		0, 0, 0, 0, 0, 0, 0, 1.8049377454060167,
		1.4937788819689308, 1.1623443866678886, 1.1081721998905347,
		1.410512855845125, 1.1044638940696272, 0.9461904784859971,
		1.036560075778958, 0.8124230854096155,
	}
	assertFloatSlicesClose(t, RecursiveStaLta(input, 3, 7), want)
	if !reflect.DeepEqual(input, wantInput) {
		t.Fatalf("RecursiveStaLta modified its input: got %v, want %v", input, wantInput)
	}
}

func TestRecursiveStaLtaUsesNumPyTiny(t *testing.T) {
	input := []float64{1e-154, 1e-154, 1e-154, 1e-154, 1e-154, 1e-154}
	want := []float64{0, 0, 0.7655394178758645, 0.8672017289728844, 0.9288784554092256, 0.9631280320481961}
	assertFloatSlicesClose(t, RecursiveStaLta(input, 1, 2), want)
}

func TestDelayedStaLtaMatchesObsPy(t *testing.T) {
	input := make([]float64, 72)
	for i := range input {
		input[i] = float64(i%11-5)*0.17 + float64(i*i%7)*0.03
	}
	got := DelayedStaLta(input, 4, 7)
	assertFloatSlicesClose(t, got[:61], make([]float64, 61))
	wantTail := []float64{
		1.7603994816671136, 1.7561016886743819, 1.7566573787177564,
		1.7649152712323266, 1.7874348493992422, 1.77408415379279,
		1.7660427597986847, 1.7627714632870892, 1.775854629352401,
		1.7607789201317792, 1.7499999999999996,
	}
	assertFloatSlicesClose(t, got[61:], wantTail)

	// The initial muted NumPy slice extends past this short input and is
	// clamped to its length.
	assertFloatSlicesClose(t, DelayedStaLta(make([]float64, 12), 2, 3), make([]float64, 12))
}

func TestZDetectMatchesObsPy(t *testing.T) {
	want := []float64{
		-1.3141552918056203, -1.3141552918056203, -1.3141552918056203,
		-1.3141552918056203, -0.6659172554889466, -0.4891250637662175,
		-0.45769534079328783, 1.212008692143599, 1.4163018914676415,
		1.4752326220418852, 1.440856362540243, 0.026518828758409687,
		0.39188935831871624, 0.15616643602174343, 0.2966180105570227,
		0.453766625421671,
	}
	assertFloatSlicesClose(t, ZDetect(triggerTestData, 4), want)
}

func TestCarlStaTrigMatchesObsPy(t *testing.T) {
	want := []float64{
		-1, -1, -1, -1, -1, -1, 0.057777777777777706,
		0.12111111111111111, 0.5481481481481483, 0.3980246913580246,
		-0.7155555555555555, -0.7376543209876545, -0.08339506172839506,
		-0.05796296296296308, -0.7637037037037036, -0.6300617283950616,
	}
	assertFloatSlicesClose(t, CarlStaTrig(triggerTestData, 3, 6, 0.8, 0.2), want)
}

func TestTriggerOnsetMatchesObsPy(t *testing.T) {
	oneWave := make([]float64, 0, 158)
	for i := 0; float64(i)*0.1 < 5*math.Pi; i++ {
		oneWave = append(oneWave, math.Sin(float64(i)*0.1))
	}
	charfct := make([]float64, 0, 4*len(oneWave))
	for _, offset := range []float64{1, 2.1, 0.4, 1} {
		for _, value := range oneWave {
			charfct = append(charfct, value+offset)
		}
	}

	want := [][]int{{6, 31}, {69, 94}, {131, 181}, {215, 265}, {278, 315}, {480, 505}, {543, 568}, {605, 631}}
	if got := TriggerOnset(charfct, 1.5, 1, 50, false); !reflect.DeepEqual(got, want) {
		t.Fatalf("TriggerOnset() = %v, want %v", got, want)
	}
	wantDelete := [][]int{{6, 31}, {69, 94}, {480, 505}, {543, 568}, {605, 631}}
	if got := TriggerOnset(charfct, 1.5, 1, 50, true); !reflect.DeepEqual(got, wantDelete) {
		t.Fatalf("TriggerOnset(maxLenDelete=true) = %v, want %v", got, wantDelete)
	}
}
