package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"github.com/bogem/id3v2"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
)

func main() {

	searchTermPtr := flag.String("show", "Davidecks", "A FM4 Show")
	destDirPtr := flag.String("out-base-dir", "/music", "Location of your shows")
	progressPtr := flag.Bool("progress", false, "Print progress")

	flag.Parse()

	response, err := http.Get("https://audioapi.orf.at/fm4/api/json/current/search?q=" + url.QueryEscape(*searchTermPtr))

	if err != nil {
		fmt.Print(err.Error())
		os.Exit(1)
	}

	responseData, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Fatal(err)
	}

	parsedSearchResult := SearchResult{}

	err = json.Unmarshal(responseData, &parsedSearchResult)
	if err != nil {
		log.Fatal(err)
	}

	for _, hit := range parsedSearchResult.Hits {
		response, err := http.Get(hit.Data.Href)
		logError(err)

		responseData, err := ioutil.ReadAll(response.Body)
		logError(err)

		parsedItemResult := ItemResult{}
		err = json.Unmarshal(responseData, &parsedItemResult)
		logError(err)

		show := Show{
			Title:          getTitle(hit),
			TitleSanitized: sanitize(getTitle(hit)),
			Description:    hit.Data.Subtitle,
			BroadcastDay:   strconv.Itoa(hit.Data.BroadcastDay),
			Images:         hit.Data.Images,
			Streams:        parsedItemResult.Streams,
			Year:           getYear(parsedItemResult),
		}

		if len(parsedItemResult.Streams) > 0 {
			mp3Path := DownloadFile(
				getDownloadUrl(show),
				getOutputPath(destDirPtr, show),
				getFileName(show),
				progressPtr,
				10000)

			imagePath := saveImage(destDirPtr, show, progressPtr)
			writeId3Tag(mp3Path, imagePath, show)
		}
	}
}

func getTitle(hit Hits) string {
	return strings.TrimSpace(hit.Data.Title)
}

func sanitize(value string) string {
	return strings.Replace(strings.TrimSpace(value), " ", "_", -1)
}

func saveImage(path *string, show Show, ptr *bool) string {
	var imageUrl string
	for _, v := range show.Images[0].Versions {
		if v.Width == 434 {
			imageUrl = v.Path
		}
	}
	return DownloadFile(imageUrl, getOutputPath(path, show), "cover.jpg", ptr, 500)
}

func writeId3Tag(mp3path string, imagePath string, show Show) {

	tag, err := id3v2.Open(mp3path, id3v2.Options{Parse: false})
	if err != nil {
		log.Fatal("Error while opening mp3 file: ", err)
	}

	tag.SetTitle(fmt.Sprintf("%s - %s", show.Title, show.BroadcastDay))
	tag.SetAlbum(show.Year)
	tag.SetArtist(show.Title)
	tag.SetYear(show.Year)

	artwork, err := ioutil.ReadFile(imagePath)
	if err != nil {
		log.Fatal("Error while reading artwork file", err)
	}

	pic := id3v2.PictureFrame{
		Encoding:    id3v2.EncodingUTF8,
		MimeType:    "image/jpeg",
		PictureType: id3v2.PTFrontCover,
		Description: "Front cover",
		Picture:     artwork,
	}
	tag.AddAttachedPicture(pic)

	textFrame := id3v2.TextFrame{
		Encoding: id3v2.EncodingUTF8,
		Text:     show.Title,
	}
	tag.AddFrame(tag.CommonID("TPE2"), textFrame)

	defer func(tag *id3v2.Tag) {
		err := tag.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(tag)

	err = tag.Save()
	logError(err)
}

func getYear(parsedItemResult ItemResult) string {
	return strconv.Itoa(parsedItemResult.StartISO.Year())
}

func getFileName(show Show) string {
	return fmt.Sprintf("%s_%s.mp3",
		show.TitleSanitized,
		show.BroadcastDay)
}

func getOutputPath(destDirPtr *string, show Show) string {
	return fmt.Sprintf("%s/%s/%s",
		*destDirPtr,
		show.TitleSanitized,
		show.Year)
}

func getDownloadUrl(show Show) string {
	var shoutcastBaseUrl = "https://loopstream01.apa.at/?channel=fm4&id="
	return shoutcastBaseUrl + show.Streams[0].LoopStreamID
}

func getContentLength(url string, err error) string {
	headResp, err := http.Head(url)

	logError(err)

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(headResp.Body)

	var contentLength = headResp.Header.Get("Content-Length")
	return contentLength
}

func logError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func makeDirectoryIfNotExisting(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return os.MkdirAll(path, os.ModeDir|0755)
	}
	return nil
}

func fileExists(name string) (bool, error) {
	_, err := os.Stat(name)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	return false, err
}
