package main

import "time"

type SearchResult struct {
	Took       int  `json:"took"`
	IsTimedOut bool `json:"isTimedOut"`
	Length     int  `json:"length"`
	Total      int  `json:"total"`
	Hits       []struct {
		Data struct {
			Href                 string    `json:"href"`
			Station              string    `json:"station"`
			Entity               string    `json:"entity"`
			ID                   int       `json:"id"`
			BroadcastDay         int       `json:"broadcastDay"`
			ProgramKey           string    `json:"programKey"`
			Program              string    `json:"program"`
			Title                string    `json:"title"`
			Subtitle             string    `json:"subtitle"`
			Ressort              string    `json:"ressort"`
			State                string    `json:"state"`
			IsOnDemand           bool      `json:"isOnDemand"`
			IsGeoProtected       bool      `json:"isGeoProtected"`
			IsAdFree             bool      `json:"isAdFree"`
			Start                int64     `json:"start"`
			StartISO             time.Time `json:"startISO"`
			StartOffset          int       `json:"startOffset"`
			ScheduledStart       int64     `json:"scheduledStart"`
			ScheduledStartISO    time.Time `json:"scheduledStartISO"`
			ScheduledStartOffset int       `json:"scheduledStartOffset"`
			End                  int64     `json:"end"`
			EndISO               time.Time `json:"endISO"`
			EndOffset            int       `json:"endOffset"`
			ScheduledEnd         int64     `json:"scheduledEnd"`
			ScheduledEndISO      time.Time `json:"scheduledEndISO"`
			ScheduledEndOffset   int       `json:"scheduledEndOffset"`
			NiceTime             int64     `json:"niceTime"`
			NiceTimeISO          time.Time `json:"niceTimeISO"`
			NiceTimeOffset       int       `json:"niceTimeOffset"`
			Description          string    `json:"description"`
			PressRelease         string    `json:"pressRelease"`
			Moderator            string    `json:"moderator"`
			URL                  string    `json:"url"`
			Images               []struct {
				Alt      string `json:"alt"`
				Mode     string `json:"mode"`
				Text     string `json:"text"`
				Category string `json:"category"`
				HashCode int    `json:"hashCode"`
				Versions []struct {
					Path     string `json:"path"`
					Width    int    `json:"width"`
					HashCode int    `json:"hashCode"`
				} `json:"versions"`
				Copyright string `json:"copyright"`
			} `json:"images"`
			Tags    []interface{} `json:"tags"`
			Oe1Tags []interface{} `json:"oe1tags"`
		} `json:"data"`
		Highlights struct {
			Title []string `json:"title"`
		} `json:"highlights"`
	} `json:"hits"`
	Suggest []struct {
		Text        string  `json:"text"`
		Highlighted string  `json:"highlighted"`
		Score       float64 `json:"score"`
	} `json:"suggest"`
}
