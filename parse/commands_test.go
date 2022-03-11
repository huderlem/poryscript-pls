package parse

import (
	"reflect"
	"strings"
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

.macro noop
.endm
`
	expected := []Command{
		{
			Name:          "msgbox",
			Kind:          lsp.CIKFunction,
			Documentation: "Buffers the given text and calls the relevant standard message script (see gStdScripts).",
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
			Documentation: "Gives 'amount' of the specified 'item' to the player and prints a message with fanfare. If the player doesn't have space for all the items then as many are added as possible, the",
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
		{
			Name:          "noop",
			Kind:          lsp.CIKFunction,
			Documentation: "",
			Parameters:    []CommandParam{},
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

func TestParseCommandDocumentation(t *testing.T) {
	input := `@ Single line  
	.macro macrofoo
	
  @ 	This has   
	@2 lines 
	.macro macrobar

	.macro macrobaz

	@ This line is not connected   

	@ This line is connected 
@
	@ to the macro
	.macro macrotest
	`
	tests := []struct {
		index    int
		expected string
	}{
		{index: -1, expected: ""},
		{index: 0, expected: ""},
		{index: strings.Index(input, "macrofoo"), expected: "Single line"},
		{index: strings.Index(input, "macrobar"), expected: "This has 2 lines"},
		{index: strings.Index(input, "macrobaz"), expected: ""},
		{index: strings.Index(input, "macrotest"), expected: "This line is connected to the macro"},
	}

	for i, tt := range tests {
		result := parseCommandDocumentation(input, tt.index)
		if result != tt.expected {
			t.Errorf("test[%d]: incorrect result. Expected '%s', Got '%s'", i, tt.expected, result)
		}
	}
}

func TestReadLine(t *testing.T) {
	input := "\tThe quick \tbrown\r\n  fox   \n\njumped over the fence."
	tests := []struct {
		index    int
		expected string
	}{
		{index: -1, expected: ""},
		{index: 0, expected: "\tThe quick \tbrown"},
		{index: 5, expected: "quick \tbrown"},
		{index: 17, expected: ""},
		{index: 18, expected: ""},
		{index: 19, expected: "  fox   "},
		{index: 29, expected: "jumped over the fence."},
		{index: 50, expected: "."},
		{index: len(input), expected: ""},
	}

	for i, tt := range tests {
		result := readLine(input, tt.index)
		if result != tt.expected {
			t.Errorf("test[%d]: incorrect resulting line. Expected '%s', Got '%s'", i, tt.expected, result)
		}
	}
}

func TestRewindToPreviousLineStart(t *testing.T) {
	input := "\tThe quick \tbrown\r\n  fox   \n\njumped over the fence."
	tests := []struct {
		index    int
		expected int
	}{
		{index: 0, expected: 0},
		{index: 16, expected: 0},
		{index: 18, expected: 0},
		{index: 19, expected: 0},
		{index: 26, expected: 0},
		{index: 27, expected: 0},
		{index: 28, expected: 19},
		{index: 50, expected: 28},
	}

	for i, tt := range tests {
		index := tt.index
		rewindToPreviousLineStart(input, &index)
		if index != tt.expected {
			t.Errorf("test[%d]: incorrect resulting index. Expected %d, Got %d", i, tt.expected, index)
		}
	}
}

func TestRewindToLineStart(t *testing.T) {
	input := "\tThe quick \tbrown\r\n  fox   \n\njumped over the fence."
	tests := []struct {
		index    int
		expected int
	}{
		{index: 0, expected: 0},
		{index: 16, expected: 0},
		{index: 18, expected: 0},
		{index: 19, expected: 19},
		{index: 27, expected: 19},
		{index: 28, expected: 28},
		{index: 29, expected: 29},
		{index: 50, expected: 29},
	}

	for i, tt := range tests {
		index := tt.index
		rewindToLineStart(input, &index)
		if index != tt.expected {
			t.Errorf("test[%d]: incorrect resulting index. Expected %d, Got %d", i, tt.expected, index)
		}
	}
}

func TestSkipWhitespace(t *testing.T) {
	input := "\tThe quick \tbrown\r\n  fox   "
	tests := []struct {
		index    int
		expected int
	}{
		{index: 0, expected: 1},
		{index: 1, expected: 1},
		{index: 3, expected: 3},
		{index: 4, expected: 5},
		{index: 9, expected: 9},
		{index: 10, expected: 12},
		{index: 17, expected: 17},
		{index: 24, expected: len(input)},
	}

	for i, tt := range tests {
		index := tt.index
		skipWhitespace(input, &index)
		if index != tt.expected {
			t.Errorf("test[%d]: incorrect resulting index. Expected %d, Got %d", i, tt.expected, index)
		}
	}
}

func TestParseAssemblyConstants(t *testing.T) {
	input := `
	.macro case condition:req, dest:req
	compare VAR_0x8000, \condition
	goto_if_eq \dest
	.endm

	@ Message box types
	MSGBOX_NPC = 2
	NO_MUSIC = FALSE
MSGBOX_DEFAULT = 4
  		MSGBOX_YESNO = 5

	YES = 1
NO  = 0`
	expected := []Command{
		{Name: "MSGBOX_NPC", Kind: lsp.CIKConstant, Detail: "2"},
		{Name: "NO_MUSIC", Kind: lsp.CIKConstant, Detail: "FALSE"},
		{Name: "MSGBOX_DEFAULT", Kind: lsp.CIKConstant, Detail: "4"},
		{Name: "MSGBOX_YESNO", Kind: lsp.CIKConstant, Detail: "5"},
		{Name: "YES", Kind: lsp.CIKConstant, Detail: "1"},
		{Name: "NO", Kind: lsp.CIKConstant, Detail: "0"},
	}
	results := parseAssemblyConstants(input)
	if len(expected) != len(results) {
		t.Fatalf("Wrong number of parsed assembler constants. Expected=%d, Got=%d", len(expected), len(results))
	}
	for i, result := range results {
		if !reflect.DeepEqual(result, expected[i]) {
			t.Fatalf("Test Case %d: parsed assembler constants is wrong.\nExpected:\n%v\n\nGot:\n%v", i, expected[i], result)
		}
	}
}

func TestParseMovementConstants(t *testing.T) {
	input := `
	.macro create_movement_action name:req, value:req
	.macro \name
	.byte \value
	.endm
	.endm

	create_movement_action face_down, MOVEMENT_ACTION_FACE_DOWN
create_movement_action face_up, MOVEMENT_ACTION_FACE_UP
  	create_movement_action face_left, MOVEMENT_ACTION_FACE_LEFT`
	expected := []Command{
		{Name: "face_down", Kind: lsp.CIKConstant, Detail: "MOVEMENT_ACTION_FACE_DOWN"},
		{Name: "face_up", Kind: lsp.CIKConstant, Detail: "MOVEMENT_ACTION_FACE_UP"},
		{Name: "face_left", Kind: lsp.CIKConstant, Detail: "MOVEMENT_ACTION_FACE_LEFT"},
	}
	results := parseMovementConstants(input)
	if len(expected) != len(results) {
		t.Fatalf("Wrong number of parsed movement constants. Expected=%d, Got=%d", len(expected), len(results))
	}
	for i, result := range results {
		if !reflect.DeepEqual(result, expected[i]) {
			t.Fatalf("Test Case %d: parsed movement constants is wrong.\nExpected:\n%v\n\nGot:\n%v", i, expected[i], result)
		}
	}
}
