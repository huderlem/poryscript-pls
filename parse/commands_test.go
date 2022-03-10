package parse

import (
	"reflect"
	"testing"

	"github.com/huderlem/poryscript-pls/lsp"
)

func TestParseMacroCommands(t *testing.T) {
	input := `
@ Buffers the given text and calls the relevant standard message script (see gStdScripts).
.macro 	    msgbox text:req, type=MSGBOX_DEFAULT optarg
loadword 0, \text
.endm

@ Gives 'amount' of the specified 'item' to the player and prints a message with fanfare.
@ If the player doesn't have space for all the items then as many are added as possible, the
.macro giveitem amount=1, 	 item:req
setorcopyvar VAR_0x8000, \item
.endm
`
	expected := []Command{
		{
			Name:          "msgbox",
			Kind:          lsp.CIKFunction,
			Documentation: "",
			Parameters: []CommandParam{
				{
					Name: "text",
					Kind: CommandParamRequired,
				},
				{
					Name:    "type",
					Kind:    CommandParamDefault,
					Default: "MSGBOX_DEFAULT",
				},
				{
					Name: "optarg",
					Kind: CommandParamOptional,
				},
			},
		},
		{
			Name:          "giveitem",
			Kind:          lsp.CIKFunction,
			Documentation: "",
			Parameters: []CommandParam{
				{
					Name:    "amount",
					Kind:    CommandParamDefault,
					Default: "1",
				},
				{
					Name: "item",
					Kind: CommandParamRequired,
				},
			},
		},
	}
	results := parseMacroCommands(input)
	if len(expected) != len(results) {
		t.Fatalf("Wrong number of parsed macro commands. Expected=%d, Got=%d", len(expected), len(results))
	}
	for i, result := range results {
		if !reflect.DeepEqual(result, expected[i]) {
			t.Fatalf("Test Case %d: parsed macro command is wrong.\nExpected:\n%v\n\nGot:\n%v", i, expected[i], result)
		}
	}
}
