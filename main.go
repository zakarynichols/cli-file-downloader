package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"
)

type Download struct {
	Url           string
	TargetPath    string
	TotalSections int
}

func main() {
	startTime := time.Now()
	d := Download{
		Url:           "https://unsplash.com/photos/-zqe55fIOq8/download",
		TargetPath:    "speedtest",
		TotalSections: 10,
	}
	err := d.Do()
	if err != nil {
		log.Printf("An error occured while downloading the file: %s\n", err)
	}
	fmt.Printf("Download completed in %v seconds\n", time.Now().Sub(startTime).Seconds())
}

func (d Download) Do() error {
	fmt.Println("Checking URL...")
	// Create new HTTP request
	r, err := d.getNewRequest("HEAD")
	if err != nil {
		return err
	}

	// Make the HTTP request
	resp, err := http.DefaultClient.Do(r)
	if err != nil {
		return err
	}
	fmt.Printf("Got %v\n", resp.StatusCode)
	if resp.StatusCode > 299 {
		return errors.New(fmt.Sprintf("Can't process. Response is %v", resp.StatusCode))
	}

	// Get size of download in bytes
	size, err := strconv.Atoi(resp.Header.Get("Content-Length"))
	if err != nil {
		return err
	}
	fmt.Printf("Size is %v bytes\n", size)

	// Log Headers
	for name, values := range r.Header {
		// Loop over all values for the name.
		for _, value := range values {
			fmt.Println(name, value)
		}
	}

	return nil
}

func (d Download) getNewRequest(method string) (*http.Request, error) {
	// Create a new HTTP request
	r, err := http.NewRequest(
		method,
		d.Url,
		nil,
	)

	if err != nil {
		return nil, err
	}
	// Set HTTP headers
	r.Header.Set("User-Agent", "File Downloader")

	return r, nil
}
