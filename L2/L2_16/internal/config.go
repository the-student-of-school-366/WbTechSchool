package internal

import "time"

type Config struct {
	StartURL       string
	MaxDepth       int
	Workers        int
	OutputDir      string
	UserAgent      string
	RequestTimeout time.Duration
	RespectRobots  bool
}
