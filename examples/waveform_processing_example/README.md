# Waveform processing example

This example reads one MiniSEED trace and applies common preprocessing at the
`waveform.Trace` level:

1. remove mean and linear trend;
2. apply a five-percent cosine taper;
3. apply a fourth-order, zero-phase 0.1–10 Hz Butterworth band-pass;
4. report segment metadata and peak absolute amplitude.

Run it from the repository root:

```sh
go run ./examples/waveform_processing_example
```

Use `-input` to select another MiniSEED file. Processing is performed per
segment, so a time gap is never filtered across.
