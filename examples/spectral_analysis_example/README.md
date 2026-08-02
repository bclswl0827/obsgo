# Spectral analysis example

This example demonstrates the `Trace`-level spectrum APIs on a MiniSEED
trace:

- a Hann-windowed one-sided amplitude spectrum;
- a demeaned one-sided periodogram PSD;
- an ObsPy-style Welch PSD using `DefaultWelchOptions` (256-sample Hann
  windows, 50% overlap, constant detrending, and no incomplete tail window).

Run it from the repository root:

```sh
go run ./examples/spectral_analysis_example
```

Use `-input` to select another file. Each discontinuous waveform segment is
analyzed independently.
