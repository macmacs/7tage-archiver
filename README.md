# Radio Show Archiver - 7Tage

Archives the last aired show to an `mp3` file and tags it with info and image.

```bash
Usage of bin/fm4-archiver:

download
  -out-base-dir string
        Location of your shows (default "/music")

printshowids
  Prints a list of show IDs

```

## CLI

Run e.g.

```bash
$ 7tage-archiver download -show "4GL" -out-base-dir .
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
  download \
  -show "4GL"
```