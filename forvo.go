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

func Get(req Req) (*Resp, error) {
	if *bench {
		t0 := time.Now()
		defer func() {
			fmt.Println(time.Since(t0))
		}()
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
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("bad forvo HTPP status: %s", resp.Status)
	}
	buf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var pr Resp
	if err := json.Unmarshal(buf, &pr); err != nil {
		return nil, fmt.Errorf("forvo response could not be unmarshalled as json: %v", err)
	}
	return &pr, nil
}
