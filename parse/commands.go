package parse

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/huderlem/poryscript-pls/lsp"
)

// Command represents a gen 3 Pokémon decomp scripting command
type Command struct {
	Name           string
	Documentation  string
	Detail         string
	Kind           CommandKind
	CompletionKind lsp.CompletionItemKind
	InsertText     string
	Parameters     []CommandParam
}

type CommandKind int

const (
	_ CommandKind = iota
	CommandScriptMacro
	CommandAssemblyConstant
	CommandMovement
	CommandPoryscriptKeyword
)

// CommandParam represents a macro parameter for a scripting command.
type CommandParam struct {
	Name    string
	Kind    CommandParamKind
	Default string
}

// CommandParamKind is the type of a scripting macro parameter.
type CommandParamKind int

const (
	_ CommandParamKind = iota
	CommandParamRequired
	CommandParamDefault
	CommandParamOptional
	CommandParamVarArg
)

// Returns the lsp.CompletionItem representation of a Command.
func (c Command) ToCompletionItem() lsp.CompletionItem {
	kind := c.CompletionKind
	if kind == 0 {
		kind = lsp.CIKKeyword
	}
	result := lsp.CompletionItem{
		Label:         c.Name,
		Kind:          kind,
		Documentation: c.Documentation,
		Detail:        c.Detail,
	}
	if len(c.InsertText) > 0 {
		result.InsertText = c.InsertText
		result.InsertTextFormat = lsp.ITFSnippet
	}
	return result
}

// Gets the parameters label for signature help.
func (c Command) GetParamsLabel() string {
	var sb strings.Builder
	sb.WriteString(c.Name)
	sb.WriteRune('(')
	labels := []string{}
	for _, p := range c.Parameters {
		labels = append(labels, p.getLabelName())
	}
	sb.WriteString(strings.Join(labels, ", "))
	sb.WriteRune(')')
	return sb.String()
}

func (c Command) GetParamInformation() []lsp.ParameterInformation {
	result := []lsp.ParameterInformation{}
	for _, p := range c.Parameters {
		result = append(result, lsp.ParameterInformation{
			Label:         p.Name,
			Documentation: p.getParamKindLabel(),
		})
	}
	return result
}

func (c Command) HasVarargParam() bool {
	numParams := len(c.Parameters)
	if numParams == 0 {
		return false
	}
	return c.Parameters[numParams-1].Kind == CommandParamVarArg
}

func (c CommandParam) getLabelName() string {
	switch c.Kind {
	case CommandParamRequired:
		return c.Name
	case CommandParamDefault:
		return fmt.Sprintf("[%s=%s]", c.Name, c.Default)
	case CommandParamOptional:
		return fmt.Sprintf("[%s]", c.Name)
	case CommandParamVarArg:
		return fmt.Sprintf("[%s=...]", c.Name)
	default:
		return ""
	}
}

func (c CommandParam) getParamKindLabel() lsp.MarkupContent {
	switch c.Kind {
	case CommandParamDefault:
		return lsp.MarkupContent{
			Value: fmt.Sprintf("%s=%s", c.Name, c.Default),
			Kind:  lsp.MarkupKindMarkdown,
		}
	case CommandParamOptional:
		return lsp.MarkupContent{
			Value: fmt.Sprintf("%s *Optional*", c.Name),
			Kind:  lsp.MarkupKindMarkdown,
		}
	case CommandParamRequired:
		return lsp.MarkupContent{
			Value: fmt.Sprintf("%s *Required*", c.Name),
			Kind:  lsp.MarkupKindMarkdown,
		}
	case CommandParamVarArg:
		return lsp.MarkupContent{
			Value: fmt.Sprintf("%s *Vararg*", c.Name),
			Kind:  lsp.MarkupKindMarkdown,
		}
	default:
		return lsp.MarkupContent{
			Value: "",
			Kind:  lsp.MarkupKindPlaintext,
		}
	}
}

// ParseCommands parses the various types of commands from the given
// file content.
func ParseCommands(content string) []Command {
	if len(content) == 0 {
		return []Command{}
	}
	commands := parseMacroCommands(content)
	commands = append(commands, parseAssemblyConstants(content)...)
	commands = append(commands, parseMovementConstants(content)...)
	return commands
}

// Parses the script macro commands from the given file content.
func parseMacroCommands(content string) []Command {
	commands := []Command{}
	re, _ := regexp.Compile(`\.macro[ \t]+(\w+)[ \t]*([ \t,\w:=]*)`)
	for _, match := range re.FindAllStringSubmatchIndex(content, -1) {
		nameStart, nameEnd := match[2], match[3]
		paramStart, paramEnd := match[4], match[5]
		command := Command{
			Name:           content[nameStart:nameEnd],
			Kind:           CommandScriptMacro,
			CompletionKind: lsp.CIKFunction,
			Documentation:  parseCommandDocumentation(content, nameStart),
			Parameters:     parseMacroParameters(content[paramStart:paramEnd]),
		}
		commands = append(commands, command)
	}
	return commands
}

// Parses the parameters from a script macro definiition.
func parseMacroParameters(input string) []CommandParam {
	if len(input) == 0 {
		return []CommandParam{}
	}
	params := []CommandParam{}
	re, _ := regexp.Compile(`(\w+)([:=]?)(\w+)*`)
	for _, match := range re.FindAllStringSubmatch(input, -1) {
		var defaultValue string
		kind := CommandParamOptional
		if match[2] == ":" && match[3] == "req" {
			kind = CommandParamRequired
		} else if match[2] == ":" && match[3] == "vararg" {
			kind = CommandParamVarArg
		} else if match[2] == "=" {
			kind = CommandParamDefault
			defaultValue = match[3]
		}
		param := CommandParam{
			Name:    match[1],
			Kind:    kind,
			Default: defaultValue,
		}
		params = append(params, param)
	}
	return params
}

// Parses the multiline documentation preceding the script macro command.
// The given index should be the index of the input string of the macro
// command name.
func parseCommandDocumentation(input string, index int) string {
	lines := []string{}
	for {
		prev := index
		rewindToPreviousLineStart(input, &index)
		if prev == index {
			// We've reached the start of the input string.
			break
		}
		i := index
		skipWhitespace(input, &i)
		if index >= len(input) || input[i] != '@' {
			break
		}
		// Skip past the '@' character, and read the line.
		line := strings.TrimSpace(readLine(input, i+1))
		if len(line) > 0 {
			// Prepend to the list because we're gathering the lines in reverse order.
			lines = append([]string{line}, lines...)
		}
	}
	return strings.Join(lines, " ")
}

func readLine(input string, index int) string {
	if index < 0 || index >= len(input) {
		return ""
	}
	end := index
	for end < len(input) && input[end] != '\r' && input[end] != '\n' {
		end++
	}
	return input[index:end]
}

// Moves index to the start of the previous line in the input.
// If there is no previous line, the index is moved to the start
// of the current line.
func rewindToPreviousLineStart(input string, index *int) {
	rewindToLineStart(input, index)
	if *index <= 0 {
		return
	}
	*index--
	rewindToLineStart(input, index)
}

// Moves index to the start of the current line.
// If the given index is already at the start of the line,
// then this function does nothing.
func rewindToLineStart(input string, index *int) {
	for *index > 0 && input[*index-1] != '\n' {
		*index--
	}
}

// Advances the index past any spaces or tabs in the input string.
// This does NOT advance past newline characters.
// If the end of the input string is reached, *index will result
// in the length of the string. Callers should check that the resulting
// index is still in bounds of the string.
func skipWhitespace(input string, index *int) {
	for *index < len(input) {
		cur := input[*index]
		if !(cur == ' ' || cur == '\t') {
			return
		}
		*index++
	}
}

// Parses the assembler constants from the given file content.
func parseAssemblyConstants(content string) []Command {
	commands := []Command{}
	re, _ := regexp.Compile(`(?m)^[\t ]*(\w+)[\t ]*=[\t ]*([\w\d]+)[\w\t]*$`)
	for _, match := range re.FindAllStringSubmatch(content, -1) {
		command := Command{
			Name:           match[1],
			Kind:           CommandAssemblyConstant,
			CompletionKind: lsp.CIKConstant,
			Detail:         match[2],
		}
		commands = append(commands, command)
	}
	return commands
}

// Parses the movement-related assembler constants from the given file content.
func parseMovementConstants(content string) []Command {
	commands := []Command{}
	re, _ := regexp.Compile(`(?m)^[\t ]*(?:create_movement_action)[\t ]* ([\w\d]+)(\s*,[\t ]*([\w\d]*))?$`)
	for _, match := range re.FindAllStringSubmatch(content, -1) {
		command := Command{
			Name:           match[1],
			Kind:           CommandMovement,
			CompletionKind: lsp.CIKConstant,
			Detail:         match[3],
		}
		commands = append(commands, command)
	}
	return commands
}

// Keyword commands reserved by Poryscript's language.
var KeywordCommands = []Command{
	{
		Name:           "script",
		Detail:         "Script (Poryscript)",
		Kind:           CommandPoryscriptKeyword,
		CompletionKind: lsp.CIKClass,
		InsertText:     "script ${0:MyScript} {\n    \n}",
	},
	{
		Name:           "movement",
		Detail:         "Movement (Poryscript)",
		Kind:           CommandPoryscriptKeyword,
		CompletionKind: lsp.CIKClass,
		InsertText:     "movement ${0:MyMovement} {\n    \n}",
	},
	{
		Name:           "mart",
		Detail:         "Mart (Poryscript)",
		Kind:           CommandPoryscriptKeyword,
		CompletionKind: lsp.CIKClass,
		InsertText:     "mart ${0:MyMart} {\n    \n}",
	},
	{
		Name:           "mapscripts",
		Detail:         "Mapscript Section (Poryscript)",
		Kind:           CommandPoryscriptKeyword,
		CompletionKind: lsp.CIKClass,
		InsertText:     "mapscripts ${0:MyMapscripts} {\n    \n}",
	},
	{
		Name:           "text",
		Detail:         "Text (Poryscript)",
		Kind:           CommandPoryscriptKeyword,
		CompletionKind: lsp.CIKClass,
		InsertText:     "text ${0:MyString} {\n    \n}",
	},
	{
		Name:           "raw",
		Detail:         "Raw Section (Poryscript)",
		Kind:           CommandPoryscriptKeyword,
		CompletionKind: lsp.CIKClass,
		InsertText:     "raw `\n$0\n`",
	},
	{
		Name:           "local",
		Detail:         "Local Scope (Poryscript)",
		Kind:           CommandPoryscriptKeyword,
		CompletionKind: lsp.CIKKeyword,
	},
	{
		Name:           "global",
		Detail:         "Global Scope (Poryscript)",
		Kind:           CommandPoryscriptKeyword,
		CompletionKind: lsp.CIKKeyword,
	},
	{
		Name:           "format",
		Detail:         "Format String",
		Kind:           CommandPoryscriptKeyword,
		CompletionKind: lsp.CIKFunction,
		InsertText:     "format(\"$0\")",
	},
	{
		Name:           "moves",
		Detail:         "Inline Moves",
		Kind:           CommandPoryscriptKeyword,
		CompletionKind: lsp.CIKFunction,
		InsertText:     "moves(\"$0\")",
	},
	{
		Name:           "var",
		Detail:         "Get the value of a variable",
		Kind:           CommandPoryscriptKeyword,
		CompletionKind: lsp.CIKReference,
		InsertText:     "var(${0:VAR_ID})",
	},
	{
		Name:           "flag",
		Detail:         "Get the value of a flag",
		Kind:           CommandPoryscriptKeyword,
		CompletionKind: lsp.CIKReference,
		InsertText:     "flag(${0:FLAG_ID})",
	},
	{
		Name:           "defeated",
		Detail:         "Get the status of a trainer",
		Kind:           CommandPoryscriptKeyword,
		CompletionKind: lsp.CIKReference,
		InsertText:     "defeated(${0:TRAINER_ID})",
	},
	{
		Name:       "poryswitch",
		Detail:     "Compile time switch",
		InsertText: "poryswitch(${0:SWITCH_CONDITION}) {\n    _:\n}",
		Parameters: []CommandParam{
			{
				Name: "SWITCH_CONDITION",
				Kind: CommandParamRequired,
			},
		},
	},
	{
		Name:           "if",
		Kind:           CommandPoryscriptKeyword,
		CompletionKind: lsp.CIKKeyword,
	},
	{
		Name:           "elif",
		Kind:           CommandPoryscriptKeyword,
		CompletionKind: lsp.CIKKeyword,
	},
	{
		Name:           "else",
		Kind:           CommandPoryscriptKeyword,
		CompletionKind: lsp.CIKKeyword,
	},
	{
		Name:           "while",
		Kind:           CommandPoryscriptKeyword,
		CompletionKind: lsp.CIKKeyword,
	},
	{
		Name:           "switch",
		Kind:           CommandPoryscriptKeyword,
		CompletionKind: lsp.CIKKeyword,
	},
	{
		Name:           "case",
		Kind:           CommandPoryscriptKeyword,
		CompletionKind: lsp.CIKKeyword,
	},
	{
		Name:           "break",
		Kind:           CommandPoryscriptKeyword,
		CompletionKind: lsp.CIKKeyword,
	},
	{
		Name:           "continue",
		Kind:           CommandPoryscriptKeyword,
		CompletionKind: lsp.CIKKeyword,
	},
}
