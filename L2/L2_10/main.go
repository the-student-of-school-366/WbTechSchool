package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"
)

type Config struct {
	Column       int    // -k N: sort by column N (1-indexed)
	Numeric      bool   // -n: compare according to string numerical value
	Reverse      bool   // -r: reverse the result of comparisons
	Unique       bool   // -u: output only unique lines
	HumanNumeric bool   // -h: compare human readable numbers (e.g., 2K 1G)
	Separator    string // -t: field separator (default: tab)
}

var humanSuffixes = map[byte]float64{
	'K': 1024,
	'k': 1024,
	'M': 1024 * 1024,
	'm': 1024 * 1024,
	'G': 1024 * 1024 * 1024,
	'g': 1024 * 1024 * 1024,
	'T': 1024 * 1024 * 1024 * 1024,
	't': 1024 * 1024 * 1024 * 1024,
	'P': 1024 * 1024 * 1024 * 1024 * 1024,
	'p': 1024 * 1024 * 1024 * 1024 * 1024,
}

func main() {
	config := parseFlags()

	lines, err := readLines(flag.Args())
	if err != nil {
		fmt.Fprintf(os.Stderr, "sort: %v\n", err)
		os.Exit(1)
	}

	sortLines(lines, config)

	if config.Unique {
		lines = uniqueLines(lines, config)
	}

	writer := bufio.NewWriter(os.Stdout)
	defer writer.Flush()

	for _, line := range lines {
		fmt.Fprintln(writer, line)
	}
}

func parseFlags() Config {
	var config Config

	flag.IntVar(&config.Column, "k", 0, "sort by column N (1-indexed)")
	flag.BoolVar(&config.Numeric, "n", false, "compare according to string numerical value")
	flag.BoolVar(&config.Reverse, "r", false, "reverse the result of comparisons")
	flag.BoolVar(&config.Unique, "u", false, "output only unique lines")
	flag.BoolVar(&config.HumanNumeric, "h", false, "compare human readable numbers (e.g., 2K 1G)")
	flag.StringVar(&config.Separator, "t", "\t", "field separator")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [OPTIONS] [FILE]...\n", os.Args[0])
		fmt.Fprintln(os.Stderr, "\nSort lines from FILE(s) or stdin.\n")
		fmt.Fprintln(os.Stderr, "Options:")
		flag.PrintDefaults()
	}

	flag.Parse()

	return config
}

func readLines(files []string) ([]string, error) {
	if len(files) == 0 {
		return readFromReader(os.Stdin)
	}

	allLines := make([]string, 0, 1024)

	for _, filename := range files {
		lines, err := readFromFile(filename)
		if err != nil {
			return nil, err
		}
		allLines = append(allLines, lines...)
	}

	return allLines, nil
}

func readFromFile(filename string) ([]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	return readFromReader(file)
}

func readFromReader(r io.Reader) ([]string, error) {
	var lines []string
	scanner := bufio.NewScanner(r)

	const maxCapacity = 1024 * 1024
	buf := make([]byte, 64*1024)
	scanner.Buffer(buf, maxCapacity)

	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	return lines, scanner.Err()
}

func getKey(line string, config Config) string {
	key := line

	if config.Column > 0 {
		columns := strings.Split(line, config.Separator)
		if config.Column <= len(columns) {
			key = columns[config.Column-1]
		} else {
			key = ""
		}
	}

	return key
}

func compare(a, b string, config Config) int {
	keyA := getKey(a, config)
	keyB := getKey(b, config)

	var result int

	switch {
	case config.HumanNumeric:
		result = compareHumanNumeric(keyA, keyB)
	case config.Numeric:
		result = compareNumeric(keyA, keyB)
	default:
		result = strings.Compare(keyA, keyB)
	}

	if config.Reverse {
		result = -result
	}

	return result
}

func compareNumeric(a, b string) int {
	numA, errA := parseNumber(a)
	numB, errB := parseNumber(b)

	if errA != nil && errB != nil {
		return strings.Compare(a, b)
	}
	if errA != nil {
		return -1
	}
	if errB != nil {
		return 1
	}

	return compareFloats(numA, numB)
}

func parseNumber(s string) (float64, error) {
	return strconv.ParseFloat(strings.TrimSpace(s), 64)
}

func compareFloats(a, b float64) int {
	switch {
	case a < b:
		return -1
	case a > b:
		return 1
	default:
		return 0
	}
}

func compareHumanNumeric(a, b string) int {
	numA, errA := parseHumanNumber(a)
	numB, errB := parseHumanNumber(b)

	if errA != nil && errB != nil {
		return strings.Compare(a, b)
	}
	if errA != nil {
		return -1
	}
	if errB != nil {
		return 1
	}

	return compareFloats(numA, numB)
}

func parseHumanNumber(s string) (float64, error) {
	s = strings.TrimSpace(s)
	if len(s) == 0 {
		return 0, fmt.Errorf("empty string")
	}

	lastChar := s[len(s)-1]
	if multiplier, ok := humanSuffixes[lastChar]; ok {
		numStr := s[:len(s)-1]
		num, err := strconv.ParseFloat(numStr, 64)
		if err != nil {
			return 0, err
		}
		return num * multiplier, nil
	}

	return strconv.ParseFloat(s, 64)
}

func sortLines(lines []string, config Config) {
	sort.SliceStable(lines, func(i, j int) bool {
		return compare(lines[i], lines[j], config) < 0
	})
}

func keysEqual(a, b string, config Config) bool {
	keyA := getKey(a, config)
	keyB := getKey(b, config)

	switch {
	case config.HumanNumeric:
		return compareHumanNumeric(keyA, keyB) == 0
	case config.Numeric:
		return compareNumeric(keyA, keyB) == 0
	default:
		return keyA == keyB
	}
}

func uniqueLines(lines []string, config Config) []string {
	if len(lines) == 0 {
		return lines
	}

	result := make([]string, 0, len(lines))
	result = append(result, lines[0])

	for i := 1; i < len(lines); i++ {
		if !keysEqual(lines[i], lines[i-1], config) {
			result = append(result, lines[i])
		}
	}

	return result
}
