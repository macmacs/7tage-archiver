package main

import (
	"encoding/json"
	"os"
	"testing"
)

func loadBroadcast(t *testing.T, file string) Broadcast {
	t.Helper()
	data, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	b := Broadcast{}
	if err := json.Unmarshal(data, &b); err != nil {
		t.Fatal(err)
	}
	return b
}

func TestContentSegments(t *testing.T) {
	b := loadBroadcast(t, "../_testdata/davidecks.json")
	show := createShow(b)

	got := contentSegments(show)

	if len(got) != 2 {
		t.Fatalf("got %d segments, want 2", len(got))
	}

	streamStart := b.Streams[0].Start
	streamEnd := b.Streams[0].End
	// Items: 0=News (N), 1,2=content (B), 3=ad/weather (W), 4=content (B).
	// The news and the ad are cut out, leaving the content around them.
	want0 := segment{
		offset:    b.Items[0].End - streamStart,   // after the leading news
		offsetEnd: b.Items[3].Start - streamStart, // up to the ad
	}
	want1 := segment{
		offset:    b.Items[3].End - streamStart, // after the ad
		offsetEnd: streamEnd - streamStart,      // to the end of the stream
	}

	if got[0] != want0 {
		t.Errorf("segment 0 got %+v want %+v", got[0], want0)
	}
	if got[1] != want1 {
		t.Errorf("segment 1 got %+v want %+v", got[1], want1)
	}
}

func TestContentSegmentsNoItems(t *testing.T) {
	show := Show{Streams: []Streams{{LoopStreamID: "id", Start: 0, End: 1000}}}

	if got := contentSegments(show); got != nil {
		t.Errorf("got %+v, want nil for show without items", got)
	}
}

func TestGetSegmentUrl(t *testing.T) {
	show := Show{Streams: []Streams{{Progressive: "https://loopstreamfm4.apa.at?channel=fm4&id=LoopStreamID"}}}

	got := getSegmentUrl(show, segment{offset: 175000, offsetEnd: 3561000})
	want := "https://loopstreamfm4.apa.at?channel=fm4&id=LoopStreamID&offset=175000&offsetende=3561000"

	if got != want {
		t.Errorf("got %q want %q", got, want)
	}
}
