package main

import (
	"L2_16/internal"
	"flag"
	"fmt"
	"log"
	"net/url"
	"time"
)

func main() {
	cfg := parseFlags()

	baseURL, err := url.Parse(cfg.StartURL)
	if err != nil {
		log.Fatalf("Invalid start URL %q: %v", cfg.StartURL, err)
	}

	startTime := time.Now()

	site := internal.NewCrawler(cfg, baseURL)
	site.Run()

	fmt.Printf("Download complete in %s\n", time.Since(startTime).Round(time.Millisecond))
}

func parseFlags() internal.Config {
	startURL := flag.String("url", "", "URL to download (required)")
	maxDepth := flag.Int("depth", 2, "Maximum recursion depth for following links")
	workers := flag.Int("workers", 8, "Number of concurrent download workers")
	outputDir := flag.String("output", "site", "Output directory for downloaded files")
	userAgent := flag.String("user-agent", "VSiteDownloader", "User-Agent string for HTTP requests")
	requestTimeout := flag.Duration("timeout", 30*time.Second, "Timeout for each HTTP request")
	respectRobots := flag.Bool("robots", true, "Respect robots.txt rules")

	flag.Parse()

	if *startURL == "" {
		if flag.NArg() > 0 {
			*startURL = flag.Arg(0)
		} else {
			log.Fatal("Usage: wget -url <URL> [options]\n  or:  wget <URL> [options]")
		}
	}

	return internal.Config{
		StartURL:       *startURL,
		MaxDepth:       *maxDepth,
		Workers:        *workers,
		OutputDir:      *outputDir,
		UserAgent:      *userAgent,
		RequestTimeout: *requestTimeout,
		RespectRobots:  *respectRobots,
	}
}
