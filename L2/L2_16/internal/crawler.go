package internal

import (
	"fmt"
	"log"
	"net/url"
	"strings"
	"sync"
)

type Task struct {
	URL   string
	Depth int
}

type Crawler struct {
	cfg     Config
	baseURL *url.URL
	fetcher *Fetcher
	storage *Storage
	robots  *Checker

	visited   map[string]bool
	visitedMu sync.Mutex

	tasks chan Task
	wg    sync.WaitGroup
}

func NewCrawler(cfg Config, baseURL *url.URL) *Crawler {
	return &Crawler{
		cfg:     cfg,
		baseURL: baseURL,
		fetcher: NewFetcher(cfg.UserAgent, cfg.RequestTimeout),
		storage: NewStorage(cfg.OutputDir),
		robots:  NewChecker(baseURL, cfg.UserAgent, cfg.RespectRobots),
		visited: make(map[string]bool),
		tasks:   make(chan Task, 1000),
	}
}

func (c *Crawler) Run() {
	for range c.cfg.Workers {
		c.wg.Add(1)
		go c.worker()
	}

	c.enqueue(Task{URL: c.cfg.StartURL, Depth: 0})

	c.wg.Wait()
	close(c.tasks)
}

func (c *Crawler) worker() {
	defer c.wg.Done()

	for task := range c.tasks {
		c.processPage(task)
	}
}

func (c *Crawler) enqueue(task Task) {
	if !c.markVisited(task.URL) {
		return
	}
	c.wg.Add(1)
	go func() {
		c.tasks <- task
	}()
}

func (c *Crawler) markVisited(rawURL string) bool {
	c.visitedMu.Lock()
	defer c.visitedMu.Unlock()

	if c.visited[rawURL] {
		return false
	}
	c.visited[rawURL] = true
	return true
}

func (c *Crawler) processPage(task Task) {
	defer c.wg.Done()

	if task.Depth > c.cfg.MaxDepth {
		return
	}

	if !c.robots.IsAllowed(task.URL) {
		log.Printf("[BLOCKED] robots.txt disallows: %s", task.URL)
		return
	}

	body, err := c.fetcher.FetchPage(task.URL)
	if err != nil {
		log.Printf("[ERROR] Failed to fetch page %s: %v", task.URL, err)
		return
	}

	doc, err := ParseHTML(body)
	if err != nil {
		log.Printf("[ERROR] Failed to parse HTML %s: %v", task.URL, err)
		return
	}

	RewriteLinks(doc, task.URL, c.baseURL.String(), c.cfg.OutputDir)

	localPath := c.storage.URLToLocalPath(task.URL)
	if err := c.storage.SaveHTML(doc, localPath); err != nil {
		log.Printf("[ERROR] Failed to save page %s: %v", task.URL, err)
		return
	}
	fmt.Printf("[PAGE] %s -> %s\n", task.URL, localPath)

	resourceURLs := ExtractResourceURLs(doc, task.URL)
	for _, resURL := range resourceURLs {
		go c.downloadResource(resURL)
	}

	if task.Depth < c.cfg.MaxDepth {
		linkURLs := ExtractLinkURLs(doc, task.URL, c.baseURL.String())
		for _, linkURL := range linkURLs {
			c.enqueue(Task{URL: linkURL, Depth: task.Depth + 1})
		}
	}
}

func (c *Crawler) downloadResource(resourceURL string) {
	if !c.markVisited(resourceURL) {
		return
	}

	if !c.robots.IsAllowed(resourceURL) {
		return
	}

	body, err := c.fetcher.FetchRaw(resourceURL)
	if err != nil {
		log.Printf("[ERROR] Failed to download resource %s: %v", resourceURL, err)
		return
	}

	localPath := c.storage.URLToLocalPath(resourceURL)
	if err := c.storage.SaveFile(localPath, body); err != nil {
		log.Printf("[ERROR] Failed to save resource %s: %v", resourceURL, err)
		return
	}

	fmt.Printf("[RESOURCE] %s -> %s\n", resourceURL, localPath)
}

func isSameDomain(rawURL, baseURL string) bool {
	return strings.HasPrefix(rawURL, baseURL)
}
