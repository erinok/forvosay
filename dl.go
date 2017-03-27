package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
)

func CacheMP3s(req Req, resp Resp) []error {
	ch := make(chan error, len(resp.Items))
	for _, item := range resp.Items {
		fmt.Println(item.PathMP3)
		go cachedMP3(item.PathMP3, req.CacheMP3Fname(item.PathMP3), ch)
	}
	var errs []error
	for range resp.Items {
		if err := <-ch; err != nil {
			errs = append(errs, err)
		}
	}
	return errs
}

func cachedMP3(url, fname string, result chan error) {
	if _, err := os.Stat(fname); err == nil {
		// cached we are done
		result <- nil
		return
	}
	// not cached, download it
	r, err := http.Get(url)
	if err != nil {
		result <- err
		return
	}
	defer r.Body.Close()
	if r.StatusCode < 200 || r.StatusCode >= 300 {
		result <- fmt.Errorf("bad download status for MP3 file: %v", r.Status)
		return
	}
	buf, err := ioutil.ReadAll(r.Body)
	if err != nil {
		result <- fmt.Errorf("error while downloading MP3: %v", err)
		return
	}
	if err := ioutil.WriteFile(fname, buf, 0666); err != nil {
		result <- fmt.Errorf("error saving MP3: %v", err)
		return
	}
	// we're good
	result <- nil
}
