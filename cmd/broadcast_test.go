package main

import (
	"net/http"
	"reflect"
	"testing"

	"github.com/jarcoal/httpmock"
)

func TestCleanProgressiveURL(t *testing.T) {
	// Real v5.0 values for Davidecks 42628.
	tmpl := "https://loopstreamfm4.apa.at?channel=fm4&id=2026-06-20_1859_tl_54_7DaysSat5_180163.mp3{&offset}{&offsetende}{&shoutcast}{&player}{&referer}{&userid}"
	prog := "https://loopstreamfm4.apa.at?channel=fm4&id=2026-06-20_1859_tl_54_7DaysSat5_180163.mp3&offset=173000&offsetende=7203000{&shoutcast}{&player}{&referer}{&userid}"
	want := "https://loopstreamfm4.apa.at?channel=fm4&id=2026-06-20_1859_tl_54_7DaysSat5_180163.mp3"

	// Prefers the template (no pre-filled offsets to strip).
	if got := cleanProgressiveURL(prog, tmpl); got != want {
		t.Errorf("template path: got %q want %q", got, want)
	}
	// Falls back to the expanded URL and strips the baked offsets + tokens.
	if got := cleanProgressiveURL(prog, ""); got != want {
		t.Errorf("fallback path: got %q want %q", got, want)
	}
}

// TestGetBroadcastV5Mapping drives the full v5.0 path end-to-end: the wrapped
// payload is unwrapped, ISO timestamps are mapped to epoch-ms, the stream's
// Progressive URL is cleaned, and trim.go's contentSegments still cuts the
// leading news and the two mid/trailing ad spots. This is the regression guard
// for the migration off the legacy v4.0 href + loopstream01.apa.at host.
func TestGetBroadcastV5Mapping(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	broadcastUrl := "https://audioapi.orf.at/fm4/api/json/5.0/broadcast/42628?items=1000"
	httpmock.RegisterResponder("GET", broadcastUrl,
		func(req *http.Request) (*http.Response, error) {
			return httpmock.NewJsonResponse(200, httpmock.File("../_testdata/broadcast_42628_full_v5.json"))
		},
	)

	// getBroadcast appends ?items= itself; the caller passes the bare href.
	broadcast := getBroadcast("https://audioapi.orf.at/fm4/api/json/5.0/broadcast/42628")
	show := createShow(broadcast)

	if show.Title != "Davidecks" {
		t.Fatalf("Title got %q want %q", show.Title, "Davidecks")
	}

	// Stream Start/End map to the broadcast range, so streamEnd-streamStart ==
	// duration and item offset math resolves to ms-from-broadcast-start; this
	// is what trim.go hands to the loopstream &offset/&offsetende params.
	wantDuration := int64(7203000)
	if got := show.Streams[0].End - show.Streams[0].Start; got != wantDuration {
		t.Errorf("stream duration got %d want %d", got, wantDuration)
	}

	got := contentSegments(show)
	want := []segment{
		{offset: 248500, offsetEnd: 3561000},  // after the leading news, up to the first ad
		{offset: 3613000, offsetEnd: 7153000}, // between the two ads, to the last ad
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("segments got %+v want %+v", got, want)
	}

	// And the download URL for the first segment targets the per-station host.
	segURL := getSegmentUrl(show, want[0])
	wantURL := "https://loopstreamfm4.apa.at?channel=fm4&id=2026-06-20_1859_tl_54_7DaysSat5_180163.mp3&offset=248500&offsetende=3561000"
	if segURL != wantURL {
		t.Errorf("segment URL got %q want %q", segURL, wantURL)
	}
}
