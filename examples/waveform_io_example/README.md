# Waveform IO example

This example reads MiniSEED through the waveform adapter, interpolates record
gaps during ingest, writes the result as the repository's metadata-preserving
ASCII format, reads it back, writes SAC, and reads SAC back again.

Run it from the repository root:

```sh
go run ./examples/waveform_io_example
```

Generated files are written below `examples/waveform_io_example/output` by
default. Use `-input` and `-output` to select different paths. SAC stores a
single continuous trace, so this example explicitly interpolates MiniSEED
record gaps before writing it.
