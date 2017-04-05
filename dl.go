package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
)

type MaybeMP3 struct {
	Fname     string
	Err       error
	WasCached bool
}

func CacheMP3s(req Req, resp Resp, cb func(MaybeMP3)) {
	ch := make(chan MaybeMP3, len(resp.Items))
	for _, item := range resp.Items {
		go cachedMP3(item.PathMP3, req.CacheMP3Fname(item.PathMP3), ch)
	}
	for range resp.Items {
		cb(<-ch)
	}
}

func cachedMP3(url, fname string, result chan MaybeMP3) {
	if !*refreshCache {
		if _, err := os.Stat(fname); err == nil {
			// cached we are done
			result <- MaybeMP3{fname, nil, true}
			return
		}
	}
	// not cached, download it
	r, err := http.Get(url)
	if err != nil {
		result <- MaybeMP3{"", err, false}
		return
	}
	defer r.Body.Close()
	if r.StatusCode < 200 || r.StatusCode >= 300 {
		result <- MaybeMP3{"", fmt.Errorf("bad download status for MP3 file: %v", r.Status), false}
		return
	}
	buf, err := ioutil.ReadAll(r.Body)
	if err != nil {
		result <- MaybeMP3{"", fmt.Errorf("error while downloading MP3: %v", err), false}
		return
	}
	if err := ioutil.WriteFile(fname, buf, 0666); err != nil {
		result <- MaybeMP3{"", fmt.Errorf("error saving MP3: %v", err), false}
		return
	}
	// we're good
	result <- MaybeMP3{fname, nil, false}
}
