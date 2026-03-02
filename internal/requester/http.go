package requester

import (
    "fmt"
    "net/http"
    "time"
)

type Client struct {
    http *http.Client
}

func New(timeoutSecs int) *Client {
    return &Client{
        http: &http.Client{
            Timeout: time.Duration(timeoutSecs) * time.Second,
            CheckRedirect: func(req *http.Request, via []*http.Request) error {
                return http.ErrUseLastResponse // don't follow redirects
            },
        },
    }
}

type Result struct {
    URL        string
    StatusCode int
    Size       int64
}

func (c *Client) Probe(url string) (*Result, error) {
    resp, err := c.http.Get(url)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    return &Result{
        URL:        url,
        StatusCode: resp.StatusCode,
        Size:       resp.ContentLength,
    }, nil
}

func BuildURL(base, path string) string {
    return fmt.Sprintf("%s/%s", base, path)
}