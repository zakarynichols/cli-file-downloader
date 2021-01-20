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
		TargetPath:    "bike.png",
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
	r, err := http.NewRequest(
		("HEAD"),
		d.Url,
		nil,
	)
	if err != nil {
		return err
	}
	// Set HTTP headers
	r.Header.Set("User-Agent", "File Downloader")

	// Make the HTTP request
	resp, err := http.DefaultClient.Do(r)
	if err != nil {
		return err
	}
	fmt.Printf("Response Status Code: %v\n", resp.StatusCode)
	if resp.StatusCode > 299 {
		return errors.New(fmt.Sprintf("Can't process. Response is %v", resp.StatusCode))
	}
	// Log Headers
	for name, values := range resp.Header {
		fmt.Println("name:", name, "values:", values)
		// Loop over all values for the name.
		for _, value := range values {
			fmt.Println("name:", name, "value:", value)
		}
	}

	// Get size of download in bytes
	size, err := strconv.Atoi(resp.Header.Get("Content-Length"))
	if err != nil {
		return err
	}

	var sections = make([][2]int, d.TotalSections)
	eachSize := size / d.TotalSections

	for i := range sections {
		if i == 0 {
			// set first byte to zero
			sections[i][0] = 0
		} else {
			// set the first byte in the section to the last byte of the
			// previous one + 1 to account for the increment
			sections[i][0] = sections[i-1][1] + 1
		}
		if i < d.TotalSections-1 {
			// set the last bytes in the section equal to
			// the first bytes
			sections[i][1] = sections[i][0] + eachSize
		} else {
			// set the last byte in the section equal to the
			// total size
			sections[i][1] = size
		}
	}

	fmt.Printf("sections %v", sections)

	return nil
}
