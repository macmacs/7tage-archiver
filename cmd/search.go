package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

func SearchBroadcastUrls(searchQuery string) []string {

	fmt.Printf("   Searching for '%s' ...\n\n", searchQuery)

	parsedSearchResult, err := getSearchResults(searchQuery)
	logError(err)

	fmt.Printf("   Found following show:\n")

	var result []string

	for _, hit := range parsedSearchResult.Hits {
		if hit.Data.Entity == "Broadcast" &&
			strings.Contains(
				strings.ToLower(hit.Data.Title),
				strings.ToLower(searchQuery)) {
			fmt.Printf("\n   Name:            %s\n", hit.Data.Title)
			fmt.Printf("   ProgramKey:      %s\n", hit.Data.ProgramKey)
			fmt.Printf("   BroadcastDay:    %d\n", hit.Data.BroadcastDay)
			fmt.Printf("   Href:            %s\n", hit.Data.Href)
			fmt.Printf("   StartISO:        %s\n", hit.Data.StartISO)
			fmt.Printf("   Weekday:         %s\n", hit.Data.StartISO.Weekday())
			fmt.Printf("   Duration (min):  %d\n", int64(hit.Data.EndISO.Sub(hit.Data.StartISO).Minutes()))
			fmt.Printf("   Offset (hours):  %f\n\n", time.Since(hit.Data.StartISO).Hours()*-1)

			result = append(result, hit.Data.Href)
		}
	}
	return result
}

func getSearchResults(searchTerm string) (SearchResult, error) {
	response, err := http.Get("https://audioapi.orf.at/fm4/api/json/current/search?q=" + url.QueryEscape(searchTerm))
	responseData, err := io.ReadAll(response.Body)

	parsedSearchResult := SearchResult{}
	err = json.Unmarshal(responseData, &parsedSearchResult)

	if len(parsedSearchResult.Hits) == 0 {
		if len(parsedSearchResult.Suggest) > 0 {
			fmt.Printf("   No results found for %s.\n\n", searchTerm)
			fmt.Printf("   But did you mean ")
			for i, s := range parsedSearchResult.Suggest {
				if i < len(parsedSearchResult.Suggest)-1 {
					fmt.Printf("'%s' or ", s.Text)
				} else {

					fmt.Printf("'%s'?\n\n", s.Text)
				}
			}
			log.Println("Please try again.")
		} else {
			log.Println("No search results!")
		}
	}
	return parsedSearchResult, err
}
