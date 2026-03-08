// package main

// import (
//     "flag"
//     "fmt"
//     "os"

//     "github.com/shubhcoding01/swiftprobe/internal/fuzzer"
// )

// func main() {
//     url      := flag.String("u", "", "Target URL (e.g., http://example.com)")
//     wordlist := flag.String("w", "", "Path to wordlist file")
//     threads  := flag.Int("t", 50, "Number of concurrent threads")
//     timeout  := flag.Int("timeout", 5, "HTTP request timeout in seconds")
//     codes    := flag.String("mc", "200,301,302,403", "Match HTTP status codes (comma-separated)")

//     flag.Parse()

//     if *url == "" || *wordlist == "" {
//         fmt.Println("Usage: swiftprobe -u http://example.com -w wordlist.txt")
//         flag.PrintDefaults()
//         os.Exit(1)
//     }

//     cfg := fuzzer.Config{
//         TargetURL:    *url,
//         WordlistPath: *wordlist,
//         Threads:      *threads,
//         TimeoutSecs:  *timeout,
//         MatchCodes:   *codes,
//     }

//     fuzzer.Run(cfg)
// }

package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/shubhcoding01/swiftprobe/internal/fuzzer"
)

// multiFlag allows -H to be repeated multiple times
// e.g. -H "Authorization: Bearer token" -H "X-Custom: value"
type multiFlag []string

func (m *multiFlag) String() string     { return strings.Join(*m, ", ") }
func (m *multiFlag) Set(v string) error { *m = append(*m, v); return nil }

func main() {
	url := flag.String("u", "", "Target URL (e.g. http://example.com)")
	wordlist := flag.String("w", "", "Path to wordlist file")
	threads := flag.Int("t", 50, "Number of concurrent threads")
	timeout := flag.Int("timeout", 5, "HTTP request timeout in seconds")
	codes := flag.String("mc", "200,201,301,302,401,403,500", "Match HTTP status codes (comma-separated)")
	exclude := flag.String("fc", "404", "Exclude HTTP status codes (comma-separated)")
	exts := flag.String("x", "", "File extensions to append e.g. php,bak,html,json")
	ua := flag.String("ua", "Mozilla/5.0 (compatible; SwiftProbe/1.0)", "Custom User-Agent string")
	follow := flag.Bool("follow", false, "Follow HTTP redirects (default: capture and log)")
	verbose := flag.Bool("v", false, "Verbose mode — show errors and live progress")
	output := flag.String("o", "", "Save results to output file e.g. results.txt")

	// Repeatable -H flag for custom headers
	var headers multiFlag
	flag.Var(&headers, "H", "Custom header e.g. -H \"Authorization: Bearer token\" (repeatable)")

	flag.Parse()

	// ── Validate required flags ──────────────────────────────
	if *url == "" {
		fmt.Fprintln(os.Stderr, "[ERROR] -u (target URL) is required")
		fmt.Println("\nUsage: swiftprobe -u http://example.com -w wordlist.txt [OPTIONS]")
		flag.PrintDefaults()
		os.Exit(1)
	}
	if *wordlist == "" {
		fmt.Fprintln(os.Stderr, "[ERROR] -w (wordlist path) is required")
		fmt.Println("\nUsage: swiftprobe -u http://example.com -w wordlist.txt [OPTIONS]")
		flag.PrintDefaults()
		os.Exit(1)
	}

	// ── Parse extensions into slice ──────────────────────────
	var extensions []string
	if *exts != "" {
		for _, e := range strings.Split(*exts, ",") {
			e = strings.TrimSpace(e)
			if e != "" {
				extensions = append(extensions, e)
			}
		}
	}

	// ── Parse -H headers into map ────────────────────────────
	// Accepts both "Key: Value" and "Key:Value" formats
	headerMap := make(map[string]string)
	for _, h := range headers {
		parts := strings.SplitN(h, ":", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			val := strings.TrimSpace(parts[1])
			headerMap[key] = val
		} else {
			fmt.Fprintf(os.Stderr,
				"[WARN] Skipping malformed header: %q — use \"Key: Value\" format\n", h)
		}
	}

	// ── Launch fuzzer ────────────────────────────────────────
	fuzzer.Run(fuzzer.Config{
		TargetURL:       *url,
		WordlistPath:    *wordlist,
		Threads:         *threads,
		TimeoutSecs:     *timeout,
		MatchCodes:      *codes,
		ExcludeCodes:    *exclude,
		Extensions:      extensions,
		UserAgent:       *ua,
		Headers:         headerMap,
		FollowRedirects: *follow,
		Verbose:         *verbose,
		OutputFile:      *output,
	})
}
