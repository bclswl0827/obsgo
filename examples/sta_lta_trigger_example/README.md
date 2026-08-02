# STA/LTA trigger example

This example reads the vertical EHZ MiniSEED waveform from `examples/_data`,
applies a second-order 0.5–10 Hz Butterworth band-pass, calculates a recursive
STA/LTA characteristic function, and reports detected trigger onsets.

Run it from the repository root:

```sh
go run ./examples/sta_lta_trigger_example
```

The default input is:

```text
examples/_data/2026.211.16.53.54.0541.SHAKE.AS.00.EHZ.D.mseed
```

`main_compare.py` contains the corresponding Python/NumPy comparison
implementation.
