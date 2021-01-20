package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"
)

type Download struct {
	Url        string
	TargetPath string
	Chunks     int
}

func main() {
	startTime := time.Now()
	d := Download{
		Url:        "https://unsplash.com/photos/-zqe55fIOq8/download",
		TargetPath: "bike.png",
		Chunks:     10,
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

	var chunks = make([][2]int, d.Chunks)
	eachSize := size / d.Chunks

	for i := range chunks {
		if i == 0 {
			// set first byte to zero
			chunks[i][0] = 0
		} else {
			// set the first byte in the section to the last byte of the
			// previous one + 1 to account for the increment
			chunks[i][0] = chunks[i-1][1] + 1
		}
		if i < d.Chunks-1 {
			// set the last bytes in the section equal to
			// the first bytes
			chunks[i][1] = chunks[i][0] + eachSize
		} else {
			// set the last byte in the section equal to the
			// total size
			chunks[i][1] = size
		}
	}

	fmt.Printf("chunks %v", chunks)

	var wg sync.WaitGroup

	for i, s := range chunks {
		wg.Add(1)
		go func(i int, s [2]int) {
			defer wg.Done()
			err := d.downloadChunk(i, s)
			if err != nil {
				panic(err)
			}
		}(i, s)
	}
	wg.Wait()

	return nil
}

func (d Download) downloadChunk(i int, c [2]int) error {
	r, err := http.NewRequest(
		"GET",
		d.Url,
		nil,
	)
	if err != nil {
		return err
	}
	// Set the Range Headers to our chunks of bytes
	// that we'll pass in our goroutine
	r.Header.Set("Range", fmt.Sprintf("bytes=%v-%v", c[0], c[1]))
	resp, err := http.DefaultClient.Do(r)
	if err != nil {
		return err
	}
	if resp.StatusCode > 299 {
		return errors.New(fmt.Sprintf("Response failed. Status code is: %v", resp.StatusCode))
	}
	fmt.Printf("Downloaded %v bytes for chunk %v\n", resp.Header.Get("Content-Length"), i)
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(fmt.Sprintf("section-%v.tmp", i), b, os.ModePerm)
	if err != nil {
		return err
	}
	return nil
}
