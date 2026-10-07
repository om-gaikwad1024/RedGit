package diff

import (
	"bytes"
	"fmt"
)

type Hunk struct {
	OldStart int
	OldCount int
	NewStart int
	NewCount int
	Lines    []DiffLine
}

func ComputeUnifiedDiff(original, modified []byte, oldPath, newPath string) string {
	origLines := splitLines(original)
	modLines := splitLines(modified)
	
	diff := ComputeDiff(original, modified)
	hunks := groupIntoHunks(diff.Lines, origLines, modLines)
	
	var buf bytes.Buffer
	
	
	fmt.Fprintf(&buf, "--- a/%s\n", oldPath)
	fmt.Fprintf(&buf, "+++ b/%s\n", newPath)
	
	
	for _, hunk := range hunks {
		fmt.Fprintf(&buf, "@@ -%d,%d +%d,%d @@\n", 
			hunk.OldStart, hunk.OldCount, hunk.NewStart, hunk.NewCount)
		
		for _, line := range hunk.Lines {
			switch line.Type {
			case "add":
				fmt.Fprintf(&buf, "+%s\n", line.Content)
			case "remove":
				fmt.Fprintf(&buf, "-%s\n", line.Content)
			case "context":
				fmt.Fprintf(&buf, " %s\n", line.Content)
			}
		}
	}
	
	return buf.String()
}

func groupIntoHunks(lines []DiffLine, origLines, modLines []string) []Hunk {
	var hunks []Hunk
	var currentHunk *Hunk
	
	contextLines := 3
	oldLine := 1
	newLine := 1
	
	for i, line := range lines {
		if currentHunk == nil {
			// Start new hunk
			currentHunk = &Hunk{
				OldStart: oldLine,
				NewStart: newLine,
				Lines:    []DiffLine{},
			}
		}
		
		currentHunk.Lines = append(currentHunk.Lines, line)
		
		switch line.Type {
		case "remove":
			currentHunk.OldCount++
			oldLine++
		case "add":
			currentHunk.NewCount++
			newLine++
		case "context":
			currentHunk.OldCount++
			currentHunk.NewCount++
			oldLine++
			newLine++
		}
		
		
		if i+1 < len(lines) {
			nextChanges := countChangesAhead(lines[i+1:], contextLines*2)
			if nextChanges == 0 && len(currentHunk.Lines) > 0 {
				hunks = append(hunks, *currentHunk)
				currentHunk = nil
			}
		}
	}
	
	
	if currentHunk != nil && len(currentHunk.Lines) > 0 {
		hunks = append(hunks, *currentHunk)
	}
	
	return hunks
}

func countChangesAhead(lines []DiffLine, maxContext int) int {
	changes := 0
	contextCount := 0
	
	for _, line := range lines {
		if line.Type == "context" {
			contextCount++
			if contextCount > maxContext {
				break
			}
		} else {
			changes++
			contextCount = 0
		}
	}
	
	return changes
}

func splitLines(content []byte) []string {
	if len(content) == 0 {
		return []string{}
	}
	
	var lines []string
	start := 0
	
	for i := 0; i < len(content); i++ {
		if content[i] == '\n' {
			lines = append(lines, string(content[start:i]))
			start = i + 1
		}
	}
	
	
	if start < len(content) {
		lines = append(lines, string(content[start:]))
	}
	
	return lines
}