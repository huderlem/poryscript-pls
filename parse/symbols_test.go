package parse

import (
	"reflect"
	"testing"

	"github.com/huderlem/poryscript-pls/lsp"
)

func TestParseSymbols(t *testing.T) {
	input := `
	script MyScript { // script nope { 
		lockall
	}
text MyText { "foo" } text MyText2 {"bar"}
  	mapscripts MyMapScripts {} movement MyMovement {}
	mart MyMart {}`
	expected := []Symbol{
		{Name: "MyScript", Position: lsp.Position{Line: 1, Character: 8}, Uri: "test.pory", Kind: SymbolKindScript},
		{Name: "MyText", Position: lsp.Position{Line: 4, Character: 5}, Uri: "test.pory", Kind: SymbolKindText},
		{Name: "MyText2", Position: lsp.Position{Line: 4, Character: 27}, Uri: "test.pory", Kind: SymbolKindText},
		{Name: "MyMovement", Position: lsp.Position{Line: 5, Character: 39}, Uri: "test.pory", Kind: SymbolKindMovementScript},
		{Name: "MyMapScripts", Position: lsp.Position{Line: 5, Character: 14}, Uri: "test.pory", Kind: SymbolKindMapScripts},
		{Name: "MyMart", Position: lsp.Position{Line: 6, Character: 6}, Uri: "test.pory", Kind: SymbolKindMart},
	}
	results := ParseSymbols(input, "test.pory")
	if len(expected) != len(results) {
		t.Fatalf("Wrong number of parsed Poryscript symbols. Expected=%d, Got=%d", len(expected), len(results))
	}
	for i, result := range results {
		if !reflect.DeepEqual(result, expected[i]) {
			t.Errorf("Test Case %d: parsed Poryscript symbol is wrong.\nExpected:\n%v\n\nGot:\n%v", i, expected[i], result)
		}
	}

	if len(ParseSymbols("", "test.pory")) != 0 {
		t.Errorf("ParseSymbols with empty string should return an empty array")
	}
}
