# Rotation and polarization example

This example reads three MiniSEED components, performs a time-aligned ZNE to
ZRT rotation, and analyzes one-second sliding polarization windows. The
polarization method can be covariance-based Flinn (`flinn`) or particle-motion
(`pm`).

Run it from the repository root:

```sh
go run ./examples/rotation_polarization_example
go run ./examples/rotation_polarization_example -back-azimuth 210 -method pm
```

The waveform layer checks component sample rates and timestamps, then rotates
only overlapping samples. It does not concatenate samples across gaps.
