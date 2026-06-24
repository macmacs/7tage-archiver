package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
)

// soundUrlPattern matches sound.orf.at Sendung URLs of the form
// https://sound.orf.at/radio/fm4/sendung/<broadcastId>[/<titleSlug>]
// Only the fm4 station is supported (matches the rest of this FM4-focused tool).
var soundUrlPattern = regexp.MustCompile(`^https?://sound\.orf\.at/radio/fm4/sendung/(\d+)(?:/[\w-]+)?$`)

// programKeyPattern matches a bare fm4 programKey (e.g. "4DD", "4DKM"). Unlike
// the episode id in a Sendung URL, a programKey is stable across episodes, so
// it is the right identifier for a recurring (e.g. cron-driven) download.
var programKeyPattern = regexp.MustCompile(`^[0-9A-Z]{2,8}$`)

// ResolveBroadcastUrls resolves a show reference into the broadcast href URLs
// of all its episodes still inside the 30-day on-demand window. The reference
// is either a sound.orf.at Sendung URL (which points at a single episode whose
// programKey is resolved first) or a bare, stable programKey (e.g. "4DD").
func ResolveBroadcastUrls(showRef string) []string {
	if matches := soundUrlPattern.FindStringSubmatch(showRef); matches != nil {
		broadcastId := matches[1]
		programKey := getProgramKey(broadcastId)
		log.Printf("Resolved show with programKey %s from URL", programKey)
		log.Println("Found following show:")
		return getProgramEpisodes(programKey)
	}

	if programKeyPattern.MatchString(showRef) {
		log.Printf("Using programKey %s directly", showRef)
		log.Println("Found following show:")
		return getProgramEpisodes(showRef)
	}

	log.Fatalf("expected a sound.orf.at Sendung URL "+
		"('https://sound.orf.at/radio/fm4/sendung/<id>[/<slug>]') or a programKey "+
		"(e.g. '4DD'), got: %s", showRef)
	return nil
}

func logUnexpectedStatus(response *http.Response, url string) {
	if response.StatusCode != http.StatusOK {
		log.Fatalf("GET %s returned %s", url, response.Status)
	}
}

// getProgramKey fetches broadcast/{id} on the v5.0 API and returns the
// show's programKey (e.g. "4DD" for the Davidecks broadcast id 42628).
// Note: the v5.0 API wraps the broadcast inside {"timezoneOffset":...,"payload":{...}},
// unlike the v4.0 search-href URLs returned by the 'search'/'download' flow
// where the broadcast sits at the top level. We unwrap here so the rest of
// the pipeline (getBroadcast over v4.0 hrefs) stays unchanged.
func getProgramKey(broadcastId string) string {
	broadcastUrl := fmt.Sprintf("https://audioapi.orf.at/fm4/api/json/5.0/broadcast/%s", broadcastId)

	response, err := http.Get(broadcastUrl)
	logError(err)
	logUnexpectedStatus(response, broadcastUrl)
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			logError(err)
		}
	}(response.Body)

	responseData, err := io.ReadAll(response.Body)
	logError(err)

	var wrapper struct {
		Payload struct {
			ProgramKey string `json:"programKey"`
		} `json:"payload"`
	}
	err = json.Unmarshal(responseData, &wrapper)
	logError(err)

	return wrapper.Payload.ProgramKey
}

// getProgramEpisodes calls broadcasts/program/{programKey} on the v5.0 API
// and returns v4.0-style broadcast href URLs (one per episode) that the
// existing getBroadcast/DownloadFile pipeline consumes unchanged. The v5.0
// listing endpoint returns each episode's programKey and broadcastDay, which
// are the components of the (stable) v4.0 broadcast URL
// audioapi.orf.at/{station}/api/json/4.0/broadcast/{programKey}/{broadcastDay}.
// Going via v4.0 hrefs keeps the rest of this tool format-agnostic and avoids
// needing to unwrap the v5.0 {'payload': {...}} envelope in getBroadcast.
func getProgramEpisodes(programKey string) []string {
	url := fmt.Sprintf("https://audioapi.orf.at/fm4/api/json/5.0/broadcasts/program/%s", programKey)

	response, err := http.Get(url)
	logError(err)
	logUnexpectedStatus(response, url)
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			logError(err)
		}
	}(response.Body)

	responseData, err := io.ReadAll(response.Body)
	logError(err)

	// The endpoint returns {"timezoneOffset": ..., "payload": [ <broadcast>, ... ]}
	var wrapper struct {
		Payload []struct {
			BroadcastDay int    `json:"broadcastDay"`
			Title        string `json:"title"`
			ProgramKey   string `json:"programKey"`
		} `json:"payload"`
	}
	err = json.Unmarshal(responseData, &wrapper)
	logError(err)

	if len(wrapper.Payload) == 0 {
		log.Println("No episodes found for this show.")
		return nil
	}

	var urls []string
	for _, episode := range wrapper.Payload {
		log.Println("")
		log.Printf("Name:            %s", episode.Title)
		log.Printf("ProgramKey:      %s", episode.ProgramKey)
		log.Printf("BroadcastDay:    %d", episode.BroadcastDay)
		href := fmt.Sprintf("https://audioapi.orf.at/fm4/api/json/4.0/broadcast/%s/%d",
			episode.ProgramKey, episode.BroadcastDay)
		log.Printf("Href:            %s", href)
		urls = append(urls, href)
	}
	log.Println("")
	return urls
}
