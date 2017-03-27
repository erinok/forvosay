// Command forvosay downloads and plays pronunciations from Forvo.com (using mplayer).
//
// Results are cached in ~/.forvocache.
package main

import (
	"flag"
	"fmt"
	"os"
)

var word = flag.String("word", "", "say this `word` or phrase")
var lang = flag.String("lang", "", "2-letter language `code`")
var refreshCache = flag.Bool("refresh", false, "download results even if already in cache")
var numDL = flag.Int("num", 3, "(`max`) number of pronunciations to download and play")

func main() {
	flag.Usage = func() {
		fmt.Fprint(os.Stderr, os.Args[0], `

download and play pronunciations from Forvo.com.
Results are cached in ~/.forvocache.

dependencies:

- FORVO_API_KEY must be set in your environment
- mplayer must be in your PATH

options:
`)
		flag.PrintDefaults()
	}
	flag.Parse()
	if *word == "" {
		fatal("must pass -word")
	}
	if *lang == "" {
		fatal("must pass -lang")
	}
	if apiKey == "" {
		fatal("must set FORVO_API_KEY in environment")
	}
	if *numDL < 1 {
		fatal("-num must be >= 1")
	}
	req := Req{*word, *lang}
	resp, err := CacheResp(req)
	if err != nil {
		fatal("could not download results:", err)
	}
	if len(resp.Items) > *numDL {
		fmt.Println("playing first", *numDL, "of", len(resp.Items), "pronunciation(s)...")
		resp.Items = resp.Items[:*numDL]
	} else {
		fmt.Println("playing", len(resp.Items), "pronunciation(s)...")
	}
	errs := false
	i := 0
	CacheMP3s(req, *resp, func(mp3 MaybeMP3) {
		i += 1
		if mp3.Err != nil {
			fmt.Fprintln(os.Stderr, "could not download mp3:", mp3.Err)
			errs = true
			return
		}
		fmt.Println(i)
		err := PlayMP3(mp3.Fname)
		if err != nil {
			fatal("could not play mp3:", err)
		}
	})
	if errs {
		os.Exit(1)
	}
}

func fatal(v ...interface{}) {
	fmt.Fprintln(os.Stderr, v...)
	os.Exit(1)
}
