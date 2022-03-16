package parse

import (
	"bufio"
	"regexp"
	"strings"

	"github.com/huderlem/poryscript-pls/lsp"
)

// ConstantSymbol represents a Poryscript constant.
type ConstantSymbol struct {
	Name     string
	Position lsp.Position
	Uri      string
}

// Returns the lsp.CompletionItem representation of a ConstantSymbol.
func (c ConstantSymbol) ToCompletionItem() lsp.CompletionItem {
	return lsp.CompletionItem{
		Label: c.Name,
		Kind:  lsp.CIKConstant,
	}
}

// Returns the lsp.Location representation of a ConstantSymbol.
func (c ConstantSymbol) ToLocation() lsp.Location {
	return lsp.Location{
		URI: lsp.DocumentURI(c.Uri),
		Range: lsp.Range{
			Start: c.Position,
			End: lsp.Position{
				Line:      c.Position.Line,
				Character: c.Position.Character + len(c.Name),
			},
		},
	}
}

// Parses the Poryscript constants from the given file content.
func ParseConstants(content string, uri string) []ConstantSymbol {
	if len(content) == 0 {
		return []ConstantSymbol{}
	}
	constants := []ConstantSymbol{}
	re, _ := regexp.Compile(`\bconst\s+(\w+)\s*=`)
	scanner := bufio.NewScanner(strings.NewReader(content))
	lineNumber := 0
	for scanner.Scan() {
		line := stripComment(scanner.Text())
		for _, match := range re.FindAllStringSubmatchIndex(line, -1) {
			nameStart, nameEnd := match[2], match[3]
			command := ConstantSymbol{
				Name: line[nameStart:nameEnd],
				Position: lsp.Position{
					Line:      lineNumber,
					Character: match[2],
				},
				Uri: uri,
			}
			constants = append(constants, command)
		}
		lineNumber++
	}
	return constants
}

// Strips Poryscript comments from the given string.
func stripComment(line string) string {
	for i := 0; i < len(line); i++ {
		if line[i] == '#' {
			return line[:i]
		}
		if i < len(line)-1 {
			if line[i] == '/' && line[i+1] == '/' {
				return line[:i]
			}
		}
	}
	return line
}
