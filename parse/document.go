package parse

import (
	"bufio"
	"regexp"
	"strings"
)

// Gets the full token at the given position in the content.
func GetTokenAt(content string, line int, column int) string {
	if line < 0 || column < 0 {
		return ""
	}
	// Scan to the targeted line.
	curLine := 0
	scanner := bufio.NewScanner(strings.NewReader(content))
	l := ""
	for scanner.Scan() {
		l = stripComment(scanner.Text())
		if curLine == line {
			break
		}
		curLine++
	}
	if curLine != line || column > len(l) {
		return ""
	}
	start, end := getWordBounds(l, column)
	if start >= end {
		return ""
	}
	return l[start:end]
}

func getWordBounds(line string, column int) (int, int) {
	return getWordStart(line, column), getWordEnd(line, column)
}

var wordRe = regexp.MustCompile(`\w`)

func getWordStart(line string, column int) int {
	if column == len(line) || (column < len(line) && column >= 0 && !wordRe.MatchString(line[column:column+1])) {
		column--
	}
	for {
		if column < 0 {
			return 0
		}
		if column >= len(line) || !wordRe.MatchString(line[column:column+1]) {
			break
		}
		column--
	}
	return column + 1
}

func getWordEnd(line string, column int) int {
	for column < len(line) && column >= 0 {
		if !wordRe.MatchString(line[column : column+1]) {
			break
		}
		column++
	}
	return column
}
