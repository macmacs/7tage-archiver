package main

import (
	"net/http"
	"reflect"
	"testing"

	"github.com/jarcoal/httpmock"
)

const broadcastIdDavidecks = "42628"
const soundUrlDavidecks = "https://sound.orf.at/radio/fm4/sendung/42628/davidecks"

func TestResolveBroadcastUrlsFromSoundUrl(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	// Step 1: broadcast/{id} resolves the programKey from the Sound Sendung URL.
	// The v5.0 API wraps the broadcast inside {"timezoneOffset":...,"payload":{...}}.
	broadcastUrl := "https://audioapi.orf.at/fm4/api/json/5.0/broadcast/" + broadcastIdDavidecks
	httpmock.RegisterResponder("GET", broadcastUrl,
		func(req *http.Request) (*http.Response, error) {
			return httpmock.NewJsonResponse(200, httpmock.File("../_testdata/broadcast_42628_v5.json"))
		},
	)

	// Step 2: broadcasts/program/{programKey} lists every episode.
	programUrl := "https://audioapi.orf.at/fm4/api/json/5.0/broadcasts/program/4DD"
	httpmock.RegisterResponder("GET", programUrl,
		func(req *http.Request) (*http.Response, error) {
			return httpmock.NewJsonResponse(200, httpmock.File("../_testdata/program_4DD.json"))
		},
	)

	got := ResolveBroadcastUrls(soundUrlDavidecks)
	want := []string{
		"https://audioapi.orf.at/fm4/api/json/5.0/broadcast/42628",
		"https://audioapi.orf.at/fm4/api/json/5.0/broadcast/42536",
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %q want %q", got, want)
	}
}

func TestResolveBroadcastUrlsFromProgramKey(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	// A bare programKey skips the broadcast/{id} lookup and lists episodes directly.
	programUrl := "https://audioapi.orf.at/fm4/api/json/5.0/broadcasts/program/4DD"
	httpmock.RegisterResponder("GET", programUrl,
		func(req *http.Request) (*http.Response, error) {
			return httpmock.NewJsonResponse(200, httpmock.File("../_testdata/program_4DD.json"))
		},
	)

	got := ResolveBroadcastUrls("4DD")
	want := []string{
		"https://audioapi.orf.at/fm4/api/json/5.0/broadcast/42628",
		"https://audioapi.orf.at/fm4/api/json/5.0/broadcast/42536",
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %q want %q", got, want)
	}
}

func TestResolveBroadcastUrlsFromSoundUrlWithoutSlug(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	broadcastUrl := "https://audioapi.orf.at/fm4/api/json/5.0/broadcast/" + broadcastIdDavidecks
	httpmock.RegisterResponder("GET", broadcastUrl,
		func(req *http.Request) (*http.Response, error) {
			return httpmock.NewJsonResponse(200, httpmock.File("../_testdata/broadcast_42628_v5.json"))
		},
	)
	programUrl := "https://audioapi.orf.at/fm4/api/json/5.0/broadcasts/program/4DD"
	httpmock.RegisterResponder("GET", programUrl,
		func(req *http.Request) (*http.Response, error) {
			return httpmock.NewJsonResponse(200, httpmock.File("../_testdata/program_4DD.json"))
		},
	)

	// A Sound Sendung URL without the trailing title slug must also resolve.
	got := ResolveBroadcastUrls("https://sound.orf.at/radio/fm4/sendung/42628")
	if len(got) != 2 {
		t.Errorf("expected 2 broadcast URLs, got %d (%q)", len(got), got)
	}
}

func TestSoundUrlPattern(t *testing.T) {
	cases := []struct {
		url   string
		match bool
		bcID  string
	}{
		{"https://sound.orf.at/radio/fm4/sendung/42628/davidecks", true, "42628"},
		{"https://sound.orf.at/radio/fm4/sendung/42628", true, "42628"},
		{"http://sound.orf.at/radio/fm4/sendung/42628/davidecks", true, "42628"},
		{"https://sound.orf.at/radio/oe1/sendung/42628/davidecks", false, ""}, // non-fm4 not supported
		{"https://sound.orf.at/podcast/oe3/fruehstueck-bei-mir/x", false, ""}, // podcast, not radio/sendung
		{"https://sound.orf.at/radio/fm4/sendung/", false, ""},                // missing id
		{"https://fm4.orf.at/player/20220801/OGMO", false, ""},                // old player URL
		{"not a url", false, ""},
	}
	for _, c := range cases {
		m := soundUrlPattern.FindStringSubmatch(c.url)
		gotMatch := m != nil
		if gotMatch != c.match {
			t.Errorf("match=%v want %v for %s", gotMatch, c.match, c.url)
			continue
		}
		if c.match && m[1] != c.bcID {
			t.Errorf("broadcastId=%q want %q for %s", m[1], c.bcID, c.url)
		}
	}
}

func TestGetProgramKey(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	broadcastUrl := "https://audioapi.orf.at/fm4/api/json/5.0/broadcast/" + broadcastIdDavidecks
	httpmock.RegisterResponder("GET", broadcastUrl,
		func(req *http.Request) (*http.Response, error) {
			return httpmock.NewJsonResponse(200, httpmock.File("../_testdata/broadcast_42628_v5.json"))
		},
	)

	got := getProgramKey(broadcastIdDavidecks)
	want := "4DD"
	if got != want {
		t.Errorf("got %q want %q", got, want)
	}
}

func TestGetProgramEpisodes(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	programUrl := "https://audioapi.orf.at/fm4/api/json/5.0/broadcasts/program/4DD"
	httpmock.RegisterResponder("GET", programUrl,
		func(req *http.Request) (*http.Response, error) {
			return httpmock.NewJsonResponse(200, httpmock.File("../_testdata/program_4DD.json"))
		},
	)

	got := getProgramEpisodes("4DD")
	want := []string{
		"https://audioapi.orf.at/fm4/api/json/5.0/broadcast/42628",
		"https://audioapi.orf.at/fm4/api/json/5.0/broadcast/42536",
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %q want %q", got, want)
	}
}
