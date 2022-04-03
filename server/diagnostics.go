package server

import (
	"context"
	"errors"

	"github.com/huderlem/poryscript-pls/lsp"
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
	_, err = p.ParseProgram()
	if err == nil {
		s.connection.Notify(ctx, "textDocument/publishDiagnostics", diagnostics)
		return nil
	}

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
	s.connection.Notify(ctx, "textDocument/publishDiagnostics", diagnostics)
	return nil
}
