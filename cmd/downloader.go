package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

func PrintDownloadPercent(done chan int64, path string, total int64, interval int) {
	var stop = false
	file, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(file)

	for {
		select {
		case <-done:
			stop = true
		default:
			fi, err := file.Stat()
			if err != nil {
				log.Fatal(err)
			}

			size := fi.Size()
			if size == 0 {
				size = 1
			}

			var percent = float64(size) / float64(total) * 100
			fmt.Printf("%.0f", percent)
			fmt.Print("% ")
		}

		if stop {
			break
		}
		time.Sleep(time.Millisecond * time.Duration(interval))
	}
}

func DownloadFile(url string, outDir string, filename string, progressPtr *bool, interval int) string {
	log.Printf("Downloading file %s.\n", filename)

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

	start := time.Now()
	logError(err)

	defer func(out *os.File) {
		err := out.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(out)

	contentLength := -2.0
	i := 0
	for contentLength < 0 && i < 5 {
		contentLengthString := getContentLength(url, err)
		contentLength, _ = strconv.ParseFloat(contentLengthString, 64)
		i++
	}

	log.Printf("File size: %.2f Mb", contentLength/1024/1024)

	done := make(chan int64)

	if *progressPtr {
		go PrintDownloadPercent(done, path, int64(contentLength), interval)
	}

	resp, err := http.Get(url)
	logError(err)

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(resp.Body)

	n, err := io.Copy(out, resp.Body)
	logError(err)

	done <- n

	elapsed := time.Since(start)
	log.Printf("Download completed in %s.\n", elapsed)
	return path
}
