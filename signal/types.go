package signal

// Passband identifies the frequency range retained by an IIR filter.
type Passband int

const (
	LowPass Passband = iota
	BandPass
	HighPass
)

const (
	LOW_PASS  = LowPass
	BAND_PASS = BandPass
	HIGH_PASS = HighPass
)

// TaperType identifies a supported time-domain taper.
type TaperType string

const (
	Hanning TaperType = "HANNING"
	Hamming TaperType = "HAMMING"
	Cosine  TaperType = "COSINE"
)

const (
	HANNING = Hanning
	HAMMING = Hamming
	COSINE  = Cosine
)

// ResponseOutput is the requested physical unit after response removal.
type ResponseOutput string

const (
	Displacement ResponseOutput = "DISP"
	Velocity     ResponseOutput = "VEL"
	Acceleration ResponseOutput = "ACC"
)

// LineFitResult contains a least-squares line fit in sample-time coordinates.
type LineFitResult struct {
	Slope       float64
	Intercept   float64
	Sigma       float64
	SigmaSlope  float64
	SigmaOffset float64
	Correlation float64
}
