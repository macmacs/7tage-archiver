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
	"time"
)

func main() {

	searchTermPtr := flag.String("show", "Davidecks", "A FM4 Show")
	destDirPtr := flag.String("out-base-dir", "/path/to/save/all/your/shows/to/", "FM4/Sendungen")
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

	hrefs := []string{}

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

	fmt.Println(hrefs)

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
	defer tag.Close()

	tag.SetTitle(show.Title)
	tag.SetAlbum(show.Year)
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
	tag.AddFrame(tag.CommonID("albumartist"), textFrame)

	time.Sleep(time.Millisecond * 500)

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

func PrintDownloadPercent(done chan int64, path string, total int64, interval int) {
	var stop bool = false
	file, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	for {
		select {
		case <-done:
			stop = true
		default:
			fi, err := file.Stat()
			if err != nil {
				log.Fatal(err)
			}

			size := fi.Size()
			if size == 0 {
				size = 1
			}

			var percent float64 = float64(size) / float64(total) * 100
			fmt.Printf("%.0f", percent)
			fmt.Print("% ")
		}

		if stop {
			break
		}
		time.Sleep(time.Millisecond * time.Duration(interval))
	}
}

func DownloadFile(url string, outDir string, filename string, progressPtr *bool, interval int) string {

	log.Printf("Downloading file %s.", filename)

	err := makeDirectoryIfNotExisting(outDir)
	logErrorAndExit(err, 7)

	var path = outDir + "/" + filename

	fileIsExisting, err := fileExists(path)
	logError(err)

	if fileIsExisting {
		log.Println("File " + path + " already exists. Skipping download.")
		return path
	}

	out, err := os.Create(path)

	start := time.Now()
	logError(err)

	defer out.Close()
	contentLength := -2.0
	i := 0
	for contentLength < 0 && i < 5 {
		contentLengthString := getContentLength(url, err)
		contentLength, _ = strconv.ParseFloat(contentLengthString, 64)
		i++
	}

	done := make(chan int64)

	if *progressPtr {
		go PrintDownloadPercent(done, path, int64(contentLength), interval)
	}

	resp, err := http.Get(url)

	logErrorAndExit(err, 4)

	defer resp.Body.Close()

	n, err := io.Copy(out, resp.Body)

	logErrorAndExit(err, 5)

	done <- n

	elapsed := time.Since(start)
	log.Println("Download completed in %s.", elapsed)
	return path
}

func getContentLength(url string, err error) string {
	headResp, err := http.Head(url)

	logErrorAndExit(err, 2)

	defer headResp.Body.Close()

	var contentLength = headResp.Header.Get("Content-Length")
	log.Printf("File size: %s", contentLength)
	return contentLength
}

func logError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func logErrorAndExit(err error, code int) {
	if err != nil {
		log.Fatal(err)
		os.Exit(code)
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
