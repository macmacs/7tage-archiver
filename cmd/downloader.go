package main

import (
	"github.com/schollz/progressbar/v3"
	"io"
	"log"
	"net/http"
	"os"
)

func DownloadFile(url string, outDir string, filename string) string {
	return DownloadFileSegments([]string{url}, outDir, filename)
}

// DownloadFileSegments downloads each url in order and writes them, joined, to
// a single file. With one url it behaves like a plain download; with several it
// concatenates the slices (used to stitch together the show content around the
// removed news and ad segments).
func DownloadFileSegments(urls []string, outDir string, filename string) string {
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

	defer func(out *os.File) {
		err := out.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(out)

	for _, url := range urls {
		downloadSegment(url, filename, out)
	}

	return path
}

func downloadSegment(url string, filename string, out io.Writer) {
	log.Printf("Downloading file %s from %s.\n", filename, url)

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

	bar := progressbar.DefaultBytes(
		resp.ContentLength,
		"Downloading",
	)

	_, err = io.Copy(io.MultiWriter(out, bar), resp.Body)
	logError(err)
}
