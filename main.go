// zsuggest ranks shell history by frequency, directory context and recency.
//
// It reads the history log written by the zshaddhistory hook in ~/.zshrc,
// where each record is "<command>⋮<pwd>   " (multi-line commands span lines).
//
// Usage:
//
//	zsuggest top  [--dir DIR] [--file FILE] -- PREFIX
//	zsuggest list [--dir DIR] [--file FILE] [-n N] -- PREFIX
package main

import (
	"bufio"
	"flag"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const sep = "⋮"

type record struct {
	cmd string
	dir string
}

type candidate struct {
	cmd       string
	score     float64
	hereCount int
	total     int
	lastSeen  int
}

// Scoring weights. A use in the current directory counts far more than a
// use elsewhere; uses in parent/child directories (same project) sit between.
const (
	weightSameDir    = 1.0
	weightRelatedDir = 0.35
	weightOtherDir   = 0.08
	recencyHalfLife  = 2500.0 // commands, not time
	recencyFloor     = 0.25   // old commands never drop below this share
)

func main() {
	if len(os.Args) < 2 {
		usage()
	}
	mode := os.Args[1]
	fs := flag.NewFlagSet(mode, flag.ExitOnError)
	cwd, _ := os.Getwd()
	dir := fs.String("dir", cwd, "directory context")
	file := fs.String("file", defaultHistFile(), "history log")
	n := fs.Int("n", 15, "max results for list")
	fs.Parse(os.Args[2:])
	prefix := strings.Join(fs.Args(), " ")

	records, err := load(*file)
	if err != nil {
		os.Exit(1)
	}
	ranked := rank(records, prefix, filepath.Clean(*dir))

	switch mode {
	case "top":
		if strings.TrimSpace(prefix) != "" && len(ranked) > 0 {
			fmt.Print(ranked[0].cmd)
		}
	case "list":
		w := bufio.NewWriter(os.Stdout)
		defer w.Flush()
		for i, c := range ranked {
			if i >= *n {
				break
			}
			// NUL-separated so multi-line commands survive; tab splits label from command.
			fmt.Fprintf(w, "%s\t%s\x00", label(c), c.cmd)
		}
	default:
		usage()
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, "usage: zsuggest top|list [--dir DIR] [--file FILE] [-n N] -- PREFIX")
	os.Exit(2)
}

func defaultHistFile() string {
	if f := os.Getenv("ZSUGGEST_HISTFILE"); f != "" {
		return f
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".zsh_history_ext")
}

func label(c candidate) string {
	if c.hereCount > 0 {
		return fmt.Sprintf("★ %3d× here", c.hereCount)
	}
	return fmt.Sprintf("  %3d× all ", c.total)
}

func load(path string) ([]record, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return parse(bufio.NewScanner(f)), nil
}

func parse(sc *bufio.Scanner) []record {
	sc.Buffer(make([]byte, 64*1024), 4*1024*1024)
	var records []record
	var pending []string
	for sc.Scan() {
		line := sc.Text()
		i := strings.LastIndex(line, sep)
		if i < 0 {
			// Continuation of a multi-line command.
			pending = append(pending, line)
			continue
		}
		pending = append(pending, line[:i])
		cmd := strings.TrimSpace(strings.Join(pending, "\n"))
		pending = pending[:0]
		dir := strings.TrimRight(line[i+len(sep):], " ")
		if cmd == "" {
			continue
		}
		records = append(records, record{cmd: cmd, dir: dir})
	}
	return records
}

func dirWeight(cwd, dir string) (float64, bool) {
	switch {
	case dir == "":
		return weightOtherDir, false
	case dir == cwd:
		return weightSameDir, true
	case isWithin(cwd, dir) || isWithin(dir, cwd):
		return weightRelatedDir, false
	}
	return weightOtherDir, false
}

// isWithin reports whether child is strictly inside parent. The home
// directory and root are too broad to count as a shared project.
func isWithin(child, parent string) bool {
	home, _ := os.UserHomeDir()
	if parent == "/" || parent == home {
		return false
	}
	return strings.HasPrefix(child, parent+"/")
}

// rank orders the distinct commands starting with prefix. An empty prefix
// ranks everything, which the list menu uses to show a directory's top commands.
func rank(records []record, prefix, cwd string) []candidate {
	byCmd := map[string]*candidate{}
	last := len(records) - 1
	for i, r := range records {
		if r.cmd == prefix || !strings.HasPrefix(r.cmd, prefix) {
			continue
		}
		c := byCmd[r.cmd]
		if c == nil {
			c = &candidate{cmd: r.cmd}
			byCmd[r.cmd] = c
		}
		w, here := dirWeight(cwd, r.dir)
		decay := math.Pow(0.5, float64(last-i)/recencyHalfLife)
		c.score += w * (recencyFloor + (1-recencyFloor)*decay)
		c.total++
		c.lastSeen = i
		if here {
			c.hereCount++
		}
	}
	out := make([]candidate, 0, len(byCmd))
	for _, c := range byCmd {
		out = append(out, *c)
	}
	sort.Slice(out, func(a, b int) bool {
		if out[a].score != out[b].score {
			return out[a].score > out[b].score
		}
		return out[a].lastSeen > out[b].lastSeen
	})
	return out
}
