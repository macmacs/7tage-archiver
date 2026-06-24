package main

import "sort"

// removeTypes are the broadcast item types cut out of a download: News and the
// Weather/ad spot. Everything else (show content, jingles, and the untagged
// audio between tagged items) is kept.
var removeTypes = map[string]bool{"N": true, "W": true}

// segment is a slice of a stream expressed as millisecond offsets relative to
// streams[0].start, ready to be passed to the loopstream offset/offsetende
// query params.
type segment struct {
	offset    int64
	offsetEnd int64
}

// contentSegments returns the ranges to download with the news and ad/weather
// spots removed. It starts from the full stream and cuts out the intervals of
// the removed item types, keeping everything in between (tagged items are
// sparse, so anything not explicitly a removed type is real show audio).
// Returns nil when there is no stream or nothing to cut, signalling a plain
// full-stream download (unchanged legacy behaviour).
func contentSegments(show Show) []segment {
	if len(show.Streams) == 0 {
		return nil
	}
	streamStart := show.Streams[0].Start
	streamEnd := show.Streams[0].End
	if streamEnd <= streamStart {
		return nil
	}

	type interval struct{ start, end int64 }
	var cuts []interval
	for _, item := range show.Items {
		if !removeTypes[item.Type] {
			continue
		}
		s, e := item.Start, item.End
		if s < streamStart {
			s = streamStart
		}
		if e > streamEnd {
			e = streamEnd
		}
		if e > s {
			cuts = append(cuts, interval{s, e})
		}
	}
	if len(cuts) == 0 {
		return nil
	}

	sort.Slice(cuts, func(i, j int) bool { return cuts[i].start < cuts[j].start })
	merged := cuts[:1]
	for _, c := range cuts[1:] {
		last := &merged[len(merged)-1]
		if c.start <= last.end {
			if c.end > last.end {
				last.end = c.end
			}
		} else {
			merged = append(merged, c)
		}
	}

	// Keep the gaps between (and around) the cuts.
	var segments []segment
	cursor := streamStart
	for _, c := range merged {
		if c.start > cursor {
			segments = append(segments, segment{cursor - streamStart, c.start - streamStart})
		}
		cursor = c.end
	}
	if cursor < streamEnd {
		segments = append(segments, segment{cursor - streamStart, streamEnd - streamStart})
	}
	return segments
}
