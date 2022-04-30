package parse

import (
	"bufio"
	"fmt"
	"regexp"
	"strings"

	"github.com/huderlem/poryscript-pls/lsp"
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

type CommandCallParts struct {
	Command    string
	OpenParen  lsp.Position
	CloseParen lsp.Position
	Commas     []lsp.Position
}

func GetCommandCallParts(content string, line int, column int) (CommandCallParts, error) {
	if line < 0 || column < 0 {
		return CommandCallParts{}, fmt.Errorf("line and column must be >= 0. line=%d, column=%d", line, column)
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
		return CommandCallParts{}, fmt.Errorf("line %d doesn't exist in the content", line)
	}
	l += "\n"

	var openParen *lsp.Position
	var i int
	for i = column; i >= 0; i-- {
		if l[i] == '(' {
			openParen = &lsp.Position{Line: line, Character: i}
			break
		}
	}
	if openParen == nil {
		return CommandCallParts{}, fmt.Errorf("line %d doesn't have an open parenthesis", line)
	}
	i--
	cmdEnd := i
	for i >= 0 && wordRe.MatchString(string(l[i])) {
		i--
	}
	if i == cmdEnd {
		return CommandCallParts{}, fmt.Errorf("line %d had no command before the parenthesis", line)
	}
	commandName := l[i+1 : openParen.Character]

	var closeParen *lsp.Position
	commas := []lsp.Position{}
	for i = openParen.Character; i < len(l); i++ {
		if l[i] == ')' {
			closeParen = &lsp.Position{Line: line, Character: i}
			break
		} else if l[i] == ',' {
			commas = append(commas, lsp.Position{Line: line, Character: i})
		}
	}
	if closeParen == nil {
		return CommandCallParts{}, fmt.Errorf("line %d doesn't have a closing parenthesis", line)
	}

	return CommandCallParts{
		Command:    commandName,
		OpenParen:  *openParen,
		CloseParen: *closeParen,
		Commas:     commas,
	}, nil
}
