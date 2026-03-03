package output

import (
	"fmt"
	"sync"
	"time"

	"github.com/fatih/color"
)

// Result holds the data for a single probe result
type Result struct {
	URL        string
	Path       string
	StatusCode int
	Size       int64
	Redirected string // redirect location if any
}

// Printer handles all terminal output in a thread-safe way
type Printer struct {
	mu      sync.Mutex
	verbose bool
	total   int
	found   int
}

// New creates a new Printer instance
func New(verbose bool) *Printer {
	return &Printer{verbose: verbose}
}

// color definitions
var (
	green   = color.New(color.FgGreen, color.Bold).SprintFunc()
	yellow  = color.New(color.FgYellow, color.Bold).SprintFunc()
	red     = color.New(color.FgRed, color.Bold).SprintFunc()
	cyan    = color.New(color.FgCyan, color.Bold).SprintFunc()
	magenta = color.New(color.FgMagenta, color.Bold).SprintFunc()
	white   = color.New(color.FgWhite, color.Bold).SprintFunc()
	gray    = color.New(color.FgHiBlack).SprintFunc()
)

// Banner prints the SwiftProbe ASCII banner on startup
func Banner() {
	fmt.Println(cyan(`
 ____          _  __ _   ____            _
/ ___|_      _(_)/ _| |_|  _ \ _ __ ___ | |__   ___
\___ \ \ /\ / / | |_| __| |_) | '__/ _ \| '_ \ / _ \
 ___) \ V  V /| |  _| |_|  __/| | | (_) | |_) |  __/
|____/ \_/\_/ |_|_|  \__|_|   |_|  \___/|_.__/ \___|
`))
	fmt.Println(gray("  High-Speed Directory Fuzzer | github.com/yourname/swiftprobe"))
	fmt.Println()
}

// PrintConfig displays the scan configuration before starting
func PrintConfig(url, wordlist, codes string, threads, timeout int) {
	fmt.Println(white("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Printf("  %s  %s\n", cyan("Target   :"), url)
	fmt.Printf("  %s  %s\n", cyan("Wordlist :"), wordlist)
	fmt.Printf("  %s  %d\n", cyan("Threads  :"), threads)
	fmt.Printf("  %s  %ds\n", cyan("Timeout  :"), timeout)
	fmt.Printf("  %s  %s\n", cyan("Codes    :"), codes)
	fmt.Printf("  %s  %s\n", cyan("Started  :"), time.Now().Format("2006-01-02 15:04:05"))
	fmt.Println(white("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Println()
}

// PrintResult prints a single found result in a color-coded format (thread-safe)
func (p *Printer) PrintResult(r Result) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.total++
	p.found++

	code := colorizeCode(r.StatusCode)
	size := formatSize(r.Size)

	line := fmt.Sprintf("  %s  %-45s  %s", code, "/"+r.Path, gray(size))

	// append redirect location if available
	if r.Redirected != "" {
		line += fmt.Sprintf("  %s %s", gray("→"), yellow(r.Redirected))
	}

	fmt.Println(line)
}

// PrintError prints an error message when verbose mode is on
func (p *Printer) PrintError(path string, err error) {
	if !p.verbose {
		return
	}
	p.mu.Lock()
	defer p.mu.Unlock()

	p.total++
	fmt.Printf("  %s  /%s  %s\n", gray("[ERR]"), path, gray(err.Error()))
}

// PrintProgress prints a live progress update (thread-safe)
func (p *Printer) PrintProgress(scanned, total int) {
	if !p.verbose {
		return
	}
	p.mu.Lock()
	defer p.mu.Unlock()

	pct := 0
	if total > 0 {
		pct = (scanned * 100) / total
	}
	fmt.Printf("\r  %s  Scanned: %d/%d (%d%%)   ", cyan("[~]"), scanned, total, pct)
}

// PrintSummary prints the final scan summary
func (p *Printer) PrintSummary(elapsed time.Duration) {
	p.mu.Lock()
	defer p.mu.Unlock()

	fmt.Println()
	fmt.Println(white("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Printf("  %s  %s\n", cyan("Finished :"), time.Now().Format("2006-01-02 15:04:05"))
	fmt.Printf("  %s  %s\n", cyan("Duration :"), elapsed.Round(time.Millisecond))
	fmt.Printf("  %s  %d\n", cyan("Total    :"), p.total)
	fmt.Printf("  %s  %s\n", cyan("Found    :"), green(fmt.Sprintf("%d", p.found)))
	fmt.Println(white("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
}

// colorizeCode returns a color-coded string for the HTTP status code
func colorizeCode(code int) string {
	label := fmt.Sprintf("[%d]", code)
	switch {
	case code == 200:
		return green(label)
	case code == 201 || code == 204:
		return green(label)
	case code == 301 || code == 302 || code == 307 || code == 308:
		return yellow(label)
	case code == 401:
		return magenta(label)
	case code == 403:
		return red(label)
	case code == 404:
		return gray(label)
	case code == 500:
		return red(label)
	default:
		return white(label)
	}
}

// formatSize returns a human-readable size string
func formatSize(size int64) string {
	if size < 0 {
		return "[size: unknown]"
	}
	if size < 1024 {
		return fmt.Sprintf("[size: %dB]", size)
	}
	if size < 1024*1024 {
		return fmt.Sprintf("[size: %.1fKB]", float64(size)/1024)
	}
	return fmt.Sprintf("[size: %.1fMB]", float64(size)/(1024*1024))
}