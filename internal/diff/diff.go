package diff

import (
	"bytes"
	"fmt"
	"strings"
)

type DiffLine struct {
	Type    string // "add", "remove", "context"
	Content string
	LineNum int
}

type Diff struct {
	Lines []DiffLine
}

func ComputeDiff(original, modified []byte) *Diff {
	origLines := strings.Split(string(original), "\n")
	modLines := strings.Split(string(modified), "\n")
	
	diff := &Diff{Lines: []DiffLine{}}
	
	lcs := longestCommonSubsequence(origLines, modLines)
	
	i, j := 0, 0
	for _, common := range lcs {
		for i < len(origLines) && origLines[i] != common {
			diff.Lines = append(diff.Lines, DiffLine{
				Type:    "remove",
				Content: origLines[i],
				LineNum: i + 1,
			})
			i++
		}
		
		for j < len(modLines) && modLines[j] != common {
			diff.Lines = append(diff.Lines, DiffLine{
				Type:    "add",
				Content: modLines[j],
				LineNum: j + 1,
			})
			j++
		}
		
		if i < len(origLines) {
			diff.Lines = append(diff.Lines, DiffLine{
				Type:    "context",
				Content: origLines[i],
				LineNum: i + 1,
			})
			i++
			j++
		}
	}
	
	for i < len(origLines) {
		diff.Lines = append(diff.Lines, DiffLine{
			Type:    "remove",
			Content: origLines[i],
			LineNum: i + 1,
		})
		i++
	}
	
	for j < len(modLines) {
		diff.Lines = append(diff.Lines, DiffLine{
			Type:    "add",
			Content: modLines[j],
			LineNum: j + 1,
		})
		j++
	}
	
	return diff
}

func longestCommonSubsequence(a, b []string) []string {
	m, n := len(a), len(b)
	dp := make([][]int, m+1)
	for i := range dp {
		dp[i] = make([]int, n+1)
	}
	
	for i := 1; i <= m; i++ {
		for j := 1; j <= n; j++ {
			if a[i-1] == b[j-1] {
				dp[i][j] = dp[i-1][j-1] + 1
			} else {
				dp[i][j] = max(dp[i-1][j], dp[i][j-1])
			}
		}
	}
	
	var lcs []string
	i, j := m, n
	for i > 0 && j > 0 {
		if a[i-1] == b[j-1] {
			lcs = append([]string{a[i-1]}, lcs...)
			i--
			j--
		} else if dp[i-1][j] > dp[i][j-1] {
			i--
		} else {
			j--
		}
	}
	
	return lcs
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func (d *Diff) Format() string {
	var buf bytes.Buffer
	
	for _, line := range d.Lines {
		switch line.Type {
		case "add":
			fmt.Fprintf(&buf, "+%s\n", line.Content)
		case "remove":
			fmt.Fprintf(&buf, "-%s\n", line.Content)
		case "context":
			fmt.Fprintf(&buf, " %s\n", line.Content)
		}
	}
	
	return buf.String()
}