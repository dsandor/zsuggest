package main

import (
	"bufio"
	"strings"
	"testing"
)

func records(t *testing.T, s string) []record {
	t.Helper()
	return parse(bufio.NewScanner(strings.NewReader(s)))
}

func TestParseMultiLineAndBlank(t *testing.T) {
	got := records(t, "gs⋮/a   \n⋮/a   \nfor x in 1 2; do\necho $x\ndone⋮/b   \n")
	if len(got) != 2 {
		t.Fatalf("want 2 records, got %d: %#v", len(got), got)
	}
	if got[1].cmd != "for x in 1 2; do\necho $x\ndone" || got[1].dir != "/b" {
		t.Errorf("multi-line record parsed wrong: %#v", got[1])
	}
}

func TestDirectoryBeatsGlobalFrequency(t *testing.T) {
	var b strings.Builder
	for i := 0; i < 10; i++ {
		b.WriteString("sam deploy⋮/proj/other   \n")
	}
	b.WriteString("sam build --use-container⋮/proj/api   \n")
	b.WriteString("sam build --use-container⋮/proj/api   \n")
	got := rank(records(t, b.String()), "sam", "/proj/api")
	if got[0].cmd != "sam build --use-container" {
		t.Errorf("want local command first, got %q", got[0].cmd)
	}
	if got[0].hereCount != 2 {
		t.Errorf("hereCount = %d, want 2", got[0].hereCount)
	}
}

func TestFrequencyWithinSameDir(t *testing.T) {
	h := "git pull⋮/r   \ngit push⋮/r   \ngit push⋮/r   \ngit push⋮/r   \ngit pull⋮/r   \n"
	got := rank(records(t, h), "git p", "/r")
	if got[0].cmd != "git push" {
		t.Errorf("want most frequent first, got %q", got[0].cmd)
	}
}

func TestExactPrefixExcluded(t *testing.T) {
	got := rank(records(t, "ls⋮/r   \nls -la⋮/r   \n"), "ls", "/r")
	if len(got) != 1 || got[0].cmd != "ls -la" {
		t.Errorf("got %#v", got)
	}
}
