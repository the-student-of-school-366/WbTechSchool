package internal

import (
	"fmt"
	"log"
	"net/http"
	"net/url"

	"github.com/temoto/robotstxt"
)

type Checker struct {
	data      *robotstxt.RobotsData
	userAgent string
	enabled   bool
}

func NewChecker(baseURL *url.URL, userAgent string, respectRobots bool) *Checker {
	checker := &Checker{
		userAgent: userAgent,
		enabled:   respectRobots,
	}

	if !respectRobots {
		return checker
	}

	robotsURL := fmt.Sprintf("%s://%s/robots.txt", baseURL.Scheme, baseURL.Host)

	resp, err := http.Get(robotsURL)
	if err != nil {
		log.Printf("[WARN] Could not fetch robots.txt: %v", err)
		return checker
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("[WARN] robots.txt returned status %d", resp.StatusCode)
		return checker
	}

	data, err := robotstxt.FromResponse(resp)
	if err != nil {
		log.Printf("[WARN] Could not parse robots.txt: %v", err)
		return checker
	}

	checker.data = data
	return checker
}

func (c *Checker) IsAllowed(targetURL string) bool {
	if !c.enabled || c.data == nil {
		return true
	}

	return c.data.TestAgent(targetURL, c.userAgent)
}
