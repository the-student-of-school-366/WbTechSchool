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
	Fields    string // -f: номера полей (колонок) для вывода
	Delimiter string // -d: разделитель полей
	Separated bool   // -s: только строки с разделителем
}

type FieldRange struct {
	Start int
	End   int
}

func main() {
	config := parseFlags()

	fields, err := parseFields(config.Fields)
	if err != nil {
		fmt.Fprintf(os.Stderr, "cut: %v\n", err)
		os.Exit(1)
	}

	if err := processInput(flag.Args(), config, fields); err != nil {
		fmt.Fprintf(os.Stderr, "cut: %v\n", err)
		os.Exit(1)
	}
}

func parseFlags() Config {
	var config Config

	flag.StringVar(&config.Fields, "f", "", "select only these fields (e.g., 1,3-5)")
	flag.StringVar(&config.Delimiter, "d", "\t", "use DELIM instead of TAB for field delimiter")
	flag.BoolVar(&config.Separated, "s", false, "do not print lines not containing delimiters")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [OPTIONS] [FILE]...\n", os.Args[0])
		fmt.Fprintln(os.Stderr, "\nExtract fields from FILE(s) or stdin.\n")
		fmt.Fprintln(os.Stderr, "Options:")
		flag.PrintDefaults()
	}

	flag.Parse()

	if config.Fields == "" {
		fmt.Fprintln(os.Stderr, "cut: you must specify a list of fields")
		flag.Usage()
		os.Exit(1)
	}

	return config
}

func parseFields(fieldsStr string) ([]int, error) {
	if fieldsStr == "" {
		return nil, fmt.Errorf("no fields specified")
	}

	var ranges []FieldRange
	parts := strings.Split(fieldsStr, ",")

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		r, err := parseFieldPart(part)
		if err != nil {
			return nil, err
		}
		ranges = append(ranges, r)
	}

	return expandRanges(ranges), nil
}

func parseFieldPart(part string) (FieldRange, error) {
	if strings.Contains(part, "-") {
		return parseRange(part)
	}

	num, err := strconv.Atoi(part)
	if err != nil || num < 1 {
		return FieldRange{}, fmt.Errorf("invalid field: %s", part)
	}

	return FieldRange{Start: num, End: num}, nil
}

func parseRange(part string) (FieldRange, error) {
	parts := strings.SplitN(part, "-", 2)

	start, err := strconv.Atoi(parts[0])
	if err != nil || start < 1 {
		return FieldRange{}, fmt.Errorf("invalid range: %s", part)
	}

	end, err := strconv.Atoi(parts[1])
	if err != nil || end < 1 {
		return FieldRange{}, fmt.Errorf("invalid range: %s", part)
	}

	if start > end {
		return FieldRange{}, fmt.Errorf("invalid decreasing range: %s", part)
	}

	return FieldRange{Start: start, End: end}, nil
}

func expandRanges(ranges []FieldRange) []int {
	fieldSet := make(map[int]bool)

	for _, r := range ranges {
		for i := r.Start; i <= r.End; i++ {
			fieldSet[i] = true
		}
	}

	fields := make([]int, 0, len(fieldSet))
	for f := range fieldSet {
		fields = append(fields, f)
	}

	sort.Ints(fields)

	return fields
}

func processInput(files []string, config Config, fields []int) error {
	writer := bufio.NewWriter(os.Stdout)
	defer writer.Flush()

	if len(files) == 0 {
		return processReader(os.Stdin, config, fields, writer)
	}

	for _, filename := range files {
		if err := processFile(filename, config, fields, writer); err != nil {
			return err
		}
	}

	return nil
}

func processFile(filename string, config Config, fields []int, writer *bufio.Writer) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	return processReader(file, config, fields, writer)
}

func processReader(r io.Reader, config Config, fields []int, writer *bufio.Writer) error {
	scanner := bufio.NewScanner(r)

	const maxCapacity = 1024 * 1024
	buf := make([]byte, 64*1024)
	scanner.Buffer(buf, maxCapacity)

	for scanner.Scan() {
		line := scanner.Text()
		processLine(line, config, fields, writer)
	}

	return scanner.Err()
}

func processLine(line string, config Config, fields []int, writer *bufio.Writer) {
	hasDelimiter := strings.Contains(line, config.Delimiter)

	if !hasDelimiter {
		if config.Separated {
			return
		}
		fmt.Fprintln(writer, line)
		return
	}

	columns := strings.Split(line, config.Delimiter)
	output := extractFields(columns, fields)

	fmt.Fprintln(writer, strings.Join(output, config.Delimiter))
}

func extractFields(columns []string, fields []int) []string {
	result := make([]string, 0, len(fields))

	for _, f := range fields {
		idx := f - 1
		if idx >= 0 && idx < len(columns) {
			result = append(result, columns[idx])
		}
	}

	return result
}
