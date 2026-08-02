package signal

import (
	"math"
	"testing"
)

// Direction and PM golden values were generated with ObsPy 1.4.2 and SciPy
// 1.17.0. Ellipticity is the covariance eigenvalue-axis ratio exposed by this
// fixed-window API.

func TestObsPyPolarizationGolden(t *testing.T) {
	z := []float64{1, 0.8, 0.2, -0.4, -0.9, -0.5, 0.1, 0.7, 1.1, 0.6}
	n := []float64{0.2, 0.4, 0.8, 1, 0.5, -0.2, -0.8, -0.9, -0.3, 0.4}
	e := []float64{-0.1, 0.3, 0.7, 0.6, 0.1, -0.5, -0.7, -0.4, 0.2, 0.8}
	result, err := AnalyzePolarization(z, n, e)
	if err != nil {
		t.Fatal(err)
	}
	assertClose := func(name string, got, want, tolerance float64) {
		t.Helper()
		if math.Abs(got-want) > tolerance {
			t.Fatalf("%s = %.17g, want %.17g", name, got, want)
		}
	}
	assertClose("Flinn azimuth", result.Azimuth, 33.201425010140724, 1e-10)
	assertClose("Flinn incidence", result.Incidence, 64.78097713207988, 1e-10)
	assertClose("Flinn rectilinearity", result.Rectilinearity, 0.17111805585673145, 1e-12)
	assertClose("Flinn planarity", result.Planarity, 0.9432643602190682, 1e-12)
	assertClose("polarization ellipticity", result.Ellipticity, 0.8288819441432683, 1e-12)

	pm, err := analyzeParticleMotion(z, n, e, 0)
	if err != nil {
		t.Fatal(err)
	}
	assertClose("PM azimuth", pm.Azimuth, 38.008296027997574, 5e-4)
	assertClose("PM incidence", pm.Incidence, 50.27374088001654, 6e-4)
	assertClose("PM azimuth error", pm.AzimuthError, 0.14916711233261354, 3e-6)
	assertClose("PM incidence error", pm.IncidenceError, 0.1674896456070257, 3e-6)
}
