package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
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

	downloadCmd := flag.NewFlagSet("download", flag.ExitOnError)
	showPtr := downloadCmd.String("show", "4DD", "Show name")
	destDirPtr := downloadCmd.String("out-base-dir", "./music", "Location of your shows")

	if len(os.Args) < 2 {
		fmt.Println("expected 'download' or 'printshowids' subcommands")
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
	case "printshowids":
		printShowIDs()
	default:
		log.Println("expected 'download' or 'search' subcommands")
		os.Exit(1)
	}
}

func Download(showId string, destDir string) {

	broadcastUrl := CreateBroadcastUrl(showId)
	broadcast := getBroadcast(broadcastUrl)

	show := createShow(broadcast)

	if len(show.Streams) > 0 {
		mp3Path := DownloadFile(
			getDownloadUrl(show),
			getOutputPath(destDir, show),
			getFileName(show))

		imagePath := saveImage(destDir, show)
		writeId3Tag(mp3Path, imagePath, show)
	} else {
		log.Println("No streams found. Skipped download.")
	}

	log.Println("Done.")
}

func CreateBroadcastUrl(showId string) string {
	return fmt.Sprintf("https://audioapi.orf.at/fm4/api/json/4.0/broadcast/%s", showId)
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

	hitResponseData, err := io.ReadAll(hitResponse.Body)
	logError(err)

	broadcast := Broadcast{}
	err = json.Unmarshal(hitResponseData, &broadcast)
	logError(err)
	log.Printf("Found broadcast info of show %s, broadcasted on %s", broadcast.Title, broadcast.StartISO)
	return broadcast
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

func printShowIDs() {
	showIds := map[string]string{
		"4HOP": "House Of Pain",
		"4UL":  "Unlimited",
		"4SSU": "Sunny Side Up",
		"4CH":  "Charts",
		"4ZS":  "Zimmerservice",
		"4JZ":  "Jugend-Zimmer",
		"4DD":  "Davidecks",
		"4IS":  "Im Sumpf",
		"4HE":  "Heartbeat",
		"4HS":  "High Spirits",
		"4TV":  "Tribe Vibes",
		"4SH":  "Salon Helga",
		"4GL":  "Graue Lagune",
		"4PH":  "Fivas Ponyhof",
		"4CZ":  "Chez Hermez",
		"4BT":  "Bonustrack",
		"4PX":  "Project X",
		"4LB":  "La Boum de Luxe",
		"4SS":  "Swound Sound System",
		"4LR":  "Liquid Radio",
		"4DKM": "Digital Konfusion",
		"4SOP": "Soundpark",
	}
	for n, m := range showIds {
		fmt.Println(m, ": ", n)
	}
}
