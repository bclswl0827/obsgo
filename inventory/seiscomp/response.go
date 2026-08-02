package seiscomp

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/bclswl0827/obsgo/inventory"
)

func responseToInventory(response ResponsePAZ) (*inventory.Response, error) {
	zeros, err := parsePoleZeros(response.Zeros)
	if err != nil {
		return nil, fmt.Errorf("decode response %q zeros: %w", response.PublicID, err)
	}
	poles, err := parsePoleZeros(response.Poles)
	if err != nil {
		return nil, fmt.Errorf("decode response %q poles: %w", response.PublicID, err)
	}
	converted := &inventory.Response{
		PublicID:               response.PublicID,
		Name:                   response.Name,
		Type:                   response.Type,
		Gain:                   response.Gain,
		GainFrequency:          response.GainFrequency,
		NormalizationFactor:    response.NormalizationFactor,
		NormalizationFrequency: response.NormalizationFrequency,
		Zeros:                  zeros,
		Poles:                  poles,
	}
	if len(zeros) > 0 || len(poles) > 0 || response.Type != "" {
		converted.Stages = append(converted.Stages, inventory.Stage{
			Number: 1,
			PolesZeros: &inventory.PolesZeros{
				Name:                   response.Name,
				ResourceID:             response.PublicID,
				TransferFunctionType:   response.Type,
				NormalizationFactor:    response.NormalizationFactor,
				NormalizationFrequency: response.NormalizationFrequency,
				Zeros:                  zeros,
				Poles:                  poles,
			},
			Gain: &inventory.StageGain{Value: response.Gain, Frequency: response.GainFrequency},
		})
	}
	return converted, nil
}

func responseFromInventory(response inventory.Response) ResponsePAZ {
	zeros := response.Zeros
	poles := response.Poles
	if len(zeros) == 0 || len(poles) == 0 {
		for _, stage := range response.Stages {
			if stage.PolesZeros == nil {
				continue
			}
			if len(zeros) == 0 {
				zeros = stage.PolesZeros.Zeros
			}
			if len(poles) == 0 {
				poles = stage.PolesZeros.Poles
			}
			if response.Name == "" {
				response.Name = stage.PolesZeros.Name
			}
			if response.PublicID == "" {
				response.PublicID = stage.PolesZeros.ResourceID
			}
			if response.Type == "" {
				response.Type = stage.PolesZeros.TransferFunctionType
			}
			if response.NormalizationFactor == 0 {
				response.NormalizationFactor = stage.PolesZeros.NormalizationFactor
			}
			if response.NormalizationFrequency == 0 {
				response.NormalizationFrequency = stage.PolesZeros.NormalizationFrequency
			}
			if response.Gain == 0 && stage.Gain != nil {
				response.Gain = stage.Gain.Value
				response.GainFrequency = stage.Gain.Frequency
			}
		}
	}
	return ResponsePAZ{
		PublicID:               response.PublicID,
		Name:                   response.Name,
		Type:                   response.Type,
		Gain:                   response.Gain,
		GainFrequency:          response.GainFrequency,
		NormalizationFactor:    response.NormalizationFactor,
		NormalizationFrequency: response.NormalizationFrequency,
		NumberOfZeros:          len(zeros),
		NumberOfPoles:          len(poles),
		Zeros:                  formatPoleZeros(zeros),
		Poles:                  formatPoleZeros(poles),
	}
}

func parsePoleZeros(text string) ([]inventory.PoleZero, error) {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil, nil
	}
	values := strings.Fields(text)
	result := make([]inventory.PoleZero, 0, len(values))
	for index, value := range values {
		if !strings.HasPrefix(value, "(") || !strings.HasSuffix(value, ")") {
			return nil, fmt.Errorf("invalid pole/zero %q", value)
		}
		parts := strings.Split(strings.TrimSuffix(strings.TrimPrefix(value, "("), ")"), ",")
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid pole/zero %q", value)
		}
		realPart, err := strconv.ParseFloat(parts[0], 64)
		if err != nil {
			return nil, fmt.Errorf("invalid real part at index %d: %w", index, err)
		}
		imaginaryPart, err := strconv.ParseFloat(parts[1], 64)
		if err != nil {
			return nil, fmt.Errorf("invalid imaginary part at index %d: %w", index, err)
		}
		result = append(result, inventory.PoleZero{Number: index, Real: realPart, Imaginary: imaginaryPart})
	}
	return result, nil
}

func formatPoleZeros(values []inventory.PoleZero) string {
	parts := make([]string, 0, len(values))
	for _, value := range values {
		parts = append(parts, "("+strconv.FormatFloat(value.Real, 'g', -1, 64)+","+strconv.FormatFloat(value.Imaginary, 'g', -1, 64)+")")
	}
	return strings.Join(parts, " ")
}
