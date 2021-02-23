package main

import (
	"embed"
	"encoding/json"
	"index/suffixarray"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/blevesearch/bleve/v2"
	"github.com/lithammer/dedent"
	"github.com/lithammer/fuzzysearch/fuzzy"
	"golang.org/x/text/language"
	"golang.org/x/text/search"
)

//go:embed completeworks.txt
var completeworks string

//go:embed static
var static embed.FS

func main() {
	completeworks = normalize(completeworks)

	log.Println("Building indexes...")
	tm := NewTextMatcher(completeworks)
	sm := NewSuffixArrayMatcher(completeworks)
	scm := NewSuffixArrayIgnoreCaseMatcher(completeworks)
	fm := NewFuzzyMatcher(completeworks)
	bm, _ := NewBleveMatcher(completeworks)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/static/", http.StatusMovedPermanently)
	})
	http.Handle("/static/", http.FileServer(http.FS(static)))
	http.HandleFunc("/search", handleSearch(tm, sm, scm, fm, bm))

	addr := ":" + getenv("PORT", "3001")
	log.Printf("Listening on address %s...\n", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatal(err)
	}
}

func handleSearch(searchers ...Searcher) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		query, ok := r.URL.Query()["q"]
		if !ok || len(query[0]) < 1 {
			http.Error(w, "missing search query in URL params", http.StatusBadRequest)
			return
		}

		ch := make(chan []string, len(searchers))
		for _, s := range searchers {
			go func(s Searcher) {
				ch <- s.Search(query[0])
			}(s)
		}

		var results []string
		for i := 0; i < len(searchers); i++ {
			results = <-ch
			if len(results) != 0 {
				break
			}
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(results); err != nil {
			http.Error(w, "encoding failure", http.StatusBadRequest)
			return
		}
	}
}

func normalize(s string) string {
	s = strings.Replace(s, "\n\n", "\n", -1)
	return dedent.Dedent(s)
}

func getenv(key, defval string) string {
	port := os.Getenv("PORT")
	if port == "" {
		return defval
	}
	return port
}

type Searcher interface {
	Search(string) []string
}

type BleveMatcher struct {
	index bleve.Index
}

func NewBleveMatcher(b string) (*BleveMatcher, error) {
	mapping := bleve.NewIndexMapping()
	index, err := bleve.NewMemOnly(mapping)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(b, "\n")
	for i, line := range lines {
		index.Index(strconv.Itoa(i), line)
	}
	return &BleveMatcher{index: index}, nil
}

func (m *BleveMatcher) Search(s string) []string {
	query := bleve.NewFuzzyQuery(s)
	search := bleve.NewSearchRequest(query)
	search.Fields = []string{"*"}
	results, err := m.index.Search(search)
	if err != nil {
		return nil
	}

	lines := make([]string, 0, 16)
	for _, doc := range results.Hits {
		line, ok := doc.Fields[""].(string)
		if ok {
			lines = append(lines, line)
		}
	}
	return lines
}

type SuffixArrayMatcher struct {
	completeWorks string
	suffixArray   *suffixarray.Index
}

func NewSuffixArrayMatcher(b string) *SuffixArrayMatcher {
	return &SuffixArrayMatcher{
		completeWorks: b,
		suffixArray:   suffixarray.New([]byte(b)),
	}
}

func (m *SuffixArrayMatcher) Search(s string) []string {
	idxs := m.suffixArray.Lookup([]byte(s), -1)
	results := []string{}
	for _, idx := range idxs {
		end := strings.Index(m.completeWorks[idx:], "\n")
		if end == -1 {
			end = idx
		}
		start := strings.LastIndex(m.completeWorks[:idx], "\n")
		if start == -1 {
			start = 0
		}
		results = append(results, m.completeWorks[start:idx+end])
	}
	return results
}

type SuffixArrayIgnoreCaseMatcher struct {
	completeWorks string
	suffixArray   *suffixarray.Index
}

func NewSuffixArrayIgnoreCaseMatcher(b string) *SuffixArrayIgnoreCaseMatcher {
	return &SuffixArrayIgnoreCaseMatcher{
		completeWorks: b,
		suffixArray:   suffixarray.New([]byte(b)),
	}
}

func (m *SuffixArrayIgnoreCaseMatcher) Search(s string) []string {
	reg, err := regexp.Compile("(?i)" + s)
	if err != nil {
		return nil
	}

	idxs := m.suffixArray.FindAllIndex(reg, -1)
	results := []string{}
	for _, idx := range idxs {
		results = append(results, m.completeWorks[idx[0]-250:idx[1]+250])
	}
	return results
}

type TextMatcher struct {
	matcher *search.Matcher
	text    []byte
}

func NewTextMatcher(b string) *TextMatcher {
	matcher := search.New(language.AmericanEnglish, search.IgnoreCase)

	return &TextMatcher{
		matcher: matcher,
		text:    []byte(b),
	}
}

func (m *TextMatcher) Search(s string) []string {
	var lines []string
	text := m.text
	pattern := m.matcher.CompileString(s)
	for {
		start, end := pattern.Index(text)
		if start == -1 {
			break
		}
		lines = append(lines, string(text[start:end]))
		text = text[end:]
	}
	return lines
}

type FuzzyMatcher struct {
	lines []string
}

func NewFuzzyMatcher(b string) *FuzzyMatcher {
	lines := strings.Split(b, "\n")
	return &FuzzyMatcher{lines: lines}
}

func (m *FuzzyMatcher) Search(s string) []string {
	return fuzzy.FindNormalizedFold(s, m.lines)
}
