package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"
)

var cacheDir = os.Getenv("HOME") + "/.forvocache"

func (req Req) CacheDir() string {
	return cacheDir + "/" + req.LangCode + "/" + sanitizeFname(req.Word)
}

func (req Req) CacheFname() string {
	return req.CacheDir() + "/.resp.json"
}

func (req Req) CacheMP3Fname(index int) string {
	return fmt.Sprintf("%s/%s-%02d.mp3", req.CacheDir(), sanitizeFname(req.Word), index+1)
}

func CacheResp(req Req) (*Resp, error) {
	if !*refreshCache {
		if resp, err := getCachedResp(req); err == nil {
			return resp, nil
		}
	}
	resp, err := Get(req)
	if err != nil {
		return nil, err
	}
	if err := saveRespToCache(req, *resp); err != nil {
		fmt.Println("warning: could not save pronunciation list to cache:", err)
	}
	return resp, nil
}

func getCachedResp(req Req) (*Resp, error) {
	f, err := os.Open(req.CacheFname())
	if err != nil {
		return nil, err
	}
	defer f.Close()
	buf, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}
	var resp Resp
	if err := json.Unmarshal(buf, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func saveRespToCache(req Req, resp Resp) error {
	buf, err := json.Marshal(&resp)
	if err != nil {
		return err
	}
	fname := req.CacheFname()
	dirname := path.Dir(fname)
	if err := os.MkdirAll(dirname, 0777); err != nil {
		return err
	}
	return ioutil.WriteFile(fname, buf, 0666)
}

func sanitizeFname(s string) string {
	// this works for mac, probably need to be fancier for other OSes...
	s = strings.Replace(s, "/", "47", -1)
	s = strings.Replace(s, ":", "48", -1)
	return s
}
