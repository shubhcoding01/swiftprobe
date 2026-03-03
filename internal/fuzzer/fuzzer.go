package fuzzer

import (
    "fmt"
    "strconv"
    "strings"
    "sync"

    "github.com/fatih/color"
    "github.com/yourname/swiftprobe/internal/requester"
    "github.com/yourname/swiftprobe/internal/wordlist"
)

type Config struct {
    TargetURL    string
    WordlistPath string
    Threads      int
    TimeoutSecs  int
    MatchCodes   string
}

func Run(cfg Config) {
    // Parse match codes
    matchSet := parseMatchCodes(cfg.MatchCodes)

    // Stream wordlist
    words, err := wordlist.Stream(cfg.WordlistPath)
    if err != nil {
        fmt.Println("Error reading wordlist:", err)
        return
    }

    client := requester.New(cfg.TimeoutSecs)

    // Semaphore to limit concurrency
    sem := make(chan struct{}, cfg.Threads)
    var wg sync.WaitGroup

    green  := color.New(color.FgGreen).SprintFunc()
    yellow := color.New(color.FgYellow).SprintFunc()
    red    := color.New(color.FgRed).SprintFunc()

    fmt.Printf("\n🔍 SwiftProbe — Target: %s | Threads: %d\n\n", cfg.TargetURL, cfg.Threads)

    for word := range words {
        wg.Add(1)
        sem <- struct{}{} // acquire

        go func(path string) {
            defer wg.Done()
            defer func() { <-sem }() // release

            url := requester.BuildURL(cfg.TargetURL, path)
            result, err := client.Probe(url)
            if err != nil {
                return // silently skip errors (timeouts, etc.)
            }

            if _, ok := matchSet[result.StatusCode]; ok {
                code := result.StatusCode
                var colored string
                switch {
                case code == 200:
                    colored = green(fmt.Sprintf("[%d]", code))
                case code == 301 || code == 302:
                    colored = yellow(fmt.Sprintf("[%d]", code))
                case code == 403:
                    colored = red(fmt.Sprintf("[%d]", code))
                default:
                    colored = fmt.Sprintf("[%d]", code)
                }
                fmt.Printf("%s  /%s  (Size: %d)\n", colored, path, result.Size)
            }
        }(word)
    }

    wg.Wait()
    fmt.Println("\n✅ Scan complete.")
}

func parseMatchCodes(raw string) map[int]struct{} {
    set := make(map[int]struct{})
    for _, s := range strings.Split(raw, ",") {
        if n, err := strconv.Atoi(strings.TrimSpace(s)); err == nil {
            set[n] = struct{}{}
        }
    }
    return set
}