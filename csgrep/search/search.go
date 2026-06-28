package search

import (
	"sync"

	"github.com/amustafa/csgrep/session"
)

type Matcher interface {
	Match(text string) (matched bool, offsets [][2]int, score float64)
}

type Match struct {
	Session      *session.Session
	Message      session.Message
	MessageIndex int
	Score        float64
	Offsets      [][2]int
}

type Options struct {
	IncludeToolContent bool
	Workers            int
}

func Run(files []string, matcher Matcher, opts Options) []Match {
	if opts.Workers <= 0 {
		opts.Workers = 4
	}

	fileCh := make(chan string, len(files))
	for _, f := range files {
		fileCh <- f
	}
	close(fileCh)

	var mu sync.Mutex
	var allMatches []Match
	var wg sync.WaitGroup

	for i := 0; i < opts.Workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for path := range fileCh {
				parseOpts := session.ParseOptions{
					IncludeToolContent: opts.IncludeToolContent,
				}
				s, err := session.Parse(path, parseOpts)
				if err != nil || s == nil {
					continue
				}

				var fileMatches []Match
				for i, msg := range s.Messages {
					cleaned := session.CleanText(msg.Text)
					matched, offsets, score := matcher.Match(cleaned)
					if matched {
						msg.Text = cleaned
						fileMatches = append(fileMatches, Match{
							Session:      s,
							Message:      msg,
							MessageIndex: i,
							Score:        score,
							Offsets:      offsets,
						})
					}
				}

				if len(fileMatches) > 0 {
					mu.Lock()
					allMatches = append(allMatches, fileMatches...)
					mu.Unlock()
				}
			}
		}()
	}

	wg.Wait()
	return allMatches
}
