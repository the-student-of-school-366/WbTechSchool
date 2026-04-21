package internal

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

type Fetcher struct {
	client    *http.Client
	userAgent string
}

func NewFetcher(userAgent string, timeout time.Duration) *Fetcher {
	return &Fetcher{
		client: &http.Client{
			Timeout: timeout,
		},
		userAgent: userAgent,
	}
}

func (f *Fetcher) FetchPage(pageURL string) ([]byte, error) {
	return f.fetch(pageURL)
}

func (f *Fetcher) FetchRaw(resourceURL string) ([]byte, error) {
	return f.fetch(resourceURL)
}

func (f *Fetcher) fetch(targetURL string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, targetURL, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("User-Agent", f.userAgent)

	resp, err := f.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d for %s", resp.StatusCode, targetURL)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %w", err)
	}

	return body, nil
}
