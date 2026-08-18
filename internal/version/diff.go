package version

import (
	"bytes"
	"fmt"
	"strings"
	"unicode/utf8"
)

const maxDiffLines = 400

type Op string

const (
	OpEq  Op = "eq"
	OpAdd Op = "add"
	OpDel Op = "del"
)

type LineChange struct {
	Op   Op     `json:"op"`
	OldN int    `json:"old_n,omitempty"`
	NewN int    `json:"new_n,omitempty"`
	Text string `json:"text"`
}

type DiffResult struct {
	Equal     bool         `json:"equal"`
	OldBytes  int          `json:"old_bytes"`
	NewBytes  int          `json:"new_bytes"`
	OldLines  int          `json:"old_lines"`
	NewLines  int          `json:"new_lines"`
	Binary    bool         `json:"binary"`
	Truncated bool         `json:"truncated,omitempty"`
	Changes   []LineChange `json:"changes,omitempty"`
}

func looksBinary(b []byte) bool {
	if bytes.IndexByte(b, 0) >= 0 {
		return true
	}
	if !utf8.Valid(b) {
		return true
	}
	return false
}

func splitLines(b []byte) []string {
	if len(b) == 0 {
		return nil
	}
	s := string(b)
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\r", "\n")
	return strings.Split(s, "\n")
}

// Diff 对两份 payload 做行级 LCS 差分；过大或二进制则只给摘要。
func Diff(oldB, newB []byte) DiffResult {
	res := DiffResult{
		OldBytes: len(oldB),
		NewBytes: len(newB),
		Equal:    bytes.Equal(oldB, newB),
	}
	if res.Equal {
		return res
	}
	if looksBinary(oldB) || looksBinary(newB) {
		res.Binary = true
		return res
	}
	a := splitLines(oldB)
	b := splitLines(newB)
	res.OldLines = len(a)
	res.NewLines = len(b)
	if len(a) > maxDiffLines || len(b) > maxDiffLines {
		res.Truncated = true
		return res
	}
	res.Changes = lcsChanges(a, b)
	return res
}

func lcsChanges(a, b []string) []LineChange {
	n, m := len(a), len(b)
	dp := make([][]int, n+1)
	for i := range dp {
		dp[i] = make([]int, m+1)
	}
	for i := n - 1; i >= 0; i-- {
		for j := m - 1; j >= 0; j-- {
			if a[i] == b[j] {
				dp[i][j] = dp[i+1][j+1] + 1
			} else if dp[i+1][j] >= dp[i][j+1] {
				dp[i][j] = dp[i+1][j]
			} else {
				dp[i][j] = dp[i][j+1]
			}
		}
	}
	var out []LineChange
	i, j := 0, 0
	for i < n && j < m {
		if a[i] == b[j] {
			out = append(out, LineChange{Op: OpEq, OldN: i + 1, NewN: j + 1, Text: a[i]})
			i++
			j++
			continue
		}
		if dp[i+1][j] >= dp[i][j+1] {
			out = append(out, LineChange{Op: OpDel, OldN: i + 1, Text: a[i]})
			i++
		} else {
			out = append(out, LineChange{Op: OpAdd, NewN: j + 1, Text: b[j]})
			j++
		}
	}
	for i < n {
		out = append(out, LineChange{Op: OpDel, OldN: i + 1, Text: a[i]})
		i++
	}
	for j < m {
		out = append(out, LineChange{Op: OpAdd, NewN: j + 1, Text: b[j]})
		j++
	}
	return out
}

func Unified(d DiffResult) string {
	if d.Equal {
		return ""
	}
	if d.Binary {
		return fmt.Sprintf("binary payload changed (%d -> %d bytes)\n", d.OldBytes, d.NewBytes)
	}
	var b strings.Builder
	for _, c := range d.Changes {
		switch c.Op {
		case OpAdd:
			b.WriteString("+")
			b.WriteString(c.Text)
			b.WriteByte('\n')
		case OpDel:
			b.WriteString("-")
			b.WriteString(c.Text)
			b.WriteByte('\n')
		}
	}
	if d.Truncated {
		b.WriteString("(truncated)\n")
	}
	return b.String()
}
