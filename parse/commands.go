package parse

import "github.com/huderlem/poryscript-pls/lsp"

// Command represents a gen 3 Pokémon decomp scripting command
type Command struct {
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
func ParseCommands(content string) ([]Command, error) {
	// TODO
	// parse commands
	// parse assembly constants
	// parse movement constants
	// hardcoded Poryscript keywords
	return []Command{}, nil
}
