package main

import (
	"github.com/jarcoal/httpmock"
	"log"
	"net/http"
	"os"
	"path"
	"strings"
	"testing"
)

func TestDownloadFile(t *testing.T) {

	outDir := t.TempDir()

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	broadcastUrl := "https://loopstream01.apa.at/?channel=fm4&id=2022-08-06_1900_tl_54_7DaysSat5_131332.mp3"

	httpmock.RegisterResponder("GET", broadcastUrl,
		func(req *http.Request) (*http.Response, error) {
			return httpmock.NewBytesResponse(200, httpmock.File("../_testdata/show.mp3").Bytes()), nil
		},
	)

	httpmock.RegisterResponder("HEAD", broadcastUrl,
		func(request *http.Request) (*http.Response, error) {
			resp := httpmock.NewStringResponse(200, "")
			resp.Header.Add("Content-Length", "652016")
			return resp, nil
		},
	)

	got := DownloadFile(broadcastUrl, outDir, "fileName.mp3")
	want := path.Join(outDir, "fileName.mp3")

	if got != want {
		t.Errorf("got %q want %q", got, want)
	}

	err := os.RemoveAll(outDir)
	if err != nil {
		log.Fatal(err)
	}
}

func TestSaveImage(t *testing.T) {

	imageUrl := "https://radiobilder.orf.at/fm4/imgprog/width434/keep/4DD.jpg"

	show := Show{
		TitleSanitized: "title",
		Year:           "2022",
		Images: []Images{
			{Versions: []Versions{{
				Path:  imageUrl,
				Width: 434}},
			},
		},
	}

	imageDir := t.TempDir()

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	// mock to list out the articles
	httpmock.RegisterResponder("GET", imageUrl,
		func(req *http.Request) (*http.Response, error) {
			return httpmock.NewBytesResponse(200, httpmock.File("../_testdata/4DD.jpg").Bytes()), nil
		},
	)

	httpmock.RegisterResponder("HEAD", imageUrl,
		func(request *http.Request) (*http.Response, error) {
			resp := httpmock.NewStringResponse(200, "")
			resp.Header.Add("Content-Length", "34612")
			return resp, nil
		})

	got := saveImage(imageDir, show)
	want := "/title/2022/cover.jpg"

	if !strings.Contains(got, want) {
		t.Errorf("got %q want %q", got, want)
	}

	got2 := saveImage(imageDir, show)
	want2 := "/title/2022/cover.jpg"

	if !strings.Contains(got2, want2) {
		t.Errorf("got %q want %q", got2, want2)
	}

	err := os.RemoveAll(imageDir)
	if err != nil {
		log.Fatal(err)
	}
}
