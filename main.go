// Command forvosay downloads and plays pronunciations from Forvo.com (using afplay).
//
// Results are cached in ~/.forvocache.
//
// BUG: if forvo returns error, we still save corrupted mp3
package main

import (
	"bufio"
	"flag"
	"fmt"
	"math/rand"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"sync"
	"sync/atomic"
	"time"
	"unicode/utf8"

	"github.com/atotto/clipboard"
	"github.com/tevino/abool"
)

var word = flag.String("word", "", "say this `word` or phrase")
var forever = flag.Bool("forever", false, "say words from the clipboard (run forever)")

var lang = flag.String("lang", "", "2 or 3 letter language `code`")
var refreshCache = flag.Bool("refresh", false, "download results even if already in cache")

var numSay = flag.Int("n", 1, "`max` number of pronunciations to play; < 0 for all")
var topSay = flag.Int("top", 5, "draw the N pronunciations to play randomly from the top `T`")

var showFiles = flag.Bool("showFiles", false, "open the folder with the cached pronunciation files, instead of playing the files (using the command 'open')")
var fallback = flag.String("fallback", "", "if no pronuncations are found, fallback to using the 'say' command with this `voice`")
var nossl = flag.Bool("nossl", false, "don't use ssl when communicating with forvo.com; about twice as fast, but exposes your api key in plaintext")
var bench = flag.Bool("bench", false, "time the request to forvo.com")

var canto = flag.Bool("canto", false, "search cantonese.org for definitions")
var yi = flag.Bool("yi", false, "search yandex for images")
var gi = flag.String("gi", "", "search google.`GI` for images")
var bi = flag.Bool("bi", false, "search baidu for images")

var dict = flag.Bool("dict", false, "open dict:// (the builtin mac dictionary) for definitions")

var yt = flag.String("yt", "", "yandex translate sentences from language LA to language LB (`LA-LB`)")
var gt = flag.Bool("gt", false, "google translate sentences (must pick language using UI)")

var chrome = flag.String("chrome", "", "comma-separated list of flags that should use chrome browser (instead of system default)")

func openUrl(flag, url string) {
	var cmd *exec.Cmd
	if strings.Contains(*chrome, ","+flag+",") {
		cmd = exec.Command("open", "-a", "Google Chrome", url)
	} else {
		cmd = exec.Command("open", url)
	}
	err := cmd.Run()
	if err != nil {
		fmt.Fprint(os.Stderr, "error opening browser for -", flag, ": ", err, "\n")
	}
}

func lookupWebCanto(word string) {
	openUrl("canto", "https://cantonese.org/search.php?q="+url.QueryEscape(word))
}

func lookupWebYandexImages(word string) {
	openUrl("yi", "https://yandex.ru/images/search?text="+url.QueryEscape(word))
}

func lookupWebYandexTrans(lang, s string) {
	openUrl("yt", "https://translate.yandex.ru/?lang="+lang+"&text="+url.QueryEscape(s))
}

func lookupWebGoogleTrans(s string) {
	openUrl("gt", fmt.Sprint("https://translate.google.com/?text=", url.QueryEscape(s), "&op=trranslate"))
}

func lookupWebGoogleImages(country string, word string) {
	openUrl("gi", "https://www.google."+country+"/search?tbm=isch&q="+url.QueryEscape(word))
}

func lookupWebBaiduImages(word string) {
	url := "https://image.baidu.com/search/index?tn=baiduimage&ie=utf-8&word=" + url.QueryEscape(word)
	openUrl("bi", url)
}

func lookupDict(word string) {
	cmd := exec.Command("open", "dict://"+url.QueryEscape(word))
	// fmt.Println("command:", cmd)
	err := cmd.Run()
	if err != nil {
		fmt.Fprintln(os.Stderr, "error opening url:", err)
	}
}

func lookup(word string) error {
	return lookupFancy(word, false, false, func(_ string) int { return 0 }, func(_ string) {}, func() bool { return true })
}

func onlyMinimalPlayCounts(req Req, resp Resp, getPlayCount func(string) int) Resp {
	n := len(resp.Items)
	c := make([]int, n)
	minC := 1000
	for i, r := range resp.Items {
		fn := req.CacheMP3Fname(r.Index)
		c[i] = getPlayCount(fn)
		if c[i] < minC {
			minC = c[i]
		}

	}
	resp2 := Resp{}
	for i, item := range resp.Items {
		if c[i] == minC {
			resp2.Items = append(resp2.Items, item)
		}
	}
	return resp2
}

func lookupSentence(s string) error {
	if *yt != "" {
		lookupWebYandexTrans(*yt, s)
	}
	if *gt {
		lookupWebGoogleTrans(s)
	}
	return nil
}

// TODO: this api is jank af
func lookupFancy(word string, repeat bool, onlyForvo bool, getPlayCount func(string) int, incrPlayCount func(string), keepGoing func() bool) error {
	word = strings.TrimSpace(word)
	word = strings.ToLower(word) // pretty sure forvo doesn't distinguish by case, so go ahead and normalize and get more use out of the cache
	if !repeat && !onlyForvo {
		if *dict {
			lookupDict(word)
		}
		if *canto {
			lookupWebCanto(word)
		}
		if *yi {
			lookupWebYandexImages(word)
		}
		if *gi != "" {
			lookupWebGoogleImages(*gi, word)
		}
		if *bi {
			lookupWebBaiduImages(word)
		}
	}
	req := Req{word, *lang}
	resp, err := CacheResp(req)
	if err != nil {
		return fmt.Errorf("could not download results: %s", err)
	}
	if len(resp.Items) == 0 {
		if *fallback != "" && !onlyForvo {
			fmt.Println("no results; using 'say'")
			if err := exec.Command("say", "-v", *fallback, word).Run(); err != nil {
				return fmt.Errorf("could not 'say': %v", err)
			}
		} else {
			fmt.Println("no results")
		}
	} else {
		origN := len(resp.Items)
		*resp = onlyMinimalPlayCounts(req, *resp, getPlayCount)
		n := len(resp.Items)
		numSay := *numSay
		topSay := *topSay
		if numSay < 0 || numSay > n {
			numSay = n
		}
		if topSay < 0 || topSay > n {
			topSay = n
		}
		rand.Shuffle(topSay, func(i, j int) {
			resp.Items[i], resp.Items[j] = resp.Items[j], resp.Items[i]
		})

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
				fmt.Println(mp3.Fname, "of", origN)
				incrPlayCount(mp3.Fname)
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
	if maybeSentence(s) {
		return false // if it's (probably) a sentence, it's (probably) not a password
	}
	n := 0
	for _, r := range s {
		if r == '-' || r == '.' {
			n++
		}
	}
	return n >= 2
}

func maybeSentence(s string) bool {
	if *yt == "" && !*gt {
		return false // no translate set so always assume not sentence
	}
	if *canto {
		return utf8.RuneCountInString(s) >= 4
	}
	return len(strings.Fields(s)) >= 4
}

func shouldSkip(s string) bool {
	if utf8.RuneCountInString(s) > 1000 {
		return true
	}
	if maybePassword(s) {
		return true
	}
	return false
}

func trackPlayCounts() (get func(string) int, incr func(string)) {
	d := map[string]int{}
	mu := sync.Mutex{}
	get = func(fname string) int {
		mu.Lock()
		defer mu.Unlock()
		return d[fname]
	}
	incr = func(fname string) {
		mu.Lock()
		defer mu.Unlock()
		d[fname] = d[fname] + 1
	}
	return
}

// lookup words from clipboard forever
func lookupForever() {
	var prev string
	var w int32
	var word atomic.Value
	repeat := abool.New()
	getPlayCount, incrPlayCount := trackPlayCounts()
	go func() {
		b := bufio.NewReader(os.Stdin)
		for {
			s, err := b.ReadString('\n')
			if err != nil {
				fmt.Println("error reading input:", err, "giving up...")
				return
			}
			s = strings.TrimSpace(s)
			if s == "" {
				repeat.Set()
				continue
			}
			w, _ := word.Load().(string)
			if w == "" {
				continue
			}
			switch s {
			case "d":
				lookupDict(w)
			case "y":
				lookupWebYandexImages(w)
			case "g":
				lookupWebGoogleImages(*lang, w)
			case "c":
				lookupWebCanto(w)
			default:
				fmt.Println("unknown input:", s)
				fmt.Println("try: d, y, g, c")
			}
		}
	}()
	for i := 0; ; i++ {
		if i > 0 {
			time.Sleep(100 * time.Millisecond)
		}
		s, err := clipboard.ReadAll()
		s = strings.TrimSpace(s)
		r := repeat.IsSet()
		if err != nil || (s == prev && !r) || s == "" {
			continue
		}
		repeat.UnSet()
		word.Store(s)
		prev = s
		if i == 0 {
			// skip whatever's initially on the clipboard (somehow it's annoying to pick this up)
			continue
		}
		if shouldSkip(s) {
			fmt.Printf("skipping word that looks like a password or very long body of text")
			continue
		}
		if i > 1 && !r {
			fmt.Println()
		}
		onlyForvo := false
		if maybeSentence(s) {
			fmt.Printf("looking up sentence `%v`...\n", s)
			lookupSentence(s)
			onlyForvo = true
		}
		this := atomic.AddInt32(&w, 1)
		go func() {
			err = lookupFancy(s, r, onlyForvo, getPlayCount, incrPlayCount, func() bool { return atomic.AddInt32(&w, 0) == this })
			if err != nil {
				fmt.Printf("error looking up `%v`: %v\n", s, err)
			}
		}()
	}
}

func main() {
	flag.Usage = func() {
		fmt.Fprint(os.Stderr, os.Args[0], ` [options]

Download and play pronunciations for WORD in language LANG from Forvo.com.

Results are cached in ~/.forvocache.

dependencies:

- FORVO_API_KEY must be set in your environment
- afplay must be in your PATH (in /usr/bin on macs)

options:
`)
		flag.PrintDefaults()
	}
	flag.Parse()
	*chrome = "," + *chrome + ","
	if len(flag.Args()) > 0 {
		fatal("unknown argument:", flag.Args()[0])
	}
	rand.Seed(int64(time.Now().Unix()))
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
