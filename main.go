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
	"sync/atomic"
	"time"
	"unicode"

	"github.com/atotto/clipboard"
)

var word = flag.String("word", "", "say this `word` or phrase")
var forever = flag.Bool("forever", false, "say words from the clipboard (run forever)")
var web = flag.Bool("web", false, "open cantoese.org w/ definition page")
var lang = flag.String("lang", "", "2-letter language `code`")
var refreshCache = flag.Bool("refresh", false, "download results even if already in cache")
var numSay = flag.Int("n", 3, "(`max`) number of pronunciations to play; < 0 for all")
var showFiles = flag.Bool("showFiles", false, "open the folder with the cached pronunciation files, instead of playing the files (using the command 'open')")
var fallback = flag.String("fallback", "", "if no pronuncations are found, fallback to using the 'say' command with this `voice`")
var nossl = flag.Bool("nossl", false, "don't use ssl when communicating with forvo.com; about twice as fast, but exposes your api key in plaintext")
var bench = flag.Bool("bench", false, "time the request to forvo.com")

func lookupWeb(word string) {
	cmd := exec.Command("open", "-a", "Safari", "--", "https://cantonese.org/search.php?q="+word)
	// fmt.Println("command:", cmd)
	err := cmd.Run()
	if err != nil {
		fmt.Fprintln(os.Stderr, "error opening safari:", err)
	}
}

func lookup(word string) error {
	return lookupFancy(word, func() bool { return true })
}

func lookupFancy(word string, keepGoing func() bool) error {
	word = strings.TrimSpace(word)
	word = strings.ToLower(word) // pretty sure forvo doesn't distinguish by case, so go ahead and normalize and get more use out of the cache
	if *web {
		go lookupWeb(word)
	}
	req := Req{word, *lang}
	resp, err := CacheResp(req)
	if err != nil {
		return fmt.Errorf("could not download results: %s", err)
	}
	if len(resp.Items) == 0 {
		if *fallback != "" {
			fmt.Println("no results; using 'say'")
			if err := exec.Command("say", "-v", *fallback, word).Run(); err != nil {
				return fmt.Errorf("could not 'say': %v", err)
			}
		} else {
			fmt.Println("no results")
		}
	} else {
		numSay := *numSay
		if n := len(resp.Items); numSay < 0 || numSay > n {
			numSay = n
		}
		if *showFiles {
			if err := exec.Command("open", req.CacheDir()).Run(); err != nil {
				fatal("could not show files:", err)
			}
			numSay = 0
		}
		var errs []error
		numSaid := 0
		CacheMP3s(req, *resp, func(mp3 MaybeMP3) {
			if mp3.Err != nil {
				errs = append(errs, fmt.Errorf("could not download mp3: %v", mp3.Err))
				return
			}
			if numSaid < numSay && keepGoing() {
				numSaid++
				fmt.Println("playing", numSaid, "/", numSay, fmt.Sprint("(of ", len(resp.Items), ")"))
				err := PlayMP3(mp3.Fname)
				if err != nil {
					errs = append(errs, fmt.Errorf("could not play mp3: %v", err))
				}
			}
		})
		if len(errs) > 0 {
			return errs[0]
		}
	}
	return nil
}

func maybePassword(s string) bool {
	for _, r := range s {
		if unicode.IsPunct(r) {
			return true
		}
	}
	return false
}

// lookup words from clipboard forever
func lookupForever() {
	var prev string
	var w int32
	for i := 0; ; i++ {
		if i > 0 {
			time.Sleep(100 * time.Millisecond)
		}
		s, err := clipboard.ReadAll()
		s = strings.TrimSpace(s)
		if err != nil || s == prev || s == "" {
			continue
		}
		if maybePassword(s) {
			fmt.Printf("skipping word containing punctuation (in case it's a password)")
			continue
		}
		prev = s
		if i == 0 {
			continue
		}
		if i > 1 {
			fmt.Println()
		}
		if len(s) > 100 {
			fmt.Printf("skipping long text `%v...`, \n", s[:100])
			continue
		}
		fmt.Printf("pronouncing `%v`...\n", s)
		this := atomic.AddInt32(&w, 1)
		go func() {
			err = lookupFancy(s, func() bool { return atomic.AddInt32(&w, 0) == this })
			if err != nil {
				fmt.Printf("error looking up `%v`: %v\n", s, err)
			}
		}()
	}
}

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
	if *lang == "" {
		fatal("must pass -lang")
	}
	if apiKey == "" {
		fatal("must set FORVO_API_KEY in environment")
	}
	if *forever {
		lookupForever()
	}
	if *word == "" {
		fatal("must pass -word or -forever")
	}
	err := lookup(*word)
	if err != nil {
		fatal(err)
	}
}

func fatal(v ...interface{}) {
	fmt.Fprintln(os.Stderr, v...)
	os.Exit(1)
}
