package parse

import (
	"reflect"
	"testing"

	"github.com/huderlem/poryscript-pls/lsp"
)

func TestConstantToCompletionItem(t *testing.T) {
	tests := []struct {
		input    ConstantSymbol
		expected lsp.CompletionItem
	}{
		{
			input:    ConstantSymbol{},
			expected: lsp.CompletionItem{Kind: lsp.CIKConstant},
		},
		{
			input:    ConstantSymbol{Name: "foo_bar"},
			expected: lsp.CompletionItem{Label: "foo_bar", Kind: lsp.CIKConstant},
		},
		{
			input:    ConstantSymbol{Name: "foo_bar", Position: lsp.Position{}},
			expected: lsp.CompletionItem{Label: "foo_bar", Kind: lsp.CIKConstant},
		},
		{
			input:    ConstantSymbol{Name: "foo_bar", Position: lsp.Position{Line: 10, Character: 15}},
			expected: lsp.CompletionItem{Label: "foo_bar", Kind: lsp.CIKConstant},
		},
	}
	for i, tt := range tests {
		result := tt.input.ToCompletionItem()
		if !reflect.DeepEqual(result, tt.expected) {
			t.Errorf("Test Case %d:\nExpected:\n%v\n\nGot:\n%v", i, tt.expected, result)
		}
	}
}

func TestConstantToLocation(t *testing.T) {
	tests := []struct {
		input    ConstantSymbol
		expected lsp.Location
	}{
		{
			input:    ConstantSymbol{},
			expected: lsp.Location{},
		},
		{
			input:    ConstantSymbol{Uri: "testfile.pory"},
			expected: lsp.Location{URI: "testfile.pory"},
		},
		{
			input:    ConstantSymbol{Name: "foo", Position: lsp.Position{Line: 2, Character: 7}, Uri: "testfile.pory"},
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

func TestParseConstants(t *testing.T) {
	input := `
const FOO = 54 + 3
 const Nope asdf =  10
  	const BAR = 22 const BAZ = FOO
	# const IGNORE_ME = foo`
	expected := []ConstantSymbol{
		{Name: "FOO", Position: lsp.Position{Line: 1, Character: 6}, Uri: "testfile.pory"},
		{Name: "BAR", Position: lsp.Position{Line: 3, Character: 9}, Uri: "testfile.pory"},
		{Name: "BAZ", Position: lsp.Position{Line: 3, Character: 24}, Uri: "testfile.pory"},
	}
	results := ParseConstants(input, "testfile.pory")
	if len(expected) != len(results) {
		t.Fatalf("Wrong number of parsed Poryscript constants. Expected=%d, Got=%d", len(expected), len(results))
	}
	for i, result := range results {
		if !reflect.DeepEqual(result, expected[i]) {
			t.Errorf("Test Case %d: parsed Poryscript constants is wrong.\nExpected:\n%v\n\nGot:\n%v", i, expected[i], result)
		}
	}

	if len(ParseConstants("", "testfile.pory")) != 0 {
		t.Errorf("ParseConstants with empty string should return an empty array")
	}
}

func TestStripComments(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			input:    "",
			expected: "",
		},
		{
			input:    "start of line # comment starts",
			expected: "start of line ",
		},
		{
			input:    "start of line // comment starts",
			expected: "start of line ",
		},
		{
			input:    "# whole line is comment",
			expected: "",
		},
		{
			input:    "// whole line is comment",
			expected: "",
		},
		{
			input:    "/ not a comment",
			expected: "/ not a comment",
		},
		{
			input:    "last #",
			expected: "last ",
		},
		{
			input:    "last //",
			expected: "last ",
		},
	}
	for i, tt := range tests {
		result := stripComment(tt.input)
		if result != tt.expected {
			t.Errorf("Test Case %d: Expected: %s, Got: %s", i, tt.expected, result)
		}
	}
}
