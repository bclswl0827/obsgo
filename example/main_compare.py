#!/usr/bin/env python3

from pathlib import Path

import numpy as np
from obspy import read
from obspy.signal.trigger import recursive_sta_lta_py, trigger_onset

STA_WINDOW = 0.5
LTA_WINDOW = 30.0
TRIGGER_ON = 8.0
TRIGGER_OFF = 0.5

LOW_CUTOFF = 0.5
HIGH_CUTOFF = 10.0
FILTER_TAPS = 101


def new_bandpass_fir_filter(
    low_cutoff: float,
    high_cutoff: float,
    sample_rate: float,
    num_taps: int,
) -> np.ndarray:
    low = 2.0 * np.pi * low_cutoff / sample_rate
    high = 2.0 * np.pi * high_cutoff / sample_rate
    center = num_taps // 2
    coefficients = np.empty(num_taps, dtype=np.float64)

    for i in range(num_taps):
        n = float(i - center)
        if i == center:
            coefficient = (high - low) / np.pi
        else:
            coefficient = (np.sin(high * n) - np.sin(low * n)) / (np.pi * n)

        coefficient *= 0.54 - 0.46 * np.cos(
            2.0 * np.pi * float(i) / float(num_taps - 1)
        )
        coefficients[i] = coefficient

    return coefficients


def apply_fir_filter(data: np.ndarray, kernel: np.ndarray) -> np.ndarray:
    state = np.zeros(len(kernel) - 1, dtype=np.float64)
    extended_data = np.concatenate((state, data))

    return np.correlate(extended_data, kernel, mode="valid")


def main() -> None:
    mseed_path = Path(__file__).with_name("testdata.mseed")
    stream = read(str(mseed_path), format="MSEED")
    if not stream:
        raise RuntimeError(f"no traces found in {mseed_path}")
    print("Successfully read miniSEED file")

    trace_data = np.concatenate(
        [np.asarray(trace.data, dtype=np.float64) for trace in stream]
    )
    sample_rate = float(stream[0].stats.sampling_rate)
    start_time = stream[0].stats.starttime
    print("Successfully generated trace data")

    kernel = new_bandpass_fir_filter(
        LOW_CUTOFF,
        HIGH_CUTOFF,
        sample_rate,
        FILTER_TAPS,
    )
    filtered_data = apply_fir_filter(trace_data, kernel)
    print("Filtered trace data")

    sta_lta = recursive_sta_lta_py(
        filtered_data,
        int(STA_WINDOW * sample_rate),
        int(LTA_WINDOW * sample_rate),
    )
    print("Calculated STA/LTA successfully")

    onsets = trigger_onset(
        sta_lta,
        TRIGGER_ON,
        TRIGGER_OFF,
        max_len=np.iinfo(np.int64).max,
        max_len_delete=False,
    )
    print(f"Detected {len(onsets)} event(s)")

    for onset, _ in onsets:
        event_time = start_time + int(onset) / sample_rate
        print(f"Detected event at: {event_time}")


if __name__ == "__main__":
    main()
