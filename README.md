# Radio Show Archiver - 7Tage - 30 days

Archives FM4 shows that aired within the last 30 days to `mp3` files and tags them with info and image.

Shows can be resolved by name (via the built-in search), from a `sound.orf.at`
Sendung URL, or from a stable `programKey` (e.g. `4DD`). Prefer the programKey
for recurring downloads: a Sendung URL points at a single episode whose id ages
out of the 30-day window, while a programKey stays valid.

```bash
Usage of bin/fm4-archiver:

download
  -out-base-dir string
        Location of your shows (default "/music")
  -show string
        A Radio FM4 Show (default "Davidecks")

url
  Takes a sound.orf.at Sendung URL, e.g.
  https://sound.orf.at/radio/fm4/sendung/42628/davidecks
  or a stable programKey, e.g. 4DD
  -out-base-dir string
        Location of your shows (default "./music")

search
  -query string
        Search show by query
```

## CLI

Download a show by name:

```bash
$ 7tage-archiver download -show "Graue Lagune" -out-base-dir .
```

Download every available episode of a show from its sound.orf.at URL:

```bash
$ 7tage-archiver url https://sound.orf.at/radio/fm4/sendung/42628/davidecks -out-base-dir .
```

Or, more durably, from its stable programKey (recommended for cron jobs):

```bash
$ 7tage-archiver url 4DD -out-base-dir .
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