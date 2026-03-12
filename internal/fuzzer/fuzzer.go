// package fuzzer

// import (
//     "fmt"
//     "strconv"
//     "strings"
//     "sync"

//     "github.com/fatih/color"
//     "github.com/shubhcoding01/swiftprobe/internal/requester"
//     "github.com/shubhcoding01/swiftprobe/internal/wordlist"
// )

// type Config struct {
//     TargetURL    string
//     WordlistPath string
//     Threads      int
//     TimeoutSecs  int
//     MatchCodes   string
// }

// func Run(cfg Config) {
//     // Parse match codes
//     matchSet := parseMatchCodes(cfg.MatchCodes)

//     // Stream wordlist
//     words, err := wordlist.Stream(cfg.WordlistPath)
//     if err != nil {
//         fmt.Println("Error reading wordlist:", err)
//         return
//     }

//     client := requester.New(cfg.TimeoutSecs)

//     // Semaphore to limit concurrency
//     sem := make(chan struct{}, cfg.Threads)
//     var wg sync.WaitGroup

//     green  := color.New(color.FgGreen).SprintFunc()
//     yellow := color.New(color.FgYellow).SprintFunc()
//     red    := color.New(color.FgRed).SprintFunc()

//     fmt.Printf("\n🔍 SwiftProbe — Target: %s | Threads: %d\n\n", cfg.TargetURL, cfg.Threads)

//     for word := range words {
//         wg.Add(1)
//         sem <- struct{}{} // acquire

//         go func(path string) {
//             defer wg.Done()
//             defer func() { <-sem }() // release

//             url := requester.BuildURL(cfg.TargetURL, path)
//             result, err := client.Probe(url)
//             if err != nil {
//                 return // silently skip errors (timeouts, etc.)
//             }

//             if _, ok := matchSet[result.StatusCode]; ok {
//                 code := result.StatusCode
//                 var colored string
//                 switch {
//                 case code == 200:
//                     colored = green(fmt.Sprintf("[%d]", code))
//                 case code == 301 || code == 302:
//                     colored = yellow(fmt.Sprintf("[%d]", code))
//                 case code == 403:
//                     colored = red(fmt.Sprintf("[%d]", code))
//                 default:
//                     colored = fmt.Sprintf("[%d]", code)
//                 }
//                 fmt.Printf("%s  /%s  (Size: %d)\n", colored, path, result.Size)
//             }
//         }(word)
//     }

//     wg.Wait()
//     fmt.Println("\n✅ Scan complete.")
// }

// func parseMatchCodes(raw string) map[int]struct{} {
//     set := make(map[int]struct{})
//     for _, s := range strings.Split(raw, ",") {
//         if n, err := strconv.Atoi(strings.TrimSpace(s)); err == nil {
//             set[n] = struct{}{}
//         }
//     }
//     return set
// }

// package fuzzer

// import (
// 	"fmt"
// 	"os"
// 	"sync"
// 	"sync/atomic"
// 	"time"

// 	"github.com/shubhcoding01/swiftprobe/internal/output"
// 	"github.com/shubhcoding01/swiftprobe/internal/requester"
// 	"github.com/shubhcoding01/swiftprobe/internal/wordlist"
// )

// // Config holds all scan parameters passed in from main.go
// type Config struct {
// 	TargetURL       string
// 	WordlistPath    string
// 	Threads         int
// 	TimeoutSecs     int
// 	MatchCodes      string
// 	ExcludeCodes    string
// 	Extensions      []string
// 	UserAgent       string
// 	Headers         map[string]string
// 	FollowRedirects bool
// 	Verbose         bool
// 	OutputFile      string
// }

// // Run is the main entry point — orchestrates the full scan
// func Run(cfg Config) {
// 	if cfg.Threads <= 0 {
// 		cfg.Threads = 50
// 	}
// 	if cfg.TimeoutSecs <= 0 {
// 		cfg.TimeoutSecs = 5
// 	}
// 	if cfg.MatchCodes == "" {
// 		cfg.MatchCodes = "200,201,301,302,401,403,500"
// 	}

// 	printer := output.New(cfg.Verbose)
// 	output.Banner()
// 	output.PrintConfig(
// 		cfg.TargetURL,
// 		cfg.WordlistPath,
// 		cfg.MatchCodes,
// 		cfg.Threads,
// 		cfg.TimeoutSecs,
// 	)

// 	filter := NewResultFilter(cfg.MatchCodes, cfg.ExcludeCodes)

// 	client := requester.New(requester.Config{
// 		TimeoutSecs:     cfg.TimeoutSecs,
// 		UserAgent:       cfg.UserAgent,
// 		Headers:         cfg.Headers,
// 		FollowRedirects: cfg.FollowRedirects,
// 	})

// 	words, err := wordlist.Stream(cfg.WordlistPath)
// 	if err != nil {
// 		fmt.Fprintf(os.Stderr, "[ERROR] Could not open wordlist: %v\n", err)
// 		os.Exit(1)
// 	}

// 	sem := make(chan struct{}, cfg.Threads)
// 	var wg sync.WaitGroup
// 	var totalScanned atomic.Int64
// 	var totalFound atomic.Int64

// 	start := time.Now()

// 	for word := range words {
// 		urls := buildTargetURLs(cfg, word)

// 		for _, u := range urls {
// 			wg.Add(1)
// 			sem <- struct{}{}

// 			go func(targetURL, path string) {
// 				defer wg.Done()
// 				defer func() { <-sem }()

// 				resp, err := client.Probe(targetURL)
// 				totalScanned.Add(1)

// 				if err != nil {
// 					if requester.IsConnectionRefused(err) {
// 						fmt.Fprintf(os.Stderr,
// 							"\n[FATAL] Connection refused — is %s up?\n",
// 							cfg.TargetURL,
// 						)
// 						os.Exit(1)
// 					}
// 					printer.PrintError(path, err)
// 					return
// 				}

// 				result := &Result{
// 					URL:         targetURL,
// 					Path:        path,
// 					StatusCode:  resp.StatusCode,
// 					Size:        resp.Size,
// 					RedirectURL: resp.RedirectURL,
// 					ContentType: resp.ContentType,
// 					Latency:     resp.Latency,
// 				}

// 				result.Severity = result.classifySeverity()
// 				result.Tags = result.generateTags()

// 				if !filter.Matches(result) {
// 					return
// 				}

// 				totalFound.Add(1)

// 				printer.PrintResult(output.Result{
// 					URL:        result.URL,
// 					Path:       result.Path,
// 					StatusCode: result.StatusCode,
// 					Size:       result.Size,
// 					Redirected: result.RedirectURL,
// 				})

// 			}(u, word)
// 		}

// 		if cfg.Verbose {
// 			printer.PrintProgress(int(totalScanned.Load()), 0)
// 		}
// 	}

// 	wg.Wait()
// 	printer.PrintSummary(time.Since(start))
// }

// // buildTargetURLs returns all URL variants for a given word
// func buildTargetURLs(cfg Config, word string) []string {
// 	if len(cfg.Extensions) == 0 {
// 		return []string{requester.BuildURL(cfg.TargetURL, word)}
// 	}
// 	return requester.BuildURLWithExtensions(cfg.TargetURL, word, cfg.Extensions)
// }

package fuzzer

import (
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/shubhcoding01/swiftprobe/internal/output"
	"github.com/shubhcoding01/swiftprobe/internal/requester"
	"github.com/shubhcoding01/swiftprobe/internal/wordlist"
)

// Config holds all scan parameters passed in from main.go
type Config struct {
	TargetURL       string
	WordlistPath    string
	Threads         int
	TimeoutSecs     int
	MatchCodes      string
	ExcludeCodes    string
	Extensions      []string
	UserAgent       string
	Headers         map[string]string
	FollowRedirects bool
	Verbose         bool
	OutputFile      string
}

// Run is the main entry point — orchestrates the full scan
func Run(cfg Config) {
	if cfg.Threads <= 0 {
		cfg.Threads = 50
	}
	if cfg.TimeoutSecs <= 0 {
		cfg.TimeoutSecs = 5
	}
	if cfg.MatchCodes == "" {
		cfg.MatchCodes = "200,201,301,302,401,403,500"
	}

	printer := output.New(cfg.Verbose)
	output.Banner()
	output.PrintConfig(
		cfg.TargetURL,
		cfg.WordlistPath,
		cfg.MatchCodes,
		cfg.Threads,
		cfg.TimeoutSecs,
	)

	filter := NewResultFilter(cfg.MatchCodes, cfg.ExcludeCodes)

	client := requester.New(requester.Config{
		TimeoutSecs:     cfg.TimeoutSecs,
		UserAgent:       cfg.UserAgent,
		Headers:         cfg.Headers,
		FollowRedirects: cfg.FollowRedirects,
	})

	words, err := wordlist.Stream(cfg.WordlistPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[ERROR] Could not open wordlist: %v\n", err)
		os.Exit(1)
	}

	sem := make(chan struct{}, cfg.Threads)
	var wg sync.WaitGroup
	var totalScanned atomic.Int64
	var totalFound atomic.Int64

	start := time.Now()

	for word := range words {
		urls := buildTargetURLs(cfg, word)

		for _, u := range urls {
			wg.Add(1)
			sem <- struct{}{}

			go func(targetURL, path string) {
				defer wg.Done()
				defer func() { <-sem }()

				resp, err := client.Probe(targetURL)
				totalScanned.Add(1)

				if err != nil {
					if requester.IsConnectionRefused(err) {
						fmt.Fprintf(os.Stderr,
							"\n[FATAL] Connection refused — is %s up?\n",
							cfg.TargetURL,
						)
						os.Exit(1)
					}
					printer.PrintError(path, err)
					return
				}

				result := &Result{
					URL:         targetURL,
					Path:        path,
					StatusCode:  resp.StatusCode,
					Size:        resp.Size,
					RedirectURL: resp.RedirectURL,
					ContentType: resp.ContentType,
					Latency:     resp.Latency,
				}

				result.Severity = result.classifySeverity()
				result.Tags = result.generateTags()

				if !filter.Matches(result) {
					return
				}

				totalFound.Add(1)

				printer.PrintResult(output.Result{
					URL:        result.URL,
					Path:       result.Path,
					StatusCode: result.StatusCode,
					Size:       result.Size,
					Redirected: result.RedirectURL,
				})

			}(u, word)
		}

		if cfg.Verbose {
			printer.PrintProgress(int(totalScanned.Load()), 0)
		}
	}

	wg.Wait()
	printer.PrintSummary(time.Since(start))
}

// buildTargetURLs returns all URL variants for a given word
func buildTargetURLs(cfg Config, word string) []string {
	if len(cfg.Extensions) == 0 {
		return []string{requester.BuildURL(cfg.TargetURL, word)}
	}
	return requester.BuildURLWithExtensions(cfg.TargetURL, word, cfg.Extensions)
}
