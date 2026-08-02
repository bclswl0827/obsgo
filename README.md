# obsgo

A seismic library implemented in pure Go.

## Modules

- `signal`: detrending, tapers, IIR filters, FFT, explicitly separated
  ObsPy FFT Hilbert and OregonDSP centered-FIR Hilbert transforms,
  differentiation/integration, vector operations, and instrument-response
  removal using inventory metadata or SAC pole-zero responses.
- `waveform`: Seismogram/Trace/Stream data models, segment continuity and
  object-level signal operations. The `waveform/mseed`, `waveform/sac`, and
  `waveform/ascii` packages provide waveform IO adapters.
- `trigger`: STA/LTA characteristic functions, Z-detector, CarlSTA trigger,
  and trigger onset detection.
- `inventory`: format-neutral station metadata plus `seiscomp` and
  `stationxml` codecs. Use `inventory.Transcode` or `TranscodeBytes` to
  convert between the two formats.

Import the module that owns the required functionality, for example:

```go
import (
	"github.com/bclswl0827/obsgo/signal"
	"github.com/bclswl0827/obsgo/trigger"
)
```

## Examples

The runnable examples are under `examples/` and are executed from the
repository root:

- `peak_ground_motion_example`: remove instrument response and calculate
  three-component PGA, PGV, and PGD.
- `sta_lta_trigger_example`: band-pass filtering and recursive STA/LTA.
- `waveform_processing_example`: Trace-level preprocessing and filtering.
- `spectral_analysis_example`: spectrum, periodogram PSD, and Welch PSD.
- `rotation_polarization_example`: ZNE/ZRT rotation and polarization.
- `waveform_io_example`: MiniSEED, ASCII, and SAC round trips.
