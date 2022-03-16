package parse

import (
	"bufio"
	"regexp"
	"strings"

	"github.com/huderlem/poryscript-pls/lsp"
)

// MiscToken represents a miscellaneous token included from arbitrary
// files. It might be an assembler define or constant, for example.
type MiscToken struct {
	Name     string
	Position lsp.Position
	Type     string
	Uri      string
	Value    string
}

// Gets the CompletionItemKind for the MiscToken's type.
func (t MiscToken) getCompletionItemKind() lsp.CompletionItemKind {
	switch t.Type {
	case "special":
		return lsp.CIKFunction
	case "define":
		return lsp.CIKConstant
	default:
		return lsp.CIKValue
	}
}

// Gets the CompletionItemKind for the MiscToken's type.
func (t MiscToken) getDetail() string {
	switch t.Type {
	case "special":
		return "Special Function"
	case "define":
		return t.Value
	default:
		return ""
	}
}

func (t MiscToken) ToCompletionItem() lsp.CompletionItem {
	return lsp.CompletionItem{
		Label:  t.Name,
		Kind:   t.getCompletionItemKind(),
		Detail: t.getDetail(),
	}
}

// Returns the lsp.Location representation of a MiscToken.
func (t MiscToken) ToLocation() lsp.Location {
	return lsp.Location{
		URI: lsp.DocumentURI(t.Uri),
		Range: lsp.Range{
			Start: t.Position,
			End: lsp.Position{
				Line:      t.Position.Line,
				Character: t.Position.Character + len(t.Name),
			},
		},
	}
}

// Parses the miscellaneous tokens from the given file content and regex.
func ParseMiscTokens(content string, expression string, tokenType string, fileUri string) []MiscToken {
	if len(content) == 0 {
		return []MiscToken{}
	}
	re, err := regexp.Compile(expression)
	if err != nil {
		return []MiscToken{}
	}
	scanner := bufio.NewScanner(strings.NewReader(content))
	tokens := []MiscToken{}
	lineNumber := 0
	for scanner.Scan() {
		line := scanner.Text()
		for _, match := range re.FindAllStringSubmatchIndex(line, -1) {
			start, end := match[2], match[3]
			token := MiscToken{
				Name:     line[start:end],
				Position: lsp.Position{Character: start, Line: lineNumber},
				Type:     tokenType,
				Uri:      fileUri,
			}
			if tokenType == "define" && len(match) > 5 {
				token.Value = line[match[4]:match[5]]
			}
			tokens = append(tokens, token)
		}
		lineNumber++
	}
	return tokens
}
