package trigger

import (
	"math"
	"testing"
)

// Golden values come from ObsPy 1.4.2's clibsignal AR picker and its exported
// spr_bp_fast_bworth and spr_coef_paz primitives.

func TestObsPyARPickerPrimitiveGolden(t *testing.T) {
	filtered := []float32{0.2, -1, 0.5, 2, -0.75, 0.1, 1.2, -0.4, 0, 0.3, -0.2, 0.8}
	arBandpassFloat32(filtered, 0.05, 1, 4, 2, true)
	wantFiltered := []float32{
		-0.4096003472805023, -0.245689257979393, 0.12144982069730759,
		0.39121589064598083, 0.3495773375034332, 0.07294230163097382,
		-0.1773701310157776, -0.20024237036705017, -0.011202944442629814,
		0.19056804478168488, 0.24289199709892273, 0.18667402863502502,
	}
	assertFloat32Golden(t, "AR Butterworth", filtered, wantFiltered, 0)

	coefficients, err := arCoefficientsPAZ([]float32{0.2, -1, 0.5, 2, -0.75, 0.1, 1.2, -0.4}, 3)
	if err != nil {
		t.Fatal(err)
	}
	assertFloat32Golden(t, "AR coefficients", coefficients, []float32{-0.013308972120285034, -0.3378715515136719, 0.6048113107681274}, 0)
}

func TestObsPyARPickGolden(t *testing.T) {
	tests := []struct {
		length     int
		sampleRate float64
		pOnset     float64
		sOnset     float64
		wantP      float64
		wantS      float64
	}{
		{600, 100, 1.4, 3.2, 1.2799999713897705, 3.0199999809265137},
		{800, 80, 2, 5, 1.837499976158142, 4.824999809265137},
	}
	for _, test := range tests {
		vertical, north, east := syntheticARPickData(test.length, test.sampleRate, test.pOnset, test.sOnset)
		pTime, sTime, err := ARPick(vertical, north, east, test.sampleRate, 1, 12, 0.8, 0.1, 1, 0.2, 2, 2, 0.2, 0.3, true)
		if err != nil {
			t.Fatal(err)
		}
		if pTime != test.wantP || sTime != test.wantS {
			t.Fatalf("AR pick = (%g, %g), want (%g, %g)", pTime, sTime, test.wantP, test.wantS)
		}
	}
}

func syntheticARPickData(length int, sampleRate, pOnset, sOnset float64) ([]float64, []float64, []float64) {
	vertical := make([]float64, length)
	north := make([]float64, length)
	east := make([]float64, length)
	pIndex := int(pOnset * sampleRate)
	sIndex := int(sOnset * sampleRate)
	for index := 0; index < length; index++ {
		time := float64(index) / sampleRate
		vertical[index] = 0.02*math.Sin(2*math.Pi*0.7*time) + 0.01*math.Cos(2*math.Pi*1.3*time)
		north[index] = 0.015 * math.Sin(2*math.Pi*0.9*time+0.3)
		east[index] = 0.012 * math.Cos(2*math.Pi*1.1*time-0.2)
		if index >= pIndex {
			elapsed := time - pOnset
			vertical[index] += 0.8 * math.Exp(-2*elapsed) * math.Sin(2*math.Pi*6*elapsed)
		}
		if index >= sIndex {
			elapsed := time - sOnset
			north[index] += 0.7 * math.Exp(-1.5*elapsed) * math.Sin(2*math.Pi*4*elapsed)
			east[index] += 0.5 * math.Exp(-1.2*elapsed) * math.Cos(2*math.Pi*4.5*elapsed)
		}
	}
	return vertical, north, east
}

func assertFloat32Golden(t *testing.T, name string, got, want []float32, tolerance float64) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("%s length = %d, want %d", name, len(got), len(want))
	}
	for index := range got {
		if math.Abs(float64(got[index]-want[index])) > tolerance {
			t.Fatalf("%s[%d] = %.9g, want %.9g", name, index, got[index], want[index])
		}
	}
}
