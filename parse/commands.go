package parse

import (
	"regexp"

	"github.com/huderlem/poryscript-pls/lsp"
)

// Command represents a gen 3 Pokémon decomp scripting command
type Command struct {
	Name          string
	Documentation string
	Detail        string
	Kind          lsp.CompletionItemKind
	InsertText    string
	Parameters    []CommandParam
}

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
)

// ParseCommands parses the various types of commands from the given
// file content.
func ParseCommands(content string) []Command {
	// TODO
	// parse commands
	// parse assembly constants
	// parse movement constants
	// hardcoded Poryscript keywords
	return []Command{}
}

// Parses the script macro commands from the given file content.
func parseMacroCommands(content string) []Command {
	commands := []Command{}
	re, _ := regexp.Compile(`\.macro[ \t]+(\w+)[ \t]*([ \t,\w:=]*)`)
	for _, match := range re.FindAllStringSubmatchIndex(content, -1) {
		command := Command{
			Name:          content[match[2]:match[3]],
			Kind:          lsp.CIKFunction,
			Documentation: "",
			Parameters:    parseMacroParameters(content[match[4]:match[5]]),
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
