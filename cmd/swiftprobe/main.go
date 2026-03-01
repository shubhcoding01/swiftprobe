package main

import (
    "flag"
    "fmt"
    "os"

    "github.com/shubhcoding01/swiftprobe/internal/fuzzer"
)

func main() {
    url      := flag.String("u", "", "Target URL (e.g., http://example.com)")
    wordlist := flag.String("w", "", "Path to wordlist file")
    threads  := flag.Int("t", 50, "Number of concurrent threads")
    timeout  := flag.Int("timeout", 5, "HTTP request timeout in seconds")
    codes    := flag.String("mc", "200,301,302,403", "Match HTTP status codes (comma-separated)")

    flag.Parse()

    if *url == "" || *wordlist == "" {
        fmt.Println("Usage: swiftprobe -u http://example.com -w wordlist.txt")
        flag.PrintDefaults()
        os.Exit(1)
    }

    cfg := fuzzer.Config{
        TargetURL:    *url,
        WordlistPath: *wordlist,
        Threads:      *threads,
        TimeoutSecs:  *timeout,
        MatchCodes:   *codes,
    }

    fuzzer.Run(cfg)
}