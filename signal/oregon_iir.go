package signal

import (
	"fmt"
	"math"
	"math/cmplx"
)

// oregonPolynomial stores coefficients in ascending order, matching
// OregonDSP's Polynomial class: a[0] + a[1]s + a[2]s^2 + ... .
type oregonPolynomial struct {
	coefficients []float64
}

func newOregonPolynomial(coefficients ...float64) oregonPolynomial {
	return oregonPolynomial{coefficients: append([]float64(nil), coefficients...)}
}

func (polynomial oregonPolynomial) order() int {
	return len(polynomial.coefficients) - 1
}

func (polynomial oregonPolynomial) multiply(other oregonPolynomial) oregonPolynomial {
	coefficients := make([]float64, polynomial.order()+other.order()+1)
	for i, left := range polynomial.coefficients {
		for j, right := range other.coefficients {
			coefficients[i+j] += left * right
		}
	}
	return oregonPolynomial{coefficients: coefficients}
}

func (polynomial oregonPolynomial) add(other oregonPolynomial) oregonPolynomial {
	order := polynomial.order()
	if other.order() > order {
		order = other.order()
	}
	coefficients := make([]float64, order+1)
	for i, value := range polynomial.coefficients {
		coefficients[i] += value
	}
	for i, value := range other.coefficients {
		coefficients[i] += value
	}
	return oregonPolynomial{coefficients: coefficients}
}

func (polynomial oregonPolynomial) scale(value float64) oregonPolynomial {
	coefficients := make([]float64, len(polynomial.coefficients))
	for i, coefficient := range polynomial.coefficients {
		coefficients[i] = value * coefficient
	}
	return oregonPolynomial{coefficients: coefficients}
}

func (polynomial oregonPolynomial) trim() oregonPolynomial {
	coefficients := append([]float64(nil), polynomial.coefficients...)
	for len(coefficients) > 1 && coefficients[len(coefficients)-1] == 0 {
		coefficients = coefficients[:len(coefficients)-1]
	}
	return oregonPolynomial{coefficients: coefficients}
}

func (polynomial oregonPolynomial) evaluate(value float64) float64 {
	result := 0.0
	for i := len(polynomial.coefficients) - 1; i >= 0; i-- {
		result = result*value + polynomial.coefficients[i]
	}
	return result
}

type oregonRational struct {
	numerator   oregonPolynomial
	denominator oregonPolynomial
}

func newOregonRational(numerator, denominator oregonPolynomial) oregonRational {
	return oregonRational{numerator: numerator, denominator: denominator}
}

func (rational oregonRational) multiply(other oregonRational) oregonRational {
	return newOregonRational(
		rational.numerator.multiply(other.numerator),
		rational.denominator.multiply(other.denominator),
	)
}

func (rational *oregonRational) scaleNumerator(value float64) {
	rational.numerator = rational.numerator.scale(value)
}

// canonicalForm matches Rational.canonicalForm in OregonDSP. It normalizes
// the highest coefficient of each polynomial and returns the gain removed.
func (rational *oregonRational) canonicalForm() float64 {
	numeratorScale := rational.numerator.coefficients[rational.numerator.order()]
	denominatorScale := rational.denominator.coefficients[rational.denominator.order()]
	for i := range rational.numerator.coefficients {
		rational.numerator.coefficients[i] /= numeratorScale
	}
	for i := range rational.denominator.coefficients {
		rational.denominator.coefficients[i] /= denominatorScale
	}
	return numeratorScale / denominatorScale
}

// map substitutes a rational function for the independent variable, using
// the same modified Horner scheme as OregonDSP's Rational.map().
func (rational oregonRational) mapTo(substitution oregonRational) oregonRational {
	numerator := newOregonPolynomial(rational.numerator.coefficients[rational.numerator.order()])
	numeratorPower := newOregonPolynomial(1)
	for i := rational.numerator.order() - 1; i >= 0; i-- {
		numeratorPower = numeratorPower.multiply(substitution.denominator)
		numerator = numerator.multiply(substitution.numerator).add(
			numeratorPower.scale(rational.numerator.coefficients[i]),
		)
	}

	denominator := newOregonPolynomial(rational.denominator.coefficients[rational.denominator.order()])
	denominatorPower := newOregonPolynomial(1)
	for i := rational.denominator.order() - 1; i >= 0; i-- {
		denominatorPower = denominatorPower.multiply(substitution.denominator)
		denominator = denominator.multiply(substitution.numerator).add(
			denominatorPower.scale(rational.denominator.coefficients[i]),
		)
	}

	if rational.denominator.order() > rational.numerator.order() {
		for i := 0; i < rational.denominator.order()-rational.numerator.order(); i++ {
			numerator = numerator.multiply(substitution.denominator)
		}
	} else if rational.numerator.order() > rational.denominator.order() {
		for i := 0; i < rational.numerator.order()-rational.denominator.order(); i++ {
			denominator = denominator.multiply(substitution.denominator)
		}
	}

	return newOregonRational(numerator.trim(), denominator.trim())
}

type oregonPrototype struct {
	sections []oregonRational
}

func (prototype *oregonPrototype) add(section oregonRational) {
	prototype.sections = append(prototype.sections, section)
}

func newOregonButterworth(order int) oregonPrototype {
	prototype := oregonPrototype{}
	nRealPoles := order - 2*(order/2)
	nComplexPolePairs := order / 2
	nPoles := nRealPoles + 2*nComplexPolePairs
	if nRealPoles == 1 {
		prototype.add(newOregonRational(
			newOregonPolynomial(1),
			newOregonPolynomial(1, 1),
		))
	}

	deltaAngle := math.Pi / float64(nPoles)
	for i := 0; i < nComplexPolePairs; i++ {
		angle := -math.Pi/2 + deltaAngle/2*float64(1+nRealPoles) + float64(i)*deltaAngle
		prototype.add(newOregonRational(
			newOregonPolynomial(1),
			newOregonPolynomial(1, -2*math.Sin(angle), 1),
		))
	}
	return prototype
}

func newOregonChebyshevI(order int, epsilon float64) oregonPrototype {
	prototype := oregonPrototype{}
	alpha := (1 + math.Sqrt(1+epsilon*epsilon)) / epsilon
	p := math.Pow(alpha, 1/float64(order))
	a := 0.5 * (p - 1/p)
	b := 0.5 * (p + 1/p)

	nRealPoles := order - 2*(order/2)
	nComplexPolePairs := order / 2
	nPoles := nRealPoles + 2*nComplexPolePairs
	if nRealPoles == 1 {
		prototype.add(newOregonRational(
			newOregonPolynomial(1),
			newOregonPolynomial(a, 1),
		))
	}

	deltaAngle := math.Pi / float64(nPoles)
	for i := 0; i < nComplexPolePairs; i++ {
		angle := -math.Pi/2 + deltaAngle/2*float64(1+nRealPoles) + float64(i)*deltaAngle
		pole := complex(a*math.Sin(angle), b*math.Cos(angle))
		prototype.add(newOregonRational(
			newOregonPolynomial(1),
			newOregonPolynomial(real(pole*cmplx.Conj(pole)), -2*real(pole), 1),
		))
	}

	prototype.sections[0].scaleNumerator(1 / (math.Pow(2, float64(order-1)) * epsilon))
	return prototype
}

func newOregonChebyshevII(order int, epsilon float64) oregonPrototype {
	prototype := oregonPrototype{}
	alpha := (1 + math.Sqrt(1+epsilon*epsilon)) / epsilon
	p := math.Pow(alpha, 1/float64(order))
	a := 0.5 * (p - 1/p)
	b := 0.5 * (p + 1/p)

	nRealPoles := order - 2*(order/2)
	nComplexPolePairs := order / 2
	nPoles := nRealPoles + 2*nComplexPolePairs
	if nRealPoles == 1 {
		prototype.add(newOregonRational(
			newOregonPolynomial(1),
			newOregonPolynomial(1/a, 1),
		))
	}

	deltaAngle := math.Pi / float64(nPoles)
	for i := 0; i < nComplexPolePairs; i++ {
		angle := -math.Pi/2 + deltaAngle/2*float64(1+nRealPoles) + float64(i)*deltaAngle
		base := complex(a*math.Sin(angle), b*math.Cos(angle))
		pole := 1 / base
		zeroImaginary := 1 / math.Cos(float64(2*i+1)*math.Pi/(2*float64(order)))
		prototype.add(newOregonRational(
			newOregonPolynomial(zeroImaginary*zeroImaginary, 0, 1),
			newOregonPolynomial(real(pole*cmplx.Conj(pole)), -2*real(pole), 1),
		))
	}

	dcValue := oregonPrototypeValue(prototype, 0)
	prototype.sections[0].scaleNumerator(1 / dcValue)
	return prototype
}

func oregonPrototypeValue(prototype oregonPrototype, value float64) float64 {
	result := 1.0
	for _, section := range prototype.sections {
		result *= section.numerator.evaluate(value) / section.denominator.evaluate(value)
	}
	return result
}

func (prototype oregonPrototype) lowpass(cutoff float64) oregonPrototype {
	mapping := newOregonRational(newOregonPolynomial(0, 1), newOregonPolynomial(cutoff))
	return prototype.mapSections(mapping)
}

func (prototype oregonPrototype) highpass(cutoff float64) oregonPrototype {
	mapping := newOregonRational(newOregonPolynomial(cutoff), newOregonPolynomial(0, 1))
	return prototype.mapSections(mapping)
}

func (prototype oregonPrototype) bandpass(low, high float64) oregonPrototype {
	bandwidth := high - low
	product := low * high
	mapping := newOregonRational(
		newOregonPolynomial(product, 0, 1),
		newOregonPolynomial(0, bandwidth),
	)
	transformed := oregonPrototype{}
	gain := 1.0

	for _, section := range prototype.sections {
		mapped := section.mapTo(mapping)
		gain *= mapped.canonicalForm()
		orders := [2]int{section.numerator.order(), section.denominator.order()}
		if orders[0] < 2 && orders[1] < 2 {
			transformed.add(mapped)
			continue
		}
		if orders[1] != 2 {
			return oregonPrototype{}
		}

		factors := oregonBandpassFactors(section.denominator, bandwidth, product)
		t1 := newOregonPolynomial(0, 1)
		switch orders[0] {
		case 0:
			transformed.add(newOregonRational(t1, factors[0]))
			transformed.add(newOregonRational(t1, factors[1]))
		case 1:
			transformed.add(newOregonRational(t1, factors[0]))
			coefficients := mapped.numerator.coefficients
			transformed.add(newOregonRational(
				newOregonPolynomial(coefficients[1], coefficients[2], coefficients[3]),
				factors[1],
			))
		case 2:
			numeratorFactors := oregonBandpassFactors(section.numerator, bandwidth, product)
			transformed.add(newOregonRational(numeratorFactors[0], factors[0]))
			transformed.add(newOregonRational(numeratorFactors[1], factors[1]))
		}
	}
	if len(transformed.sections) > 0 {
		transformed.sections[0].scaleNumerator(gain)
	}
	return transformed
}

func (prototype oregonPrototype) mapSections(mapping oregonRational) oregonPrototype {
	transformed := oregonPrototype{sections: make([]oregonRational, 0, len(prototype.sections))}
	for _, section := range prototype.sections {
		transformed.add(section.mapTo(mapping))
	}
	return transformed
}

func oregonBandpassFactors(polynomial oregonPolynomial, bandwidth, product float64) [2]oregonPolynomial {
	coefficients := polynomial.coefficients
	c := coefficients[0] / coefficients[2]
	b := coefficients[1] / coefficients[2]
	discriminant := b*b - 4*c
	var factors [2]oregonPolynomial
	if discriminant >= 0 {
		root := (-b + math.Sqrt(discriminant)) / 2
		f1 := root * bandwidth / 2
		f2 := f1*f1 - product
		factors[0] = oregonBandpassFactor(complex(f1, 0) + cmplx.Sqrt(complex(f2, 0)))
		root = (-b - math.Sqrt(discriminant)) / 2
		f1 = root * bandwidth / 2
		f2 = f1*f1 - product
		factors[1] = oregonBandpassFactor(complex(f1, 0) + cmplx.Sqrt(complex(f2, 0)))
		return factors
	}

	root := complex(-b/2, math.Sqrt(-discriminant)/2)
	frequency := root * complex(bandwidth/2, 0)
	delta := frequency*frequency - complex(product, 0)
	factors[0] = oregonBandpassFactor(frequency + cmplx.Sqrt(delta))
	factors[1] = oregonBandpassFactor(frequency - cmplx.Sqrt(delta))
	return factors
}

func oregonBandpassFactor(value complex128) oregonPolynomial {
	return newOregonPolynomial(
		real(cmplx.Conj(value)*value),
		-2*real(value),
		1,
	)
}

func newOregonIIRFilter(prototype oregonPrototype, passband Passband, lowFrequency, highFrequency, samplePeriod float64) (*IIRFilter, error) {
	if math.IsNaN(samplePeriod) || math.IsInf(samplePeriod, 0) || samplePeriod <= 0 {
		return nil, fmt.Errorf("sample period must be positive")
	}
	nyquist := 0.5 / samplePeriod
	switch passband {
	case LowPass:
		if err := validateCornerFrequency("low-pass high", highFrequency, nyquist); err != nil {
			return nil, err
		}
	case HighPass:
		if err := validateCornerFrequency("high-pass low", lowFrequency, nyquist); err != nil {
			return nil, err
		}
	case BandPass:
		if err := validateCornerFrequency("band-pass low", lowFrequency, nyquist); err != nil {
			return nil, err
		}
		if err := validateCornerFrequency("band-pass high", highFrequency, nyquist); err != nil {
			return nil, err
		}
		if highFrequency <= lowFrequency {
			return nil, fmt.Errorf("band-pass corner frequencies must be increasing")
		}
	default:
		return nil, fmt.Errorf("unsupported passband %d", passband)
	}

	var transformed oregonPrototype
	switch passband {
	case LowPass:
		transformed = prototype.lowpass(math.Tan(math.Pi * highFrequency * samplePeriod))
	case HighPass:
		transformed = prototype.highpass(math.Tan(math.Pi * lowFrequency * samplePeriod))
	case BandPass:
		transformed = prototype.bandpass(
			math.Tan(math.Pi*lowFrequency*samplePeriod),
			math.Tan(math.Pi*highFrequency*samplePeriod),
		)
	}
	if len(transformed.sections) == 0 {
		return nil, fmt.Errorf("filter design produced no sections")
	}

	digitalMapping := newOregonRational(newOregonPolynomial(1, -1), newOregonPolynomial(1, 1))
	sections := make([]biquad, 0, len(transformed.sections))
	overall := newOregonRational(newOregonPolynomial(1), newOregonPolynomial(1))
	for _, section := range transformed.sections {
		mapped := section.mapTo(digitalMapping)
		overall = overall.multiply(mapped)
		numerator := mapped.numerator.coefficients
		denominator := mapped.denominator.coefficients
		scale := denominator[0]
		if scale == 0 {
			scale = 1
		}
		section := biquad{}
		section.b0 = numerator[0] / scale
		if len(numerator) > 1 {
			section.b1 = numerator[1] / scale
		}
		if len(numerator) > 2 {
			section.b2 = numerator[2] / scale
		}
		if len(denominator) > 1 {
			section.a1 = denominator[1] / scale
		}
		if len(denominator) > 2 {
			section.a2 = denominator[2] / scale
		}
		sections = append(sections, section)
	}

	return &IIRFilter{
		Numerator:    append([]float64(nil), overall.numerator.coefficients...),
		Denominator:  append([]float64(nil), overall.denominator.coefficients...),
		SamplePeriod: samplePeriod,
		sections:     sections,
	}, nil
}

func validateCornerFrequency(name string, frequency, nyquist float64) error {
	if math.IsNaN(frequency) || math.IsInf(frequency, 0) || frequency <= 0 {
		return fmt.Errorf("%s corner frequency must be finite and positive", name)
	}
	if frequency >= nyquist {
		return fmt.Errorf("%s corner frequency %g must be below Nyquist frequency %g", name, frequency, nyquist)
	}
	return nil
}
