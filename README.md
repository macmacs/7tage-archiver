# Radio Show Archiver - 7Tage - 30 days

Archives FM4 shows that aired within the last 30 days to `mp3` files and tags them with info and image.

```bash
Usage of bin/fm4-archiver:

download
  -out-base-dir string
        Location of your shows (default "/music")
  -show string
        A Radio FM4 Show (default "Davidecks")
        
search
  -query string
        Search show by query
```

## CLI

Run e.g.

```bash
$ 7tage-archiver download -show "Graue Lagune" -out-base-dir .
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
  -show "Graue Lagune"
```