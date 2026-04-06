package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"regexp"
	"sort"
	"strings"
)

func main() {
	aVal := flag.Int("A", -1, "after context")
	bVal := flag.Int("B", -1, "before context")
	cVal := flag.Int("C", -1, "context before and after")
	count := flag.Bool("c", false, "count matching lines")
	ignoreCase := flag.Bool("i", false, "ignore case")
	invert := flag.Bool("v", false, "invert match")
	fixed := flag.Bool("F", false, "fixed string")
	number := flag.Bool("n", false, "line numbers")

	flag.Parse()

	args := flag.Args()
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "grep: pattern required")
		os.Exit(2)
	}

	pattern := args[0]
	var path string
	if len(args) > 1 {
		path = args[1]
	}

	after, before := 0, 0
	if *cVal >= 0 {
		after, before = *cVal, *cVal
	}
	if *aVal >= 0 {
		after = *aVal
	}
	if *bVal >= 0 {
		before = *bVal
	}

	lines, err := readInp(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "grep: %v\n", err)
		os.Exit(2)
	}

	m, err := newMatch(pattern, *fixed, *ignoreCase, *invert)
	if err != nil {
		fmt.Fprintf(os.Stderr, "grep: invalid pattern: %v\n", err)
		os.Exit(2)
	}

	if *count {
		n := 0
		for _, line := range lines {
			if m.match(line) {
				n++
			}
		}
		fmt.Println(n)
		if n == 0 {
			os.Exit(1)
		}
		return
	}

	if after == 0 && before == 0 {
		exit := runSimple(lines, m, *number)
		os.Exit(exit)
	}

	exit := runWithContext(lines, m, before, after, *number)
	os.Exit(exit)
}

func readInp(path string) ([]string, error) {
	var r io.Reader = os.Stdin
	if path != "" {
		f, err := os.Open(path)
		if err != nil {
			return nil, err
		}
		defer f.Close()
		r = f
	}
	sc := bufio.NewScanner(r)
	var lines []string
	for sc.Scan() {
		line := strings.TrimRight(sc.Text(), "\r")
		lines = append(lines, line)
	}
	return lines, sc.Err()
}

type matcher struct {
	re         *regexp.Regexp
	pattern    string
	fixed      bool
	ignoreCase bool
	invert     bool
}

func newMatch(pattern string, fixed, ignoreCase, invert bool) (*matcher, error) {
	m := &matcher{
		pattern:    pattern,
		fixed:      fixed,
		ignoreCase: ignoreCase,
		invert:     invert,
	}
	if fixed {
		return m, nil
	}
	expr := pattern
	if ignoreCase {
		expr = "(?i)" + pattern
	}
	re, err := regexp.Compile(expr)
	if err != nil {
		return nil, err
	}
	m.re = re
	return m, nil
}

func (m *matcher) match(line string) bool {
	var ok bool
	if m.fixed {
		if m.ignoreCase {
			ok = strings.Contains(strings.ToLower(line), strings.ToLower(m.pattern))
		} else {
			ok = strings.Contains(line, m.pattern)
		}
	} else {
		ok = m.re.MatchString(line)
	}
	if m.invert {
		return !ok
	}
	return ok
}

func runSimple(lines []string, m *matcher, number bool) int {
	found := false
	for i, line := range lines {
		if !m.match(line) {
			continue
		}
		found = true
		if number {
			fmt.Printf("%d:%s\n", i+1, line)
		} else {
			fmt.Println(line)
		}
	}
	if !found {
		return 1
	}
	return 0
}

func runWithContext(lines []string, m *matcher, before, after int, number bool) int {
	type span struct{ lo, hi int }
	var spans []span
	for i, line := range lines {
		if !m.match(line) {
			continue
		}
		lo := i - before
		if lo < 0 {
			lo = 0
		}
		hi := i + after
		if hi >= len(lines) {
			hi = len(lines) - 1
		}
		spans = append(spans, span{lo, hi})
	}
	if len(spans) == 0 {
		return 1
	}

	sort.Slice(spans, func(i, j int) bool {
		if spans[i].lo != spans[j].lo {
			return spans[i].lo < spans[j].lo
		}
		return spans[i].hi < spans[j].hi
	})

	merged := []span{spans[0]}
	for k := 1; k < len(spans); k++ {
		last := &merged[len(merged)-1]
		if spans[k].lo <= last.hi+1 {
			if spans[k].hi > last.hi {
				last.hi = spans[k].hi
			}
		} else {
			merged = append(merged, spans[k])
		}
	}

	for gi, s := range merged {
		if gi > 0 {
			prev := merged[gi-1]
			if prev.hi+1 < s.lo {
				fmt.Println("--")
			}
		}
		for idx := s.lo; idx <= s.hi; idx++ {
			line := lines[idx]
			isMatch := m.match(line)
			if number {
				if isMatch {
					fmt.Printf("%d:%s\n", idx+1, line)
				} else {
					fmt.Printf("%d-%s\n", idx+1, line)
				}
			} else {
				fmt.Println(line)
			}
		}
	}
	return 0
}
