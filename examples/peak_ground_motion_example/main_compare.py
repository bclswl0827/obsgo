# -*- coding: utf-8 -*-
from obspy import Stream, read, read_inventory
from numpy import max, abs, sqrt


def main():
    INVENTORY = read_inventory("../_data/seiscomp.xml")
    FILES = [
        # "../_data/2026.211.16.53.54.0541.SHAKE.AS.00.EHE.D.mseed",
        # "../_data/2026.211.16.53.54.0541.SHAKE.AS.00.EHN.D.mseed",
        # "../_data/2026.211.16.53.54.0541.SHAKE.AS.00.EHZ.D.mseed",
        "../_data/2026.208.02.09.53.0225.SHAKE.AS.00.EHZ.D.mseed",
        "../_data/2026.208.02.09.53.0225.SHAKE.AS.00.EHE.D.mseed",
        "../_data/2026.208.02.09.53.0225.SHAKE.AS.00.EHN.D.mseed",
    ]
    CHANNELS = ["EHE", "EHN", "EHZ"]
    BANDPASS = dict(freqmin=0.1, freqmax=10.0)

    raw_data = Stream()
    for f in FILES:
        raw_data += read(f, format="mseed")
    raw_data.merge(fill_value="interpolate")

    processed_data = {"ACC": {}, "VEL": {}, "DISP": {}}
    for i, component in enumerate(CHANNELS):
        for unit in processed_data.keys():
            data_copy = raw_data[i].copy()
            data_copy.detrend("linear")
            data_copy.detrend("demean")
            data_copy.taper(max_percentage=0.05, type="cosine")
            data_copy.remove_response(
                inventory=INVENTORY,
                output=unit,
            )
            data_copy.filter("bandpass", **BANDPASS)
            processed_data[unit][component] = data_copy.data * 100

    results = {}
    for key in processed_data.keys():
        value_arr = []
        for idx, _ in enumerate(processed_data[key][CHANNELS[0]]):
            val = 0
            for comp in CHANNELS:
                val += (processed_data[key][comp][idx] ** 2)
            value_arr.append(sqrt(abs(val)))
        results[key] = max(value_arr)

    print(">>> Peak Ground Motion (3-Component Vector Sum)")
    print(f"PGA  (cm/s/s): {results['ACC']:.4f}")
    print(f"PGV  (cm/s)  : {results['VEL']:.4f}")
    print(f"PGD  (cm)    : {results['DISP']:.4f}")


if __name__ == "__main__":
    main()
