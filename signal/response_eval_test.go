package signal

import (
	"math"
	"math/cmplx"
	"testing"

	"github.com/bclswl0827/obsgo/inventory"
)

func TestEvalResponseMultiStage(t *testing.T) {
	response := &inventory.Response{Stages: []inventory.Stage{
		{
			Number: 2,
			Coefficients: &inventory.Coefficients{
				Input:        inventory.Units{Name: "V"},
				Numerators:   []float64{1, 1},
				Denominators: []float64{1, -0.5},
			},
			Decimation: &inventory.Decimation{InputSampleRate: 100, Factor: 2},
			Gain:       &inventory.StageGain{Value: 3},
		},
		{
			Number: 1,
			PolesZeros: &inventory.PolesZeros{
				Input:                inventory.Units{Name: "m/s"},
				TransferFunctionType: "LAPLACE (RADIANS/SECOND)",
				NormalizationFactor:  1,
			},
			Gain: &inventory.StageGain{Value: 2},
		},
	}}
	values, err := EvalResponse(response, []float64{0}, 50, "m/s")
	if err != nil {
		t.Fatal(err)
	}
	if got, want := real(values[0]), 24.0; math.Abs(got-want) > 1e-12 {
		t.Fatalf("response at DC = %.15g, want %.15g", got, want)
	}
}

func TestEvalResponseUnits(t *testing.T) {
	response := &inventory.Response{
		Sensitivity: &inventory.InstrumentSensitivity{Value: 2, Input: inventory.Units{Name: "cm/s"}},
	}
	values, err := EvalResponse(response, []float64{1}, 100, "m/s")
	if err != nil {
		t.Fatal(err)
	}
	if got, want := real(values[0]), 200.0; math.Abs(got-want) > 1e-12 {
		t.Fatalf("unit-converted response = %.15g, want %.15g", got, want)
	}
}

func TestEvalResponseFIRSymmetry(t *testing.T) {
	for _, test := range []struct {
		name      string
		symmetry  string
		coeffs    []float64
		frequency float64
		want      complex128
	}{
		{name: "odd", symmetry: "ODD", coeffs: []float64{0.25, 0.5}, want: 1},
		{name: "even", symmetry: "EVEN", coeffs: []float64{0.25, 0.25}, want: 1},
		{name: "none", symmetry: "NONE", coeffs: []float64{1, 2}, frequency: 25, want: complex(1, -2)},
	} {
		t.Run(test.name, func(t *testing.T) {
			response := &inventory.Response{Stages: []inventory.Stage{{
				Number: 1,
				FIR: &inventory.FIR{
					Symmetry:   test.symmetry,
					Input:      inventory.Units{Name: "m/s"},
					Numerators: test.coeffs,
				},
			}}}
			values, err := EvalResponse(response, []float64{test.frequency}, 100, "m/s")
			if err != nil {
				t.Fatal(err)
			}
			if cmplx.Abs(values[0]-test.want) > 1e-12 {
				t.Fatalf("FIR response = %v, want %v", values[0], test.want)
			}
		})
	}
}

func TestEvalResponseAsymmetricFIRCorrection(t *testing.T) {
	response := &inventory.Response{Stages: []inventory.Stage{{
		Number:     1,
		FIR:        &inventory.FIR{Symmetry: "NONE", Input: inventory.Units{Name: "m/s"}, Numerators: []float64{1}},
		Decimation: &inventory.Decimation{InputSampleRate: 100, Factor: 1, Correction: 0.01},
	}}}
	values, err := EvalResponse(response, []float64{10}, 100, "m/s")
	if err != nil {
		t.Fatal(err)
	}
	want := cmplx.Exp(complex(0, 2*math.Pi*10*0.01))
	if cmplx.Abs(values[0]-want) > 1e-12 {
		t.Fatalf("FIR correction response = %v, want %v", values[0], want)
	}
}

func TestSimulateResponse(t *testing.T) {
	response := &inventory.Response{Gain: 2}
	output, err := SimulateResponse([]float64{1, 0}, 100, response, Velocity, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(output) != 2 || math.Abs(output[0]-2) > 1e-12 || math.Abs(output[1]) > 1e-12 {
		t.Fatalf("simulated response = %v, want [2 0]", output)
	}
}
