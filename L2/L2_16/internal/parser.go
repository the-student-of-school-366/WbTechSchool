package internal

import (
	"bytes"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

var resourceSelectors = []struct {
	Selector  string
	Attribute string
}{
	{"img", "src"},
	{"script", "src"},
	{"link[rel='stylesheet']", "href"},
	{"link[rel='icon']", "href"},
}

var allLinkSelectors = []struct {
	Selector  string
	Attribute string
}{
	{"img", "src"},
	{"script", "src"},
	{"link[rel='stylesheet']", "href"},
	{"link[rel='icon']", "href"},
	{"a", "href"},
}

func ParseHTML(body []byte) (*goquery.Document, error) {
	return goquery.NewDocumentFromReader(bytes.NewReader(body))
}

func ExtractResourceURLs(doc *goquery.Document, pageURL string) []string {
	var urls []string

	for _, sel := range resourceSelectors {
		doc.Find(sel.Selector).Each(func(_ int, s *goquery.Selection) {
			val, exists := s.Attr(sel.Attribute)
			if !exists || val == "" {
				return
			}
			absoluteURL := ResolveURL(val, pageURL)
			urls = append(urls, absoluteURL)
		})
	}

	return urls
}

func ExtractLinkURLs(doc *goquery.Document, pageURL, baseURL string) []string {
	var urls []string

	doc.Find("a[href]").Each(func(_ int, s *goquery.Selection) {
		href, exists := s.Attr("href")
		if !exists || href == "" {
			return
		}

		absoluteURL := ResolveURL(href, pageURL)

		if !strings.HasPrefix(absoluteURL, baseURL) {
			return
		}

		urls = append(urls, absoluteURL)
	})

	return urls
}

func RewriteLinks(doc *goquery.Document, pageURL, baseURL, outputDir string) {
	for _, sel := range allLinkSelectors {
		doc.Find(sel.Selector).Each(func(_ int, s *goquery.Selection) {
			val, exists := s.Attr(sel.Attribute)
			if !exists || val == "" {
				return
			}

			absoluteURL := ResolveURL(val, pageURL)

			if !strings.HasPrefix(absoluteURL, baseURL) {
				return
			}

			localPath := URLToLocalPath(absoluteURL, outputDir)
			relativePath, err := filepath.Rel(outputDir, localPath)
			if err != nil {
				return
			}
			s.SetAttr(sel.Attribute, relativePath)
		})
	}
}

func ResolveURL(href, basePageURL string) string {
	u, err := url.Parse(href)
	if err != nil {
		return href
	}

	base, err := url.Parse(basePageURL)
	if err != nil {
		return href
	}

	return base.ResolveReference(u).String()
}

func URLToLocalPath(rawURL, outputDir string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return filepath.Join(outputDir, "unknown")
	}

	path := u.Path
	if path == "" || path == "/" {
		path = "/index.html"
	}
	if strings.HasSuffix(path, "/") {
		path += "index.html"
	}

	return filepath.Join(outputDir, path)
}
