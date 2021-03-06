package parse

import (
	"bufio"
	"regexp"
	"strings"

	"github.com/huderlem/poryscript-pls/lsp"
)

// Symbol represents a Poryscript symbol.
type Symbol struct {
	Name     string
	Position lsp.Position
	Uri      string
	Kind     SymbolKind
}

// SymbolKind is the type of Poryscript symbol.
type SymbolKind int

const (
	_ SymbolKind = iota
	SymbolKindScript
	SymbolKindMapScripts
	SymbolKindMovementScript
	SymbolKindMart
	SymbolKindText
	SymbolKindLabel
)

// Gets the detail text for a SymbolKind.
func (k SymbolKind) getDetail() string {
	switch k {
	case SymbolKindScript:
		return "Script"
	case SymbolKindMapScripts:
		return "Map Scripts"
	case SymbolKindMovementScript:
		return "Movement Script"
	case SymbolKindMart:
		return "Mart Items"
	case SymbolKindText:
		return "Text"
	case SymbolKindLabel:
		return "Label"
	default:
		return ""
	}
}

// Gets the CompletionItemKind for a SymbolKind.
func (k SymbolKind) getCompletionItemKind() lsp.CompletionItemKind {
	switch k {
	case SymbolKindScript, SymbolKindMapScripts, SymbolKindLabel:
		return lsp.CIKFunction
	case SymbolKindMovementScript, SymbolKindMart, SymbolKindText:
		return lsp.CIKField
	default:
		return lsp.CIKFunction
	}
}

// Returns the lsp.CompletionItem representation of a Symbol.
func (s Symbol) ToCompletionItem() lsp.CompletionItem {
	return lsp.CompletionItem{
		Label:  s.Name,
		Kind:   s.Kind.getCompletionItemKind(),
		Detail: s.Kind.getDetail(),
	}
}

// Returns the lsp.Location representation of a Symbol.
func (s Symbol) ToLocation() lsp.Location {
	return lsp.Location{
		URI: lsp.DocumentURI(s.Uri),
		Range: lsp.Range{
			Start: s.Position,
			End: lsp.Position{
				Line:      s.Position.Line,
				Character: s.Position.Character + len(s.Name),
			},
		},
	}
}

var symbolRegexes = []struct {
	re   *regexp.Regexp
	kind SymbolKind
}{
	{
		re:   regexp.MustCompile(`\bscript\s+(\w+)\s*\{`),
		kind: SymbolKindScript,
	},
	{
		re:   regexp.MustCompile(`\bmovement\s+(\w+)\s*\{`),
		kind: SymbolKindMovementScript,
	},
	{
		re:   regexp.MustCompile(`\bmart\s+(\w+)\s*\{`),
		kind: SymbolKindMart,
	},
	{
		re:   regexp.MustCompile(`\bmapscripts\s+(\w+)\s*\{`),
		kind: SymbolKindMapScripts,
	},
	{
		re:   regexp.MustCompile(`\btext\s+(\w+)\s*\{`),
		kind: SymbolKindText,
	},
	{
		re:   regexp.MustCompile(`^\s*(\w+)\s*(\(\s*(local|global)\s*\))?\s*:`),
		kind: SymbolKindLabel,
	},
}

// Parse the Poryscript symbols from the given file content.
func ParseSymbols(content string, fileUri string) []Symbol {
	if len(content) == 0 {
		return []Symbol{}
	}
	scanner := bufio.NewScanner(strings.NewReader(content))
	symbols := []Symbol{}
	lineNumber := 0
	for scanner.Scan() {
		line := stripComment(scanner.Text())
		for _, r := range symbolRegexes {
			for _, match := range r.re.FindAllStringSubmatchIndex(line, -1) {
				nameStart, nameEnd := match[2], match[3]
				symbols = append(symbols, Symbol{
					Name:     line[nameStart:nameEnd],
					Position: lsp.Position{Character: nameStart, Line: lineNumber},
					Uri:      fileUri,
					Kind:     r.kind,
				})
			}
		}
		lineNumber++
	}
	return symbols
}
