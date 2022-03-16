package parse

import (
	"reflect"
	"testing"

	"github.com/huderlem/poryscript-pls/lsp"
)

func TestMiscTokenToCompletionItem(t *testing.T) {
	tests := []struct {
		input    MiscToken
		expected lsp.CompletionItem
	}{
		{
			input:    MiscToken{},
			expected: lsp.CompletionItem{Kind: lsp.CIKValue},
		},
		{
			input:    MiscToken{Name: "Test1"},
			expected: lsp.CompletionItem{Label: "Test1", Kind: lsp.CIKValue},
		},
		{
			input:    MiscToken{Name: "Test2", Type: "special"},
			expected: lsp.CompletionItem{Label: "Test2", Kind: lsp.CIKFunction, Detail: "Special Function"},
		},
		{
			input:    MiscToken{Name: "Test3", Type: "define", Value: "Value3"},
			expected: lsp.CompletionItem{Label: "Test3", Kind: lsp.CIKConstant, Detail: "Value3"},
		},
		{
			input:    MiscToken{Name: "Test3", Type: "other"},
			expected: lsp.CompletionItem{Label: "Test3", Kind: lsp.CIKValue},
		},
	}
	for i, tt := range tests {
		result := tt.input.ToCompletionItem()
		if !reflect.DeepEqual(result, tt.expected) {
			t.Errorf("Test Case %d:\nExpected:\n%v\n\nGot:\n%v", i, tt.expected, result)
		}
	}
}

func TestMiscTokenToLocation(t *testing.T) {
	tests := []struct {
		input    MiscToken
		expected lsp.Location
	}{
		{
			input:    MiscToken{},
			expected: lsp.Location{},
		},
		{
			input:    MiscToken{Uri: "testfile.pory"},
			expected: lsp.Location{URI: "testfile.pory"},
		},
		{
			input:    MiscToken{Name: "foo", Position: lsp.Position{Line: 2, Character: 7}, Uri: "testfile.pory"},
			expected: lsp.Location{Range: lsp.Range{Start: lsp.Position{Line: 2, Character: 7}, End: lsp.Position{Line: 2, Character: 10}}, URI: "testfile.pory"},
		},
	}
	for i, tt := range tests {
		result := tt.input.ToLocation()
		if !reflect.DeepEqual(result, tt.expected) {
			t.Errorf("Test Case %d:\nExpected:\n%v\n\nGot:\n%v", i, tt.expected, result)
		}
	}
}

func TestParseMiscTokens(t *testing.T) {
	tests := []struct {
		input      string
		expression string
		type_      string
		file       string
		expected   []MiscToken
	}{
		{
			input:    ``,
			expected: []MiscToken{},
		},
		{
			input:      `stuff`,
			expression: `(?x)invalid regex`,
			expected:   []MiscToken{},
		},
		{
			input: `
gSpecials::
def_special HealPlayerParty
	def_special SetCableClubWarp
def_special DoCableClubWarp`,
			expression: `^\s*def_special\s+(\w+)`,
			type_:      "special",
			file:       "file://data/specials.inc",
			expected: []MiscToken{
				{Name: "HealPlayerParty", Position: lsp.Position{Character: 12, Line: 2}, Type: "special", Uri: "file://data/specials.inc"},
				{Name: "SetCableClubWarp", Position: lsp.Position{Character: 13, Line: 3}, Type: "special", Uri: "file://data/specials.inc"},
				{Name: "DoCableClubWarp", Position: lsp.Position{Character: 12, Line: 4}, Type: "special", Uri: "file://data/specials.inc"},
			},
		},
		{
			input: `
#define FLAG_TEST    0x4F // Unused Flag
#define FLAG_TEST_2                      0x51
#define VAR_NOPE 50
  #define FLAG_TEST_3                (FLAG_HIDDEN_ITEMS_START + 0x39)`,
			expression: `^\s*#define\s+(FLAG_\w+)\s+(.+)`,
			type_:      "define",
			file:       "file://include/constants/flags.h",
			expected: []MiscToken{
				{Name: "FLAG_TEST", Position: lsp.Position{Character: 8, Line: 1}, Type: "define", Uri: "file://include/constants/flags.h", Value: "0x4F // Unused Flag"},
				{Name: "FLAG_TEST_2", Position: lsp.Position{Character: 8, Line: 2}, Type: "define", Uri: "file://include/constants/flags.h", Value: "0x51"},
				{Name: "FLAG_TEST_3", Position: lsp.Position{Character: 10, Line: 4}, Type: "define", Uri: "file://include/constants/flags.h", Value: "(FLAG_HIDDEN_ITEMS_START + 0x39)"},
			},
		},
	}

	for i, tt := range tests {
		result := ParseMiscTokens(tt.input, tt.expression, tt.type_, tt.file)
		if !reflect.DeepEqual(result, tt.expected) {
			t.Errorf("Test Case %d: parsed misc tokens are wrong.\nExpected:\n%v\n\nGot:\n%v", i, tt.expected, result)
		}
	}
}
