package waveform

import (
	"math"
	"time"
)

type alignedPair struct {
	StartTime  time.Time
	SampleRate float64
	Left       []float64
	Right      []float64
}

type alignedTriple struct {
	StartTime  time.Time
	SampleRate float64
	Data       [3][]float64
}

func alignPairs(left, right Seismogram) ([]alignedPair, error) {
	if err := left.Validate(); err != nil {
		return nil, err
	}
	if err := right.Validate(); err != nil {
		return nil, err
	}
	if left.Unit != "" && right.Unit != "" && left.Unit != right.Unit {
		return nil, ErrUnitMismatch
	}
	result := make([]alignedPair, 0)
	for _, leftSegment := range left.Segments {
		for _, rightSegment := range right.Segments {
			if !sameRate(leftSegment.SampleRate, rightSegment.SampleRate) {
				continue
			}
			start, end := overlapTime(leftSegment, rightSegment)
			if !start.Before(end) {
				continue
			}
			leftIndex, rightIndex, alignedStart, err := alignedIndices(start, leftSegment, rightSegment)
			if err != nil {
				return nil, err
			}
			count := minIntValue(len(leftSegment.Data)-leftIndex, len(rightSegment.Data)-rightIndex)
			if count <= 0 {
				continue
			}
			result = append(result, alignedPair{
				StartTime:  alignedStart,
				SampleRate: leftSegment.SampleRate,
				Left:       append([]float64(nil), leftSegment.Data[leftIndex:leftIndex+count]...),
				Right:      append([]float64(nil), rightSegment.Data[rightIndex:rightIndex+count]...),
			})
		}
	}
	if len(result) == 0 {
		return nil, ErrTimeDiscontinuity
	}
	return result, nil
}

func alignTriples(first, second, third Seismogram) ([]alignedTriple, error) {
	if err := first.Validate(); err != nil {
		return nil, err
	}
	if err := second.Validate(); err != nil {
		return nil, err
	}
	if err := third.Validate(); err != nil {
		return nil, err
	}
	if first.Unit != "" && second.Unit != "" && first.Unit != second.Unit ||
		first.Unit != "" && third.Unit != "" && first.Unit != third.Unit {
		return nil, ErrUnitMismatch
	}
	result := make([]alignedTriple, 0)
	for _, firstSegment := range first.Segments {
		for _, secondSegment := range second.Segments {
			for _, thirdSegment := range third.Segments {
				if !sameRate(firstSegment.SampleRate, secondSegment.SampleRate) || !sameRate(firstSegment.SampleRate, thirdSegment.SampleRate) {
					continue
				}
				start, end := overlapTime(firstSegment, secondSegment)
				thirdStart, thirdEnd := thirdSegment.StartTime, thirdSegment.EndExclusive()
				if thirdStart.After(start) {
					start = thirdStart
				}
				if thirdEnd.Before(end) {
					end = thirdEnd
				}
				if !start.Before(end) {
					continue
				}
				indices, alignedStart, err := alignedTripleIndices(start, firstSegment, secondSegment, thirdSegment)
				if err != nil {
					return nil, err
				}
				count := len(firstSegment.Data) - indices[0]
				count = minIntValue(count, len(secondSegment.Data)-indices[1])
				count = minIntValue(count, len(thirdSegment.Data)-indices[2])
				if count <= 0 {
					continue
				}
				result = append(result, alignedTriple{
					StartTime:  alignedStart,
					SampleRate: firstSegment.SampleRate,
					Data: [3][]float64{
						append([]float64(nil), firstSegment.Data[indices[0]:indices[0]+count]...),
						append([]float64(nil), secondSegment.Data[indices[1]:indices[1]+count]...),
						append([]float64(nil), thirdSegment.Data[indices[2]:indices[2]+count]...),
					},
				})
			}
		}
	}
	if len(result) == 0 {
		return nil, ErrTimeDiscontinuity
	}
	return result, nil
}

func overlapTime(left, right SeismogramSegment) (time.Time, time.Time) {
	start := left.StartTime
	if right.StartTime.After(start) {
		start = right.StartTime
	}
	end := left.EndExclusive()
	if right.EndExclusive().Before(end) {
		end = right.EndExclusive()
	}
	return start, end
}

func alignedIndices(start time.Time, left, right SeismogramSegment) (int, int, time.Time, error) {
	indices, alignedStart, err := alignedSampleIndices(start, []SeismogramSegment{left, right})
	if err != nil {
		return 0, 0, time.Time{}, err
	}
	return indices[0], indices[1], alignedStart, nil
}

func alignedTripleIndices(start time.Time, first, second, third SeismogramSegment) ([3]int, time.Time, error) {
	indices, alignedStart, err := alignedSampleIndices(start, []SeismogramSegment{first, second, third})
	if err != nil {
		return [3]int{}, time.Time{}, err
	}
	return [3]int{indices[0], indices[1], indices[2]}, alignedStart, nil
}

func alignedSampleIndices(start time.Time, segments []SeismogramSegment) ([]int, time.Time, error) {
	indices := make([]int, len(segments))
	times := make([]time.Time, len(segments))
	for index, segment := range segments {
		offset := start.Sub(segment.StartTime).Seconds() * segment.SampleRate
		indices[index] = int(math.Round(offset))
		if math.Abs(offset-float64(indices[index])) > 0.25 {
			return nil, time.Time{}, ErrTimeDiscontinuity
		}
		times[index] = segment.StartTime.Add(time.Duration(float64(indices[index]) / segment.SampleRate * float64(time.Second)))
	}
	target := times[0]
	for _, value := range times[1:] {
		if value.After(target) {
			target = value
		}
	}
	for index, segment := range segments {
		if times[index].Before(target) {
			advance := int(math.Round(target.Sub(times[index]).Seconds() * segment.SampleRate))
			indices[index] += advance
			times[index] = segment.StartTime.Add(time.Duration(float64(indices[index]) / segment.SampleRate * float64(time.Second)))
		}
		if math.Abs(times[index].Sub(target).Seconds()*segment.SampleRate) > 0.25 {
			return nil, time.Time{}, ErrTimeDiscontinuity
		}
	}
	return indices, target, nil
}

func minIntValue(left, right int) int {
	if left < right {
		return left
	}
	return right
}

func buildSeismogram(segments []SeismogramSegment, unit, sourceID string) (Seismogram, error) {
	result := Seismogram{Unit: unit, SourceID: sourceID}
	for _, segment := range segments {
		if err := result.Add(segment); err != nil {
			return Seismogram{}, err
		}
	}
	return result, nil
}
