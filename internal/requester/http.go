// package requester

// import (
//     "fmt"
//     "net/http"
//     "time"
// )

// type Client struct {
//     http *http.Client
// }

// func New(timeoutSecs int) *Client {
//     return &Client{
//         http: &http.Client{
//             Timeout: time.Duration(timeoutSecs) * time.Second,
//             CheckRedirect: func(req *http.Request, via []*http.Request) error {
//                 return http.ErrUseLastResponse // don't follow redirects
//             },
//         },
//     }
// }

// type Result struct {
//     URL        string
//     StatusCode int
//     Size       int64
// }

// func (c *Client) Probe(url string) (*Result, error) {
//     resp, err := c.http.Get(url)
//     if err != nil {
//         return nil, err
//     }
//     defer resp.Body.Close()

//     return &Result{
//         URL:        url,
//         StatusCode: resp.StatusCode,
//         Size:       resp.ContentLength,
//     }, nil
// }

// func BuildURL(base, path string) string {
//     return fmt.Sprintf("%s/%s", base, path)
// }


package requester

import (
	"fmt"
	"net/http"
	"strings"
	"time"
)

// Client wraps the standard http.Client with fuzzer-specific config
type Client struct {
	http      *http.Client
	userAgent string
	headers   map[string]string
}

// Config holds optional settings for the HTTP client
type Config struct {
	TimeoutSecs int
	UserAgent   string
	Headers     map[string]string // custom headers e.g. Authorization
	FollowRedirects bool
}

// Result holds the full response data from a single probe
type Result struct {
	URL         string
	StatusCode  int
	Size        int64
	RedirectURL string
	Latency     time.Duration
	ContentType string
	Server      string // e.g. "Apache", "nginx" — useful recon info
}

// New creates a Client with sane fuzzer defaults
func New(cfg Config) *Client {
	if cfg.TimeoutSecs <= 0 {
		cfg.TimeoutSecs = 5
	}
	if cfg.UserAgent == "" {
		cfg.UserAgent = "Mozilla/5.0 (compatible; SwiftProbe/1.0)"
	}

	redirectPolicy := func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse // capture redirects, don't follow
	}
	if cfg.FollowRedirects {
		redirectPolicy = nil // use default Go behavior
	}

	return &Client{
		userAgent: cfg.UserAgent,
		headers:   cfg.Headers,
		http: &http.Client{
			Timeout:       time.Duration(cfg.TimeoutSecs) * time.Second,
			CheckRedirect: redirectPolicy,
			Transport: &http.Transport{
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 100,   // critical for high concurrency
				IdleConnTimeout:     30 * time.Second,
				DisableKeepAlives:   false, // reuse connections = faster
			},
		},
	}
}

// Probe sends a GET request to the given URL and returns a Result
func (c *Client) Probe(url string) (*Result, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to build request: %w", err)
	}

	// set user agent
	req.Header.Set("User-Agent", c.userAgent)

	// set default Accept header
	req.Header.Set("Accept", "*/*")

	// apply any custom headers (e.g. Authorization, Cookie)
	for key, val := range c.headers {
		req.Header.Set(key, val)
	}

	// fire the request and track latency
	start := time.Now()
	resp, err := c.http.Do(req)
	latency := time.Since(start)

	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	return &Result{
		URL:         url,
		StatusCode:  resp.StatusCode,
		Size:        resp.ContentLength,
		RedirectURL: resp.Header.Get("Location"),
		Latency:     latency,
		ContentType: resp.Header.Get("Content-Type"),
		Server:      resp.Header.Get("Server"),
	}, nil
}

// BuildURL safely joins a base URL and a path, avoiding double slashes
func BuildURL(base, path string) string {
	base = strings.TrimRight(base, "/")
	path = strings.TrimLeft(path, "/")
	return fmt.Sprintf("%s/%s", base, path)
}

// BuildURLWithExtensions returns multiple URLs with different extensions appended
// e.g. path="admin", exts=[".php",".html"] → ["/admin", "/admin.php", "/admin.html"]
func BuildURLWithExtensions(base, path string, extensions []string) []string {
	base = strings.TrimRight(base, "/")
	path = strings.TrimLeft(path, "/")

	// always include the bare path
	urls := []string{fmt.Sprintf("%s/%s", base, path)}

	for _, ext := range extensions {
		if !strings.HasPrefix(ext, ".") {
			ext = "." + ext
		}
		urls = append(urls, fmt.Sprintf("%s/%s%s", base, path, ext))
	}

	return urls
}

// IsTimeout checks if an error was caused by a request timeout
func IsTimeout(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "context deadline exceeded") ||
		strings.Contains(err.Error(), "timeout")
}

// IsConnectionRefused checks if the target actively refused the connection
func IsConnectionRefused(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "connection refused")
}
