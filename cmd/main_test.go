package main

import (
	"github.com/bogem/id3v2"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"strings"
	"testing"
	"time"
)
import "github.com/jarcoal/httpmock"

func TestGetBroadcast(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	broadcastUrl := "https://audioapi.orf.at/fm4/api/json/4.0/broadcast/4DD/20220806"

	// mock to list out the articles
	httpmock.RegisterResponder("GET", broadcastUrl,
		func(req *http.Request) (*http.Response, error) {
			resp, err := httpmock.NewJsonResponse(200, httpmock.File("../_testdata/davidecks.json"))
			if err != nil {
				return httpmock.NewStringResponse(500, ""), nil
			}
			return resp, nil
		},
	)

	got := getBroadcast(broadcastUrl)
	want := Broadcast{
		Title: "Davidecks",
	}

	if got.Title != want.Title {
		t.Errorf("got %q want %q", got.Title, want.Title)
	}
}

func TestCreateShow(t *testing.T) {

	b := Broadcast{
		Title:        "Title Test",
		Subtitle:     "<p>Description</p> ",
		BroadcastDay: 20220806,
		StartISO:     time.Date(2022, 8, 6, 19, 3, 0, 0, time.Local),
	}

	got := createShow(b)
	want := Show{
		Title:          "Title Test",
		TitleSanitized: "Title_Test",
		Description:    "Description",
		BroadcastDay:   "20220806",
		Year:           "2022",
		Images:         nil,
		Streams:        nil,
	}

	if got.Title != want.Title {
		t.Errorf("got %q want %q", got.Title, want.Title)
	}

	if got.TitleSanitized != want.TitleSanitized {
		t.Errorf("got %q want %q", got.TitleSanitized, want.TitleSanitized)
	}

	if got.Description != want.Description {
		t.Errorf("got %q want %q", got.Description, want.Description)
	}

	if got.Year != want.Year {
		t.Errorf("got %q want %q", got.Year, want.Year)
	}
}

func TestGetDownloadUrl(t *testing.T) {

	stream := Streams{
		LoopStreamID: "LoopStreamID",
	}
	show := Show{
		Streams: []Streams{stream},
	}

	got := getDownloadUrl(show)
	want := "https://loopstream01.apa.at/?channel=fm4&id=LoopStreamID"

	if got != want {
		t.Errorf("got %q want %q", got, want)
	}
}

func TestGetOutputPath(t *testing.T) {
	destDir := "destDir"
	got := getOutputPath(destDir, Show{TitleSanitized: "title", Year: "2022"})
	want := "destDir/title/2022"

	if got != want {
		t.Errorf("got %q want %q", got, want)
	}
}
func TestGetFileName(t *testing.T) {
	got := getFileName(Show{TitleSanitized: "title", BroadcastDay: "20220806"})
	want := "title_20220806.mp3"

	if got != want {
		t.Errorf("got %q want %q", got, want)
	}
}

func TestDownloadFile(t *testing.T) {

	outDir := t.TempDir()

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	broadcastUrl := "https://loopstream01.apa.at/?channel=fm4&id=2022-08-06_1900_tl_54_7DaysSat5_131332.mp3"

	// mock to list out the articles
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
		})

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

func TestWriteID3Tag(t *testing.T) {

	mp3path := path.Join(t.TempDir(), "file.mp3")
	imagePath := path.Join(t.TempDir(), "cover.jpg")

	log.Println(mp3path)

	copyFile("../_testdata/show.mp3", mp3path)
	copyFile("../_testdata/4DD.jpg", imagePath)

	show := Show{
		Title:          "Title Test",
		TitleSanitized: "Title_Test",
		Description:    "Description",
		BroadcastDay:   "20220806",
		Year:           "2022",
		Images:         nil,
		Streams:        nil,
	}

	writeId3Tag(mp3path, imagePath, show)

	got, err := id3v2.Open(mp3path, id3v2.Options{Parse: true})
	if err != nil {
		log.Fatal("Error while opening mp3 file: ", err)
	}

	want := "Title Test - 20220806"

	if got.Title() != want {
		t.Errorf("Title tag wrong! Got %s, wanted %s.", got.Title(), show.Title)
	}

	if got.GetFrames("APIC")[0].Size() == 0 {
		t.Error("Attached image has zero bytes!")
	}

	t.Cleanup(func() {
		err := os.RemoveAll(t.TempDir())
		if err != nil {
			log.Fatal(err)
		}
	})
}

func copyFile(src string, dst string) {
	fin, err := os.Open(src)
	if err != nil {
		log.Fatal(err)
	}
	defer func(fin *os.File) {
		err := fin.Close()
		if err != nil {
			logError(err)
		}
	}(fin)

	fout, err := os.Create(dst)
	if err != nil {
		log.Fatal(err)
	}
	defer func(fout *os.File) {
		err := fout.Close()
		if err != nil {
			logError(err)
		}
	}(fout)

	_, err = io.Copy(fout, fin)

	if err != nil {
		log.Fatal(err)
	}
}
