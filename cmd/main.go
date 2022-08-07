package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"github.com/bogem/id3v2"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func main() {

	searchCmd := flag.NewFlagSet("search", flag.ExitOnError)
	searchQuery := searchCmd.String("query", "Davidecks", "-query SEARCHSTRING")

	downloadCmd := flag.NewFlagSet("download", flag.ExitOnError)
	showPtr := downloadCmd.String("show", "Davidecks", "Show name")
	destDirPtr := downloadCmd.String("out-base-dir", "./music", "Location of your shows")
	progressPtr := downloadCmd.Bool("progress", false, "Print progress")

	if len(os.Args) < 2 {
		fmt.Println("expected 'download' or 'search' subcommands")
		os.Exit(1)
	}

	switch os.Args[1] {

	case "download":
		_ = downloadCmd.Parse(os.Args[2:])
		log.Println("subcommand 'download'")
		log.Println("  show:", *showPtr)
		log.Println("  out-base-dir:", *destDirPtr)
		log.Println("  progress:", *progressPtr)
		log.Println("  tail:", downloadCmd.Args())
		Download(*showPtr, *destDirPtr, progressPtr)
	case "search":
		_ = searchCmd.Parse(os.Args[2:])
		log.Printf("Searching for '%s' ...\n\n", *searchQuery)
		SearchBroadcastUrl(*searchQuery)
	default:
		log.Println("expected 'download' or 'search' subcommands")
		os.Exit(1)
	}
}

func Download(showSearch string, destDir string, progress *bool) {

	broadcastUrl := SearchBroadcastUrl(showSearch)
	broadcast := getBroadcast(broadcastUrl)

	show := createShow(broadcast)

	if len(show.Streams) > 0 {
		mp3Path := DownloadFile(
			getDownloadUrl(show),
			getOutputPath(destDir, show),
			getFileName(show),
			progress,
			10000)

		imagePath := saveImage(destDir, show, progress)
		writeId3Tag(mp3Path, imagePath, show)
	} else {
		log.Println("No streams found. Skipped download.")
	}

	log.Println("Done.")
}

func createShow(broadcast Broadcast) Show {
	return Show{
		Title:          trim(broadcast.Title),
		TitleSanitized: sanitize(trim(broadcast.Title)),
		Description:    removeHtmlTags(trim(broadcast.Subtitle)),
		BroadcastDay:   strconv.Itoa(broadcast.BroadcastDay),
		Images:         broadcast.Images,
		Streams:        broadcast.Streams,
		Year:           getYear(broadcast),
	}
}

func getBroadcast(broadcastUrl string) Broadcast {
	hitResponse, err := http.Get(broadcastUrl)
	logError(err)

	hitResponseData, err := ioutil.ReadAll(hitResponse.Body)
	logError(err)

	broadcast := Broadcast{}
	err = json.Unmarshal(hitResponseData, &broadcast)
	logError(err)
	log.Printf("Found broadcast info of show %s, broadcasted on %s", broadcast.Title, broadcast.StartISO)
	return broadcast
}

func SearchBroadcastUrl(searchTerm string) string {
	response, err := http.Get("https://audioapi.orf.at/fm4/api/json/current/search?q=" + url.QueryEscape(searchTerm))

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

	if len(parsedSearchResult.Hits) == 0 {
		log.Fatal("No search results! Aborting ...")
	}

	log.Printf("Found following show:\n")

	for _, hit := range parsedSearchResult.Hits {
		if hit.Data.Entity == "Broadcast" &&
			strings.Contains(hit.Data.Title, searchTerm) {
			fmt.Printf("\n   Name:            %s\n", hit.Data.Title)
			fmt.Printf("   ProgramKey:      %s\n", hit.Data.ProgramKey)
			fmt.Printf("   BroadcastDay:    %d\n", hit.Data.BroadcastDay)
			fmt.Printf("   Href:            %s\n", hit.Data.Href)
			fmt.Printf("   StartISO:        %s\n", hit.Data.StartISO)
			fmt.Printf("   Weekday:         %s\n", hit.Data.StartISO.Weekday())
			fmt.Printf("   Duration (min):  %d\n", int64(hit.Data.EndISO.Sub(hit.Data.StartISO).Minutes()))
			fmt.Printf("   Offset (hours):  %f\n\n", time.Since(hit.Data.StartISO).Hours()*-1)

			return hit.Data.Href
		}
	}
	return ""
}

func trim(str string) string {
	return strings.TrimSpace(str)
}

func removeHtmlTags(str string) string {
	r := regexp.MustCompile(`<.*?>`)
	return r.ReplaceAllString(str, "")
}

func sanitize(value string) string {
	return strings.Replace(strings.TrimSpace(value), " ", "_", -1)
}

func saveImage(path string, show Show, progress *bool) string {
	var imageUrl string
	if show.Images != nil && len(show.Images) > 0 {
		for _, v := range show.Images[0].Versions {
			if v.Width == 434 {
				imageUrl = v.Path
				return DownloadFile(imageUrl, getOutputPath(path, show), "cover.jpg", progress, 1000)
			}
		}
	} else {
		log.Println("No Cover images returned.")
	}
	return ""
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

	//log.Printf("Setting tag %s", tag)

	if imagePath != "" {
		artwork, err := ioutil.ReadFile(imagePath)
		if err != nil {
			log.Println("Error while reading artwork file", err)
		}

		pic := id3v2.PictureFrame{
			Encoding:    id3v2.EncodingUTF8,
			MimeType:    "image/jpeg",
			PictureType: id3v2.PTFrontCover,
			Description: "Front cover",
			Picture:     artwork,
		}
		tag.AddAttachedPicture(pic)

		log.Println("Attached cover.")
	} else {
		log.Println("No cover url provided. Skipped image tag.")
	}

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

func getYear(parsedItemResult Broadcast) string {
	return strconv.Itoa(parsedItemResult.StartISO.Year())
}

func getFileName(show Show) string {
	return fmt.Sprintf("%s_%s.mp3",
		show.TitleSanitized,
		show.BroadcastDay)
}

func getOutputPath(destDir string, show Show) string {
	return fmt.Sprintf("%s/%s/%s",
		destDir,
		show.TitleSanitized,
		show.Year)
}

func getDownloadUrl(show Show) string {
	var shoutcastBaseUrl = "https://loopstream01.apa.at/?channel=fm4&id="
	return shoutcastBaseUrl + show.Streams[0].LoopStreamID
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
