package main

type Show struct {
	Title          string
	TitleSanitized string
	Description    string
	BroadcastDay   string
	Year           string
	Images         []Images
	Streams        []Streams
}
