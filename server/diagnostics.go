package server

import (
	"context"
	"errors"
	"fmt"

	"github.com/huderlem/poryscript-pls/lsp"
	"github.com/huderlem/poryscript-pls/parse"
	"github.com/huderlem/poryscript/ast"
	"github.com/huderlem/poryscript/lexer"
	"github.com/huderlem/poryscript/parser"
)

// Checks the given Poryscript file content for diagnostic errors.
// Any diagnostics are immediately published to the client.
func (s *poryscriptServer) validatePoryscriptFile(ctx context.Context, fileUri string) error {
	diagnostics := lsp.PublishDiagnosticsParams{
		URI:         lsp.DocumentURI(fileUri),
		Diagnostics: []lsp.Diagnostic{},
	}
	content, err := s.getDocumentContent(ctx, fileUri)
	if err != nil {
		// TODO: log error?
		s.connection.Notify(ctx, "textDocument/publishDiagnostics", diagnostics)
		return err
	}

	p := parser.NewLintParser(lexer.New(content))
	program, err := p.ParseProgram()
	if err == nil {
		// The poryscript file is syntactically correct. Check for warnings.
		diagnostics.Diagnostics = append(diagnostics.Diagnostics, s.getPoryscriptWarnings(ctx, program, fileUri)...)
	} else {
		var parsedErr parser.ParseError
		if !errors.As(err, &parsedErr) {
			// TODO: this is an unknown error type, so we can't
			// do anything with it. Log it?
			s.connection.Notify(ctx, "textDocument/publishDiagnostics", diagnostics)
			return nil
		}

		diagnostics.Diagnostics = append(diagnostics.Diagnostics,
			lsp.Diagnostic{
				Range: lsp.Range{
					Start: lsp.Position{Line: parsedErr.LineNumberStart - 1, Character: parsedErr.CharStart},
					End:   lsp.Position{Line: parsedErr.LineNumberEnd - 1, Character: parsedErr.CharEnd},
				},
				Severity: lsp.Error,
				Source:   "Poryscript",
				Message:  parsedErr.Message,
			},
		)
	}

	s.connection.Notify(ctx, "textDocument/publishDiagnostics", diagnostics)
	return nil
}

func (s *poryscriptServer) getPoryscriptWarnings(ctx context.Context, program *ast.Program, fileUri string) []lsp.Diagnostic {
	diagnostics := []lsp.Diagnostic{}
	for _, topStatement := range program.TopLevelStatements {
		switch statement := topStatement.(type) {
		case *ast.ScriptStatement:
			diagnostics = append(diagnostics, s.getScriptWarnings(ctx, fileUri, statement)...)
		case *ast.MovementStatement:
			diagnostics = append(diagnostics, s.getMovementWarnings(ctx, fileUri, statement)...)
		}
	}
	return diagnostics
}

// Finds any diagnostic warnings inside a Poryscript script statement.
func (s *poryscriptServer) getScriptWarnings(ctx context.Context, fileUri string, script *ast.ScriptStatement) []lsp.Diagnostic {
	diagnostics := []lsp.Diagnostic{}
	for _, statement := range script.AllChildren() {
		cmd, ok := statement.(*ast.CommandStatement)
		if !ok {
			continue
		}
		// Check to see if the number of arguments given to the script command is valid.
		commands, _ := s.getCommands(ctx, fileUri)
		if command, ok := commands[cmd.Name.Value]; ok && command.Kind == parse.CommandScriptMacro {
			numRequiredParams := 0
			for _, p := range command.Parameters {
				if p.Kind == parse.CommandParamRequired {
					numRequiredParams++
				}
			}
			var message string
			if len(cmd.Args) < numRequiredParams {
				message = fmt.Sprintf("%s requires at least %s, but %s provided", cmd.Name.Value, getArgumentsString(numRequiredParams), getWasWereString(len(cmd.Args)))
			} else if len(cmd.Args) > len(command.Parameters) {
				word := "were"
				if len(cmd.Args) == 1 {
					word = "was"
				}
				message = fmt.Sprintf("%s expects a maximum of %s, but %d %s provided", cmd.Name.Value, getArgumentsString(len(command.Parameters)), len(cmd.Args), word)
			}
			if len(message) > 0 {
				diagnostics = append(diagnostics,
					lsp.Diagnostic{
						Range: lsp.Range{
							Start: lsp.Position{Line: cmd.Token.LineNumber - 1, Character: cmd.Token.StartCharIndex},
							End:   lsp.Position{Line: cmd.Token.EndLineNumber - 1, Character: cmd.Token.EndCharIndex},
						},
						Severity: lsp.Warning,
						Source:   "Poryscript",
						Message:  message,
					})
			}
		}
	}
	return diagnostics
}

// Finds any diagnostic warnings inside a Poryscript movement statement.
func (s *poryscriptServer) getMovementWarnings(ctx context.Context, fileUri string, script *ast.MovementStatement) []lsp.Diagnostic {
	diagnostics := []lsp.Diagnostic{}
	commands, _ := s.getCommands(ctx, fileUri)
	for _, cmd := range script.MovementCommands {
		if _, ok := commands[cmd.Literal]; !ok {
			diagnostics = append(diagnostics,
				lsp.Diagnostic{
					Range: lsp.Range{
						Start: lsp.Position{Line: cmd.LineNumber - 1, Character: cmd.StartCharIndex},
						End:   lsp.Position{Line: cmd.EndLineNumber - 1, Character: cmd.EndCharIndex},
					},
					Severity: lsp.Warning,
					Source:   "Poryscript",
					Message:  fmt.Sprintf("Unrecognized movement command \"%s\"", cmd.Literal),
				})
		}
	}
	return diagnostics
}

func getArgumentsString(n int) string {
	if n == 1 {
		return "1 argument"
	}
	return fmt.Sprintf("%d arguments", n)
}

func getWasWereString(n int) string {
	if n == 0 {
		return "none were"
	} else if n == 1 {
		return "only 1 was"
	} else {
		return fmt.Sprintf("only %d were", n)
	}
}
