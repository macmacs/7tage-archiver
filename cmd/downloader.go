package main

import (
	"github.com/schollz/progressbar/v3"
	"io"
	"log"
	"net/http"
	"os"
)

func DownloadFile(url string, outDir string, filename string) string {
	log.Printf("Downloading file %s from %s.\n", filename, url)

	err := makeDirectoryIfNotExisting(outDir)
	logError(err)

	var path = outDir + "/" + filename

	fileIsExisting, err := fileExists(path)
	logError(err)

	if fileIsExisting {
		log.Println("File " + path + " already exists. Skipping download.")
		return path
	}

	out, err := os.Create(path)

	logError(err)

	req, err := http.NewRequest("GET", url, nil)
	logError(err)
	resp, err := http.DefaultClient.Do(req)
	logError(err)

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			logError(err)
		}
	}(resp.Body)

	defer func(out *os.File) {
		err := out.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(out)

	bar := progressbar.DefaultBytes(
		resp.ContentLength,
		"Downloading",
	)

	_, err = io.Copy(io.MultiWriter(out, bar), resp.Body)
	logError(err)

	return path
}
