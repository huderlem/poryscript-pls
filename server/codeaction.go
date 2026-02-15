package server

import (
	"context"
	"encoding/json"
	"net/url"

	"github.com/huderlem/poryscript-pls/lsp"
	"github.com/huderlem/poryscript/lexer"
	"github.com/huderlem/poryscript/refactor"
	"github.com/huderlem/poryscript/token"
)

var styleNames = map[refactor.StringStyle]string{
	refactor.StyleAuto:         "Convert to auto string",
	refactor.StyleConcatenated: "Convert to concatenated strings",
	refactor.StyleSingleLine:   "Convert to single-line string",
}

var convertAllStyleNames = map[refactor.StringStyle]string{
	refactor.StyleAuto:         "Convert all strings in file to auto",
	refactor.StyleConcatenated: "Convert all strings in file to concatenated",
	refactor.StyleSingleLine:   "Convert all strings in file to single-line",
}

// codeActionData is the payload stored in CodeAction.Data so that
// codeAction/resolve can recompute the edit against fresh document content.
type codeActionData struct {
	URI         string `json:"uri"`
	Line        int    `json:"line"`
	Character   int    `json:"character"`
	TargetStyle int    `json:"targetStyle"`
	ConvertAll  bool   `json:"convertAll,omitempty"`
}

// onCodeAction handles the textDocument/codeAction request.
// It returns lightweight CodeActions without Edit — the edit is computed
// lazily in onCodeActionResolve to avoid stale edits.
func (s *poryscriptServer) onCodeAction(ctx context.Context, req lsp.CodeActionParams) ([]lsp.CodeAction, error) {
	uri, _ := url.QueryUnescape(string(req.TextDocument.URI))
	content, err := s.getDocumentContent(ctx, uri)
	if err != nil {
		return nil, err
	}

	// Tokenize the document.
	l := lexer.New(content)
	var tokens []token.Token
	for {
		t := l.NextToken()
		if t.Type == token.EOF {
			break
		}
		tokens = append(tokens, t)
	}

	// Find a string token at the cursor position, ignoring format() strings.
	tok, tokIdx, found := refactor.FindStringTokenAtPosition(tokens, req.Range.Start.Line, req.Range.Start.Character)
	if !found || refactor.IsFormatStringToken(tokens, tokIdx) {
		return nil, nil
	}

	// Extract source text and detect current style.
	sourceText := refactor.ExtractTokenSourceText(content, tok)
	if sourceText == "" {
		return nil, nil
	}
	currentStyle := refactor.DetectStringStyle(sourceText)

	var actions []lsp.CodeAction
	allStyles := []refactor.StringStyle{refactor.StyleAuto, refactor.StyleConcatenated, refactor.StyleSingleLine}

	// Per-string conversion actions (excluding the current style).
	for _, targetStyle := range allStyles {
		if targetStyle == currentStyle {
			continue
		}
		data, err := json.Marshal(codeActionData{
			URI:         string(req.TextDocument.URI),
			Line:        req.Range.Start.Line,
			Character:   req.Range.Start.Character,
			TargetStyle: int(targetStyle),
		})
		if err != nil {
			continue
		}
		actions = append(actions, lsp.CodeAction{
			Title: styleNames[targetStyle],
			Kind:  lsp.CAKRefactorRewrite,
			Data:  data,
		})
	}

	// Whole-file conversion actions (all 3 styles).
	for _, targetStyle := range allStyles {
		data, err := json.Marshal(codeActionData{
			URI:         string(req.TextDocument.URI),
			TargetStyle: int(targetStyle),
			ConvertAll:  true,
		})
		if err != nil {
			continue
		}
		actions = append(actions, lsp.CodeAction{
			Title: convertAllStyleNames[targetStyle],
			Kind:  lsp.CAKRefactorRewrite,
			Data:  data,
		})
	}

	return actions, nil
}

// onCodeActionResolve handles the codeAction/resolve request.
// It re-reads the document, re-finds string token(s), and computes the
// WorkspaceEdit fresh so that edits are never stale.
func (s *poryscriptServer) onCodeActionResolve(ctx context.Context, action lsp.CodeAction) (lsp.CodeAction, error) {
	var data codeActionData
	if err := json.Unmarshal(action.Data, &data); err != nil {
		return action, err
	}

	uri, _ := url.QueryUnescape(data.URI)
	content, err := s.getDocumentContent(ctx, uri)
	if err != nil {
		return action, err
	}

	// Tokenize the document.
	l := lexer.New(content)
	var tokens []token.Token
	for {
		t := l.NextToken()
		if t.Type == token.EOF {
			break
		}
		tokens = append(tokens, t)
	}

	targetStyle := refactor.StringStyle(data.TargetStyle)

	if data.ConvertAll {
		return resolveConvertAll(action, data.URI, content, tokens, targetStyle)
	}
	return resolveConvertSingle(action, data, content, tokens, targetStyle)
}

// resolveConvertSingle computes the edit for a single string conversion.
func resolveConvertSingle(action lsp.CodeAction, data codeActionData, content string, tokens []token.Token, targetStyle refactor.StringStyle) (lsp.CodeAction, error) {
	tok, tokIdx, found := refactor.FindStringTokenAtPosition(tokens, data.Line, data.Character)
	if !found || refactor.IsFormatStringToken(tokens, tokIdx) {
		return action, nil
	}

	sourceText := refactor.ExtractTokenSourceText(content, tok)
	if sourceText == "" {
		return action, nil
	}

	// If the string is already in the target style, return a no-op.
	if refactor.DetectStringStyle(sourceText) == targetStyle {
		return action, nil
	}

	indent := refactor.ComputeConversionIndent(content, tok, targetStyle)
	converted, err := refactor.ConvertString(sourceText, targetStyle, indent)
	if err != nil {
		return action, err
	}

	action.Edit = &lsp.WorkspaceEdit{
		Changes: map[string][]lsp.TextEdit{
			data.URI: {
				{
					Range:   tokenToLSPRange(tok),
					NewText: converted,
				},
			},
		},
	}
	return action, nil
}

// resolveConvertAll computes edits to convert every string in the file.
func resolveConvertAll(action lsp.CodeAction, docURI, content string, tokens []token.Token, targetStyle refactor.StringStyle) (lsp.CodeAction, error) {
	var edits []lsp.TextEdit
	for i, tok := range tokens {
		if !token.IsStringLikeToken(tok.Type) {
			continue
		}
		if refactor.IsFormatStringToken(tokens, i) {
			continue
		}
		sourceText := refactor.ExtractTokenSourceText(content, tok)
		if sourceText == "" {
			continue
		}
		if refactor.DetectStringStyle(sourceText) == targetStyle {
			continue
		}
		indent := refactor.ComputeConversionIndent(content, tok, targetStyle)
		converted, err := refactor.ConvertString(sourceText, targetStyle, indent)
		if err != nil {
			continue
		}
		edits = append(edits, lsp.TextEdit{
			Range:   tokenToLSPRange(tok),
			NewText: converted,
		})
	}

	if len(edits) > 0 {
		action.Edit = &lsp.WorkspaceEdit{
			Changes: map[string][]lsp.TextEdit{
				docURI: edits,
			},
		}
	}
	return action, nil
}

// tokenToLSPRange converts a token's position to an LSP Range.
func tokenToLSPRange(tok token.Token) lsp.Range {
	return lsp.Range{
		Start: lsp.Position{
			Line:      tok.LineNumber - 1,
			Character: tok.StartUtf8CharIndex,
		},
		End: lsp.Position{
			Line:      tok.EndLineNumber - 1,
			Character: tok.EndUtf8CharIndex,
		},
	}
}
