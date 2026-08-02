package trigger

import (
	"math"
	"testing"
)

func TestAICSimpleHasExpectedShape(t *testing.T) {
	data := []float64{0, 0, 0, 1, 1, 1, 1, 1}
	values := AICSimple(data)
	if len(values) != len(data) || values[len(values)-1] != values[len(values)-2] {
		t.Fatalf("AIC output boundary is invalid: %v", values)
	}
}

func TestEnergyRatioAndKurtosis(t *testing.T) {
	data := make([]float64, 32)
	for index := 16; index < len(data); index++ {
		data[index] = 3
	}
	ratio, err := EnergyRatio(data, 2, 8)
	if err != nil {
		t.Fatal(err)
	}
	maximum := 0.0
	for _, value := range ratio {
		if value > maximum {
			maximum = value
		}
	}
	if maximum <= 1 {
		t.Fatalf("energy ratio did not detect onset: max=%g", maximum)
	}
	kurtosis, err := KurtosisCharacteristic(data, 8)
	if err != nil {
		t.Fatal(err)
	}
	if math.IsNaN(kurtosis[len(kurtosis)-1]) {
		t.Fatal("kurtosis returned NaN")
	}
}

func TestARPickSyntheticObsPyRegression(t *testing.T) {
	vertical := make([]float64, 200)
	north := make([]float64, 200)
	east := make([]float64, 200)
	for index := 80; index < len(vertical); index++ {
		vertical[index] = 3 * math.Sin(float64(index-80)*0.3)
	}
	for index := 130; index < len(north); index++ {
		north[index] = 2 * math.Sin(float64(index-130)*0.25)
		east[index] = 1 * math.Sin(float64(index-130)*0.25)
	}
	pickP, pickS, err := ARPick(vertical, north, east, 20, 0.5, 5, 1, 0.1, 1, 0.1, 2, 2, 3, 3, true)
	if err != nil {
		t.Fatal(err)
	}
	if pickP != 3.0999999046325684 || pickS != 0 {
		t.Fatalf("AR picks = (%g,%g), want ObsPy result (3.0999999046325684,0)", pickP, pickS)
	}
}
