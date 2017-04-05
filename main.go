// Command forvosay downloads and plays pronunciations from Forvo.com (using mplayer).
//
// Results are cached in ~/.forvocache.
package main

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"strings"
	"time"
)

var word = flag.String("word", "", "say this `word` or phrase")
var lang = flag.String("lang", "", "2-letter language `code`")
var refreshCache = flag.Bool("refresh", false, "download results even if already in cache")
var numPlay = flag.Int("n", 1, "(`max`) number of pronunciations to play (random order)")
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
	if *numPlay < 1 {
		fatal("-num must be >= 1")
	}
	*word = strings.ToLower(*word) // pretty sure forvo doesn't distinguish by case, so go ahead and normalize and get more use out of the cache
	rand.Seed(time.Now().UnixNano())
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
		errs := false
		var mp3s []string
		played := false
		seen := 0
		CacheMP3s(req, *resp, func(mp3 MaybeMP3) {
			seen += 1
			if mp3.Err != nil {
				fmt.Fprintln(os.Stderr, "could not download mp3:", mp3.Err)
				errs = true
				return
			}
			mp3s = append(mp3s, mp3.Fname)
			if !played && !mp3.WasCached && len(mp3s) >= *numPlay {
				playMP3s(mp3s, len(resp.Items))
				played = true
			}
		})
		if !played {
			playMP3s(mp3s, len(resp.Items))
		}
		if errs {
			os.Exit(1)
		}
	}
}

func playMP3s(mp3s []string, tot int) {
	n := *numPlay
	if n > len(mp3s) {
		n = len(mp3s)
	}
	for i := 1; i <= n; i++ {
		fmt.Println("playing", i, "/", n, fmt.Sprint("(of ", tot, ")"))
		var mp3 string
		mp3s, mp3 = popRand(mp3s)
		if err := PlayMP3(mp3); err != nil {
			fatal("could not play mp3:", err)
		}
	}
}

func popRand(xs []string) ([]string, string) {
	i, k := rand.Intn(len(xs)), len(xs)-1
	xs[i], xs[k] = xs[k], xs[i]
	return xs[:k], xs[k]
}

func fatal(v ...interface{}) {
	fmt.Fprintln(os.Stderr, v...)
	os.Exit(1)
}
