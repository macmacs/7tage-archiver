# Show Archiver - 7Tage

Downloads the last aired show to an `mp3` file and tags it with info and image.

```bash
Usage of bin/fm4-archiver:
  -out-base-dir string
        Location of your shows (default "/music")
  -progress
        Print progress
  -show string
        A FM4 Show (default "Davidecks")
```

## CLI

Run e.g.

```bash
$ 7tage-archiver -show "Graue Lagune" -out-base-dir . -progress
```

Result:

```bash
$ eyeD3 Graue_Lagune/2022/Graue_Lagune_20220424.mp3

----------------------------------------------------------------------------
Time: 01:02:24	MPEG1, Layer III	[ 192 kb/s @ 48000 Hz - Joint stereo ]
----------------------------------------------------------------------------
ID3 v2.4:
title: Graue Lagune - 20220424
artist: Graue Lagune
album: 2022
album artist: Graue Lagune
recording date: 2022
track:
FRONT_COVER Image: [Size: 47472 bytes] [Type: image/jpeg]
Description: Front cover
----------------------------------------------------------------------------
```

## Docker

```bash
docker run --rm \
  -v /your/fm4/shows/folder/:/music \
  ghcr.io/macmacs/7tage-archiver \
  -show "Graue Lagune" \
  -progress
```