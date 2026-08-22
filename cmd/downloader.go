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
	// The loopstream backend expects the sound.orf.at frontend as the origin.
	// The v5.0 urls.progressive also embeds referer=sound.orf.at as a query
	// param, but the header is the one the server actually honours.
	req.Header.Set("Referer", "https://sound.orf.at/")
	resp, err := http.DefaultClient.Do(req)
	logError(err)

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			logError(err)
		}
	}(resp.Body)

	var w io.Writer = out
	if isTerminal(os.Stderr) {
		// Skip the bar when stderr is not a terminal (e.g. cron): the
		// carriage-return rendering would only clutter the captured output.
		bar := progressbar.DefaultBytes(
			resp.ContentLength,
			"Downloading",
		)
		w = io.MultiWriter(out, bar)
	}

	_, err = io.Copy(w, resp.Body)
	logError(err)
}

// isTerminal reports whether f is attached to a terminal, in which case a
// progress bar can render; otherwise (pipes, cron captures, log files) it
// returns false.
func isTerminal(f *os.File) bool {
	info, err := f.Stat()
	if err != nil {
		return false
	}
	return info.Mode()&os.ModeCharDevice != 0
}
