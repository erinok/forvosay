// Command forvocl players pronuncations from the Forvo website using mplayer.
//
// Results are cached in ~/.forvocache.
package main

import (
	"flag"
	"fmt"
	"os"
)

var word = flag.String("word", "", "lookup and say this `word`")
var lang = flag.String("lang", "", "2-letter language `code`")
var refreshCache = flag.Bool("refresh", false, "redownload results even if already in cache")
var numDL = flag.Int("num", 3, "(`max`) number of pronunciations to download and play")

func main() {
	flag.Usage = func() {
		fmt.Fprint(os.Stderr, os.Args[0], `download and play word pronunciations from Forvo.com
		
FORVO_API_KEY must be set in your environment

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
	resp, err := CachingGet(req)
	if err != nil {
		fatal("could not download results:", err)
	}
	fmt.Println(*resp)
	if len(resp.Items) > *numDL {
		resp.Items = resp.Items[:*numDL]
	}
	errs := CacheMP3s(req, *resp)
	for _, err := range errs {
		fmt.Fprintln(os.Stderr, "error downloading mp3:", err)
	}
	if len(errs) == len(resp.Items) {
		fatal("could not download any pronunciations: ", errs)
	}
	if err := PlayMP3s(req, *resp); err != nil {
		fatal("error playing mp3: %v", err)
	}
}

func fatal(v ...interface{}) {
	fmt.Fprintln(os.Stderr, v...)
	os.Exit(1)
}
