package main

import (
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
)

var cacheDir = os.Getenv("HOME") + "/.forvocache"

func (req Req) CacheDir() string {
	return cacheDir + "/" + req.LangCode + "/" + hexsha1(req.Word)
}

func (req Req) CacheFname() string {
	return req.CacheDir() + "/resp.json"
}

func (req Req) CacheMP3Fname(mp3url string) string {
	return req.CacheDir() + "/" + hexsha1(mp3url) + ".mp3"
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

func hexsha1(s string) string {
	return fmt.Sprintf("%0x", sha1.Sum([]byte(s)))
}
