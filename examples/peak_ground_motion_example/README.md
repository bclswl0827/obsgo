# Peak ground motion example

This example reproduces the processing flow in `_acceleration.py`:

1. read miniSEED records with `github.com/bclswl0827/mseedio`, merge them, and linearly interpolate record gaps;
2. remove the linear trend and mean;
3. apply a 5% cosine taper;
4. apply the default response-removal demean/taper and remove the instrument response for acceleration, velocity, and displacement;
5. apply a four-pole 0.1–10 Hz Butterworth band-pass;
6. convert metres to centimetres and calculate the peak three-component vector magnitude.

Run from the repository root:

```sh
go run ./examples/peak_ground_motion_example \
  -inventory ./examples/_data/seiscomp.xml \
  -east ./examples/_data/2026.211.16.53.54.0541.SHAKE.AS.00.EHE.D.mseed \
  -north ./examples/_data/2026.211.16.53.54.0541.SHAKE.AS.00.EHN.D.mseed \
  -vertical ./examples/_data/2026.211.16.53.54.0541.SHAKE.AS.00.EHZ.D.mseed
```

`-format` defaults to `auto`; `seiscomp` and `stationxml` can also be selected explicitly.

Response removal uses the fast power-of-two FFT by default. Add `-obspy-fft`
to use ObsPy's arbitrary FFT-length selection for numerical comparisons; that
mode requires substantially more CPU time in the pure Go implementation.
