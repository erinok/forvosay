// Command forvosay downloads and plays pronunciations from Forvo.com (using mplayer).
//
// Results are cached in ~/.forvocache.
package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

var word = flag.String("word", "", "say this `word` or phrase")
var lang = flag.String("lang", "", "2-letter language `code`")
var refreshCache = flag.Bool("refresh", false, "download results even if already in cache")
var numDL = flag.Int("n", 3, "(`max`) number of pronunciations to download and play")
var fallback = flag.String("fallback", "", "if no pronuncations are found, fallback to using the 'say' command with this `voice`")
var nossl = flag.Bool("nossl", false, "don't use ssl when communicating with forvo.com; about twice as fast, but exposes your api key in plaintext")
var bench = flag.Bool("bench", false, "time the request to forvo.com")

func main() {
	flag.Usage = func() {
		fmt.Fprint(os.Stderr, os.Args[0], ` -lang LANG -word WORD

Download and play pronunciations for WORD in language LANG from Forvo.com.

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
	*word = strings.ToLower(*word) // pretty sure forvo doesn't distinguish by case, so go ahead and normalize and get more use out of the cache
	req := Req{*word, *lang}
	resp, err := CacheResp(req)
	if err != nil {
		fatal("could not download results:", err)
	}
	if len(resp.Items) == 0 {
		if *fallback != "" {
			fmt.Println("no results; using 'say'")
			if err := exec.Command("say", "-v", *fallback, *word).Run(); err != nil {
				fatal("could not 'say':", err)
			}
		} else {
			fmt.Println("no results")
		}
	} else {
		tot := len(resp.Items)
		if tot > *numDL {
			resp.Items = resp.Items[:*numDL]
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
			fmt.Println("playing", i, "/", len(resp.Items), fmt.Sprint("(of ", tot, ")"))
			err := PlayMP3(mp3.Fname)
			if err != nil {
				fatal("could not play mp3:", err)
			}
		})
		if errs {
			os.Exit(1)
		}
	}
}

func fatal(v ...interface{}) {
	fmt.Fprintln(os.Stderr, v...)
	os.Exit(1)
}
