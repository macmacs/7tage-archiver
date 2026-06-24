package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	// YYYY-MM-DD: 2022-03-23
	YYYYMMDD = "2006-01-02"
	// 24h hh:mm:ss: 14:23:20
	HHMMSS24h = "15:04:05"
)

func main() {

	log.SetFlags(log.Lmsgprefix)
	log.SetPrefix(time.Now().Format(YYYYMMDD+" "+HHMMSS24h) + " >   ")

	searchCmd := flag.NewFlagSet("search", flag.ExitOnError)
	searchQuery := searchCmd.String("query", "Davidecks", "-query SEARCHSTRING")

	downloadCmd := flag.NewFlagSet("download", flag.ExitOnError)
	showPtr := downloadCmd.String("show", "Davidecks", "Show name")
	destDirPtr := downloadCmd.String("out-base-dir", "./music", "Location of your shows")

	urlCmd := flag.NewFlagSet("url", flag.ExitOnError)
	destDirUrlPtr := urlCmd.String("out-base-dir", "./music", "Location of your shows")

	if len(os.Args) < 2 {
		fmt.Println("expected 'download', 'url' or 'search' subcommands")
		os.Exit(1)
	}

	switch os.Args[1] {

	case "download":
		_ = downloadCmd.Parse(os.Args[2:])
		log.Println("subcommand 'download'")
		log.Println("  show:", *showPtr)
		log.Println("  out-base-dir:", *destDirPtr)
		log.Println("  tail:", downloadCmd.Args())
		Download(*showPtr, *destDirPtr)
	case "url":
		_ = urlCmd.Parse(os.Args[2:])
		if len(urlCmd.Args()) < 1 {
			log.Fatal("subcommand 'url' expects a sound.orf.at Sendung URL " +
				"(https://sound.orf.at/radio/fm4/sendung/42628/davidecks) " +
				"or a programKey (e.g. 4DD)")
		}
		showRef := urlCmd.Arg(0)
		log.Println("subcommand 'url'")
		log.Println("  show:", showRef)
		log.Println("  out-base-dir:", *destDirUrlPtr)
		DownloadByUrl(showRef, *destDirUrlPtr)
	case "search":
		_ = searchCmd.Parse(os.Args[2:])
		SearchBroadcastUrls(*searchQuery)
	default:
		log.Println("expected 'download', 'url' or 'search' subcommands")
		os.Exit(1)
	}
}

func Download(showSearch string, destDir string) {

	broadcastUrls := SearchBroadcastUrls(showSearch)

	downloadBroadcasts(broadcastUrls, destDir)
}

// DownloadByUrl downloads all available episodes of the show referenced by
// either a sound.orf.at Sendung URL (e.g.
// https://sound.orf.at/radio/fm4/sendung/42628/davidecks) or a stable
// programKey (e.g. "4DD"). Every episode still in the 30-day on-demand window
// is fetched. Prefer the programKey for recurring downloads: a URL's episode
// id ages out of the window after 30 days, a programKey does not.
func DownloadByUrl(showRef string, destDir string) {

	broadcastUrls := ResolveBroadcastUrls(showRef)

	downloadBroadcasts(broadcastUrls, destDir)
}

func downloadBroadcasts(broadcastUrls []string, destDir string) {

	for _, broadcastUrl := range broadcastUrls {

		broadcast := getBroadcast(broadcastUrl)

		show := createShow(broadcast)

		if len(show.Streams) > 0 {
			outDir := getOutputPath(destDir, show)
			fileName := getFileName(show)

			var mp3Path string
			if segs := contentSegments(show); len(segs) > 0 {
				urls := make([]string, len(segs))
				for i, seg := range segs {
					urls[i] = getSegmentUrl(show, seg)
				}
				mp3Path = DownloadFileSegments(urls, outDir, fileName)
			} else {
				mp3Path = DownloadFile(getDownloadUrl(show), outDir, fileName)
			}

			imagePath := saveImage(destDir, show)
			writeId3Tag(mp3Path, imagePath, show)
		} else {
			log.Println("No streams found. Skipped download.")
		}
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
		Items:          broadcast.Items,
		Year:           getYear(broadcast),
	}
}

func getBroadcast(broadcastUrl string) Broadcast {
	broadcastUrl = ensureItemsParam(broadcastUrl)
	hitResponse, err := http.Get(broadcastUrl)
	logError(err)

	hitResponseData, err := io.ReadAll(hitResponse.Body)
	logError(err)

	// The v5.0 API wraps the broadcast inside {"timezoneOffset":...,"payload":{...}},
	// unlike the old v4.0 hrefs where the broadcast sat at the top level. Unwrap
	// here and map through the broadcastV5 intermediate so the consumed Broadcast
	// stays in int64-millisecond form for trim.go.
	var wrapper struct {
		Payload broadcastV5 `json:"payload"`
	}
	err = json.Unmarshal(hitResponseData, &wrapper)
	logError(err)

	broadcast := wrapper.Payload.toBroadcast()
	log.Printf("Found broadcast info of show %s, broadcasted on %s", broadcast.Title, broadcast.StartISO)
	return broadcast
}

// ensureItemsParam guarantees the v5.0 broadcast endpoint returns its sub-items.
// The bare broadcast/{id} href returns an empty items array, which would leave
// trim.go with nothing to cut (the news block and ads would slip through). The
// API's ?items=N flag is a boolean "include items" switch rather than a cap -
// any positive N returns the full set - so a generous value is safe.
func ensureItemsParam(broadcastUrl string) string {
	u, err := url.Parse(broadcastUrl)
	if err != nil {
		return broadcastUrl
	}
	q := u.Query()
	if q.Get("items") == "" {
		q.Set("items", "1000")
	}
	u.RawQuery = q.Encode()
	return u.String()
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

func saveImage(path string, show Show) string {
	var imageUrl string
	if show.Images != nil && len(show.Images) > 0 {
		for _, v := range show.Images[0].Versions {
			if v.Width == 434 {
				imageUrl = v.Path
				return DownloadFile(imageUrl, getOutputPath(path, show), "cover.jpg")
			}
		}
	} else {
		log.Println("No Cover images returned.")
	}
	return ""
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
	// Canonical sound.orf.at stream URL for the loopStreamId, read from the
	// v5.0 payload's urls/uriTemplates (cleaned of the RFC-6570 token suffix
	// and pre-filled offset range by toBroadcast). Uses the per-station host
	// (e.g. loopstreamfm4.apa.at) instead of the legacy global loopstream01.
	return show.Streams[0].Progressive
}

// getSegmentUrl extends the stream download URL with the loopstream range
// params so the server returns only the given millisecond slice.
func getSegmentUrl(show Show, seg segment) string {
	return fmt.Sprintf("%s&offset=%d&offsetende=%d",
		getDownloadUrl(show), seg.offset, seg.offsetEnd)
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
