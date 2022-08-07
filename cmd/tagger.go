package main

import (
	"fmt"
	"github.com/bogem/id3v2"
	"io/ioutil"
	"log"
)

func writeId3Tag(mp3path string, imagePath string, show Show) {

	tag, err := id3v2.Open(mp3path, id3v2.Options{Parse: false})
	if err != nil {
		log.Fatal("Error while opening mp3 file: ", err)
	}

	tag.SetTitle(fmt.Sprintf("%s - %s", show.Title, show.BroadcastDay))
	tag.SetAlbum(show.Year)
	tag.SetArtist(show.Title)
	tag.SetYear(show.Year)

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
