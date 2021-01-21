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
	err := d.DownloadFile()
	if err != nil {
		log.Printf("An error occured while downloading the file: %s\n", err)
	}
	fmt.Printf("Download completed in %v seconds\n", time.Now().Sub(startTime).Seconds())
}

func (d Download) DownloadFile() error {
	fmt.Println("Checking URL...")
	// Create new HTTP request
	req, err := http.NewRequest(
		("HEAD"),
		d.Url,
		nil,
	)
	if err != nil {
		return err
	}
	// Set HTTP headers
	req.Header.Set("User-Agent", "File Downloader")

	// Create HTTP client
	client := &http.Client{
		Timeout: time.Second * 10,
	}

	// Make the HTTP request
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
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
	return d.mergeFiles(chunks)
}

func (d Download) downloadChunk(i int, c [2]int) error {
	req, err := http.NewRequest(
		"GET",
		d.Url,
		nil,
	)
	if err != nil {
		return err
	}
	// Set the Range Headers to our chunks of bytes
	// that we'll pass in our goroutine
	req.Header.Set("Range", fmt.Sprintf("bytes=%v-%v", c[0], c[1]))

	client := &http.Client{
		Timeout: time.Second * 10,
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode > 299 {
		return errors.New(fmt.Sprintf("Response failed. Status code is: %v", resp.StatusCode))
	}
	fmt.Printf("Downloaded %v bytes for chunk %v\n", resp.Header.Get("Content-Length"), i)

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(fmt.Sprintf("chunk-%v.tmp", i), b, os.ModePerm)
	if err != nil {
		return err
	}
	return nil
}

func (d Download) mergeFiles(sections [][2]int) error {
	f, err := os.OpenFile(d.TargetPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, os.ModePerm)
	if err != nil {
		return err
	}
	defer f.Close()
	for i := range sections {
		tmpFileName := fmt.Sprintf("chunk-%v.tmp", i)
		b, err := ioutil.ReadFile(tmpFileName)
		if err != nil {
			return err
		}
		n, err := f.Write(b)
		if err != nil {
			return err
		}
		err = os.Remove(tmpFileName)
		if err != nil {
			return err
		}
		fmt.Printf("%v bytes merged\n", n)
	}
	return nil
}
