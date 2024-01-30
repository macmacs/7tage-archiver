package main

import (
	"github.com/jarcoal/httpmock"
	"net/http"
	"reflect"
	"testing"
)

func TestSearchBroadcastUrl(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	searchUrl := "https://audioapi.orf.at/fm4/api/json/current/search?q=Swound+Sound"

	httpmock.RegisterResponder("GET", searchUrl,
		func(req *http.Request) (*http.Response, error) {
			resp, err := httpmock.NewJsonResponse(200, httpmock.File("../_testdata/searchresult.json"))
			if err != nil {
				return httpmock.NewStringResponse(500, ""), nil
			}
			return resp, nil
		},
	)

	got := SearchBroadcastUrls("Swound Sound")
	want := []string{"https://audioapi.orf.at/fm4/api/json/4.0/broadcast/4SS/20220806"}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %q want %q", got, want)
	}
}

func TestGetSearchResults(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	searchUrl := "https://audioapi.orf.at/fm4/api/json/current/search?q=Swound+Sound"

	httpmock.RegisterResponder("GET", searchUrl,
		func(req *http.Request) (*http.Response, error) {
			resp, err := httpmock.NewJsonResponse(200, httpmock.File("../_testdata/searchresult.json"))
			if err != nil {
				return httpmock.NewStringResponse(500, ""), nil
			}
			return resp, nil
		},
	)

	got, err := getSearchResults("Swound Sound")

	if err != nil {
		t.Errorf("Retrieving search results failed.")
	}

	want := SearchResult{
		Took:       0,
		IsTimedOut: false,
		Length:     0,
		Total:      2,
		Hits:       nil,
		Suggest:    nil,
	}

	if got.Total != want.Total {
		t.Errorf("Wrong amount of search results. got %d want %d", got.Total, want.Total)
	}
}

func TestGetNoSearchResults(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	searchUrl := "https://audioapi.orf.at/fm4/api/json/current/search?q=Zummerservice"

	httpmock.RegisterResponder("GET", searchUrl,
		func(req *http.Request) (*http.Response, error) {
			resp, err := httpmock.NewJsonResponse(200, httpmock.File("../_testdata/no_searchresult.json"))
			if err != nil {
				return httpmock.NewStringResponse(500, ""), nil
			}
			return resp, nil
		},
	)

	got, err := getSearchResults("Zummerservice")

	if err != nil {
		t.Errorf("Retrieving search results failed.")
	}

	want := SearchResult{
		Took:       0,
		IsTimedOut: false,
		Length:     0,
		Total:      0,
		Hits:       nil,
		Suggest:    nil,
	}

	if got.Total != want.Total {
		t.Errorf("Wrong amount of search results. got %d want %d", got.Total, want.Total)
	}

	if got.Suggest[0].Text != "zimmerservice" {
		t.Errorf("Wrong amount of search results. got %s want %s", got.Suggest[0].Text, "zimmerservice")
	}
}
