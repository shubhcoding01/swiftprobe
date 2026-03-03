package fuzzer

import (
	"fmt"
	"net/http"
	"strings"
	"time"
)

// Severity represents how interesting a result is for security testing
type Severity int

const (
	SeverityNone     Severity = iota
	SeverityLow               // 301, 302 redirects
	SeverityMedium            // 403 forbidden (exists but blocked)
	SeverityHigh              // 200 OK, 401 unauthorized
	SeverityCritical          // 500 server errors (may leak info)
)

// Result holds all data captured from a single HTTP probe
type Result struct {
	// Request info
	URL       string
	Path      string
	Method    string
	Timestamp time.Time

	// Response info
	StatusCode  int
	StatusText  string
	Size        int64
	ContentType string
	RedirectURL string
	Headers     http.Header
	Latency     time.Duration

	// Analysis
	Severity  Severity
	Tags      []string // e.g. ["redirect", "auth-required", "large-response"]
	IsMatch   bool     // did it match the user's filter criteria?
}

// NewResult constructs a Result from a raw HTTP response
func NewResult(url, path string, resp *http.Response, latency time.Duration) *Result {
	r := &Result{
		URL:       url,
		Path:      path,
		Method:    "GET",
		Timestamp: time.Now(),
		Latency:   latency,
		IsMatch:   true,
	}

	if resp != nil {
		r.StatusCode = resp.StatusCode
		r.StatusText = http.StatusText(resp.StatusCode)
		r.Size = resp.ContentLength
		r.ContentType = resp.Header.Get("Content-Type")
		r.Headers = resp.Header
		r.RedirectURL = resp.Header.Get("Location")
	}

	r.Severity = r.classifySeverity()
	r.Tags = r.generateTags()

	return r
}

// NewErrorResult constructs a Result for a failed request (timeout, connection refused, etc.)
func NewErrorResult(url, path string, err error) *Result {
	return &Result{
		URL:       url,
		Path:      path,
		Method:    "GET",
		Timestamp: time.Now(),
		StatusCode: 0,
		StatusText: err.Error(),
		Severity:  SeverityNone,
		IsMatch:   false,
		Tags:      []string{"error"},
	}
}

// classifySeverity assigns a severity level based on HTTP status code
func (r *Result) classifySeverity() Severity {
	switch {
	case r.StatusCode == 500 || r.StatusCode == 503:
		return SeverityCritical // server errors often reveal stack traces
	case r.StatusCode == 200 || r.StatusCode == 201 || r.StatusCode == 204:
		return SeverityHigh
	case r.StatusCode == 401:
		return SeverityHigh // exists but needs auth — very interesting
	case r.StatusCode == 403:
		return SeverityMedium // exists but blocked
	case r.StatusCode == 301 || r.StatusCode == 302 ||
		r.StatusCode == 307 || r.StatusCode == 308:
		return SeverityLow
	default:
		return SeverityNone
	}
}

// generateTags produces a list of descriptive tags for filtering/output
func (r *Result) generateTags() []string {
	tags := []string{}

	// Status-based tags
	switch {
	case r.StatusCode == 200:
		tags = append(tags, "accessible")
	case r.StatusCode == 401:
		tags = append(tags, "auth-required")
	case r.StatusCode == 403:
		tags = append(tags, "forbidden")
	case r.StatusCode == 301 || r.StatusCode == 302:
		tags = append(tags, "redirect")
	case r.StatusCode >= 500:
		tags = append(tags, "server-error")
	}

	// Size-based tags
	if r.Size > 1024*1024 {
		tags = append(tags, "large-response") // > 1MB
	} else if r.Size == 0 {
		tags = append(tags, "empty-response")
	}

	// Content-type tags
	ct := strings.ToLower(r.ContentType)
	switch {
	case strings.Contains(ct, "json"):
		tags = append(tags, "json")
	case strings.Contains(ct, "xml"):
		tags = append(tags, "xml")
	case strings.Contains(ct, "html"):
		tags = append(tags, "html")
	case strings.Contains(ct, "text/plain"):
		tags = append(tags, "plaintext")
	}

	// Path-based heuristic tags
	path := strings.ToLower(r.Path)
	switch {
	case strings.Contains(path, "admin"):
		tags = append(tags, "admin-panel")
	case strings.Contains(path, "backup") || strings.Contains(path, ".bak"):
		tags = append(tags, "backup-file")
	case strings.Contains(path, "api"):
		tags = append(tags, "api-endpoint")
	case strings.Contains(path, "login") || strings.Contains(path, "signin"):
		tags = append(tags, "login-page")
	case strings.Contains(path, ".env") || strings.Contains(path, "config"):
		tags = append(tags, "config-file") // high value target
	case strings.Contains(path, ".git"):
		tags = append(tags, "git-exposure") // critical finding
	}

	// Redirect tag with destination
	if r.RedirectURL != "" {
		tags = append(tags, fmt.Sprintf("→ %s", r.RedirectURL))
	}

	// Latency tag
	if r.Latency > 3*time.Second {
		tags = append(tags, "slow")
	}

	return tags
}

// SeverityLabel returns a human-readable severity string
func (r *Result) SeverityLabel() string {
	switch r.Severity {
	case SeverityCritical:
		return "CRITICAL"
	case SeverityHigh:
		return "HIGH"
	case SeverityMedium:
		return "MEDIUM"
	case SeverityLow:
		return "LOW"
	default:
		return "NONE"
	}
}

// TagString returns all tags joined as a readable string
func (r *Result) TagString() string {
	if len(r.Tags) == 0 {
		return ""
	}
	return "[" + strings.Join(r.Tags, ", ") + "]"
}

// IsInteresting returns true if the result is worth reporting
func (r *Result) IsInteresting() bool {
	return r.Severity >= SeverityLow && r.StatusCode != 404
}

// Summary returns a one-line string summary of the result
func (r *Result) Summary() string {
	return fmt.Sprintf(
		"[%d %s] /%s | size: %s | latency: %dms | severity: %s | tags: %s",
		r.StatusCode,
		r.StatusText,
		r.Path,
		formatSize(r.Size),
		r.Latency.Milliseconds(),
		r.SeverityLabel(),
		r.TagString(),
	)
}

// formatSize converts bytes to human-readable string
func formatSize(size int64) string {
	switch {
	case size < 0:
		return "unknown"
	case size < 1024:
		return fmt.Sprintf("%dB", size)
	case size < 1024*1024:
		return fmt.Sprintf("%.1fKB", float64(size)/1024)
	default:
		return fmt.Sprintf("%.1fMB", float64(size)/(1024*1024))
	}
}

// ResultFilter defines criteria for matching results
type ResultFilter struct {
	MatchCodes   map[int]struct{}
	MinSize      int64
	MaxSize      int64
	ExcludeCodes map[int]struct{}
}

// NewResultFilter builds a filter from a comma-separated status code string
func NewResultFilter(matchCodes string, excludeCodes string) *ResultFilter {
	return &ResultFilter{
		MatchCodes:   parseCodeSet(matchCodes),
		ExcludeCodes: parseCodeSet(excludeCodes),
		MinSize:      -1, // disabled by default
		MaxSize:      -1, // disabled by default
	}
}

// Matches checks whether a Result passes all filter criteria
func (f *ResultFilter) Matches(r *Result) bool {
	// Must be in match list
	if len(f.MatchCodes) > 0 {
		if _, ok := f.MatchCodes[r.StatusCode]; !ok {
			return false
		}
	}

	// Must not be in exclude list
	if _, ok := f.ExcludeCodes[r.StatusCode]; ok {
		return false
	}

	// Size filters (if set)
	if f.MinSize >= 0 && r.Size < f.MinSize {
		return false
	}
	if f.MaxSize >= 0 && r.Size > f.MaxSize {
		return false
	}

	return true
}

// parseCodeSet converts "200,301,403" → map[int]struct{}
func parseCodeSet(raw string) map[int]struct{} {
	set := make(map[int]struct{})
	for _, s := range strings.Split(raw, ",") {
		s = strings.TrimSpace(s)
		var code int
		if _, err := fmt.Sscanf(s, "%d", &code); err == nil && code > 0 {
			set[code] = struct{}{}
		}
	}
	return set
}