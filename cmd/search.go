package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

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
