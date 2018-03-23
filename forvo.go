package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"time"
)

var apiKey = os.Getenv("FORVO_API_KEY")

type Pronunciation struct {
	Id               int64
	Word             string
	Original         string
	AddTime          string
	Username         string
	Sex              string
	Country          string
	Code             string
	Langname         string
	PathMP3          string
	PathOGG          string
	Rate             int
	NumVotes         int `json:"num_votes"`
	NumPositiveVotes int `json:"num_positive_votes"`
}

type Resp struct {
	Items []Pronunciation
}

type Req struct {
	Word     string
	LangCode string // e.g. "de"
}

func (req Req) CacheDir() string {
	return cacheDir + "/" + req.LangCode + "/" + sanitizeFname(req.Word)
}

func (req Req) CacheFname() string {
	return req.CacheDir() + "/.resp.json"
}

func (req Req) CacheMP3Fname(index int) string {
	return fmt.Sprintf("%s/%s-%02d.mp3", req.CacheDir(), sanitizeFname(req.Word), index+1)
}

func Get(req Req) (*Resp, error) {
	if *bench {
		t0 := time.Now()
		defer func() { fmt.Println(time.Since(t0)) }()
	}
	if req.Word == "0" {
		req.Word = " 0" // workaround apparent forvo bug
	}
	addr := "https://apifree.forvo.com"
	if *nossl {
		addr = "http://apifree.forvo.com"
	}
	url := fmt.Sprint(
		addr,
		"/key/", apiKey,
		"/format/json",
		"/action/word-pronunciations",
		"/word/", url.PathEscape(req.Word),
		"/language/", req.LangCode,
		"/order/rate-desc",
	)
	fmt.Println("downloading pronunciation list...")
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	buf, err := ioutil.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("forvo complained: HTPP %s (%s)", resp.Status, string(buf))
	}
	if err != nil {
		return nil, fmt.Errorf("error reading forvo response body: %s", err)
	}
	var pr Resp
	if err := json.Unmarshal(buf, &pr); err != nil {
		return nil, fmt.Errorf("forvo response «%s» could not be unmarshalled as json: %v", err, string(buf))
	}
	return &pr, nil
}
