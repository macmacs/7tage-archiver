package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	id3 "github.com/mikkyang/id3-go"
	v2 "github.com/mikkyang/id3-go/v2"
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

type Show struct {
	Title       string
	Description string
	Start       string
}

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

	for i, h := range parsedSearchResult.Hits {
		fmt.Println(i, h.Data.Href)
		response, err := http.Get(h.Data.Href)
		if err != nil {
			fmt.Print(err.Error())
			os.Exit(1)
		}
		responseData, err := ioutil.ReadAll(response.Body)
		if err != nil {
			log.Fatal(err)
		}
		parsedItemResult := ItemResult{}
		err = json.Unmarshal(responseData, &parsedItemResult)
		if err != nil {
			log.Fatal(err)
		}

		if len(parsedItemResult.Streams) > 0 {
			var shoutcastBaseUrl = "https://loopstream01.apa.at/?channel=fm4&id="
			path := DownloadFile(getDownloadUrl(shoutcastBaseUrl, parsedItemResult),
				getOutputPath(destDirPtr, parsedItemResult),
				getFileName(h, parsedItemResult),
				progressPtr)

			writeId3Tag(path, parsedItemResult)
			saveImage(destDirPtr, parsedItemResult, progressPtr)
		}

	}

	fmt.Println(hrefs)

}

func saveImage(path *string, result ItemResult, ptr *bool) {
	var imageUrl string
	for _, v := range result.Images[0].Versions {
		if v.Width == 434 {
			imageUrl = v.Path
		}
	}
	DownloadFile(imageUrl, getOutputPath(path, result), "cover.jpg", ptr)
}

func writeId3Tag(path string, parsedItemResult ItemResult) {

	log.Println("Tagging mp4 file ...")
	mp3File, err := id3.Open(path)
	logError(err)

	mp3File.SetArtist(parsedItemResult.Title)
	mp3File.SetAlbum(getYear(parsedItemResult))
	ft := v2.V23FrameTypeMap["TPE2"]
	textFrame := v2.NewTextFrame(ft, parsedItemResult.Title)
	mp3File.AddFrames(textFrame)

	defer mp3File.Close()
}

func getYear(parsedItemResult ItemResult) string {
	return strconv.Itoa(parsedItemResult.StartISO.Year())
}

func getFileName(h Hits, parsedItemResult ItemResult) string {
	return fmt.Sprintf("%s_%s.mp3",
		strings.Replace(parsedItemResult.Title, " ", "_", -1),
		strconv.Itoa(parsedItemResult.BroadcastDay))
}

func getOutputPath(destDirPtr *string, parsedItemResult ItemResult) string {
	return fmt.Sprintf("%s/%s/%s",
		*destDirPtr,
		strings.Replace(parsedItemResult.Title, " ", "_", -1),
		getYear(parsedItemResult))
}

func getDownloadUrl(shoutcastBaseUrl string, parsedItemResult ItemResult) string {
	return shoutcastBaseUrl + parsedItemResult.Streams[0].LoopStreamID
}

func PrintDownloadPercent(done chan int64, path string, total int64) {
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
		time.Sleep(time.Second * 10)
	}
}

func DownloadFile(url string, outDir string, filename string, progressPtr *bool) string {

	log.Printf("Downloading file %s.", filename)

	makeDirectoryIfNotExisting(outDir)

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

	headResp, err := http.Head(url)

	logErrorAndExit(err, 2)

	defer headResp.Body.Close()

	size, err := strconv.Atoi(headResp.Header.Get("Content-Length"))

	logErrorAndExit(err, 3)

	done := make(chan int64)

	if *progressPtr {
		go PrintDownloadPercent(done, path, int64(size))
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
		return os.Mkdir(path, os.ModeDir|0755)
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
