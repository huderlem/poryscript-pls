package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"sync"

	"github.com/huderlem/poryscript-pls/config"
	"github.com/huderlem/poryscript-pls/lsp"
	"github.com/huderlem/poryscript-pls/parse"
	"github.com/huderlem/poryscript/lexer"
	"github.com/huderlem/poryscript/token"
	"github.com/sourcegraph/jsonrpc2"
)

type LspServer interface {
	Run()
}

func New() LspServer {
	server := poryscriptServer{
		config:           config.New(),
		cachedDocuments:  map[string]string{},
		cachedCommands:   map[string]map[string]parse.Command{},
		cachedConstants:  map[string]map[string]parse.ConstantSymbol{},
		cachedSymbols:    map[string]map[string]parse.Symbol{},
		cachedMiscTokens: map[string]map[string]parse.MiscToken{},
	}

	// Wrap with AsyncHandler to allow for calling client requests in the middle of
	// handling a request. Otherwise, a channel deadlock will occur and cause a panic.
	handler := jsonrpc2.AsyncHandler(jsonrpc2.HandlerWithError(server.handle))
	server.connection = jsonrpc2.NewConn(context.Background(), jsonrpc2.NewBufferedStream(StdioRWC{}, jsonrpc2.VSCodeObjectCodec{}), handler)
	return &server
}

func (server *poryscriptServer) handle(ctx context.Context, conn *jsonrpc2.Conn, request *jsonrpc2.Request) (interface{}, error) {
	switch request.Method {
	case "initialize":
		params := lsp.InitializeParams{}
		if err := json.Unmarshal(*request.Params, &params); err != nil {
			return nil, err
		}
		return server.onInitialize(ctx, params), nil
	case "initialized":
		return nil, server.onInitialized(ctx)
	case "textDocument/completion":
		params := lsp.CompletionParams{}
		if err := json.Unmarshal(*request.Params, &params); err != nil {
			return nil, err
		}
		return server.onCompletion(ctx, params)
	case "textDocument/definition":
		params := lsp.DefinitionParams{}
		if err := json.Unmarshal(*request.Params, &params); err != nil {
			return nil, err
		}
		return server.onDefinition(ctx, params)
	case "textDocument/signatureHelp":
		params := lsp.SignatureHelpParams{}
		if err := json.Unmarshal(*request.Params, &params); err != nil {
			return nil, err
		}
		return server.onSignatureHelp(ctx, params)
	case "textDocument/semanticTokens/full":
		params := lsp.SemanticTokensParams{}
		if err := json.Unmarshal(*request.Params, &params); err != nil {
			return nil, err
		}
		return server.onSemanticTokensFull(ctx, params)
	case "textDocument/didChange":
		params := lsp.DidChangeTextDocumentParams{}
		if err := json.Unmarshal(*request.Params, &params); err != nil {
			return nil, err
		}
		return nil, server.onTextDocumentDidChange(ctx, params)
	default:
		return nil, fmt.Errorf("unsupported request method '%s'", request.Method)
	}
}

// poryscriptServer is the main handler for the Poryscript LSP server. It implements the
// LspServer interface.
type poryscriptServer struct {
	connection       *jsonrpc2.Conn
	config           config.Config
	cachedDocuments  map[string]string
	cachedCommands   map[string]map[string]parse.Command
	cachedConstants  map[string]map[string]parse.ConstantSymbol
	cachedSymbols    map[string]map[string]parse.Symbol
	cachedMiscTokens map[string]map[string]parse.MiscToken
	documentsMutex   sync.Mutex
	commandsMutex    sync.Mutex
	constantsMutex   sync.Mutex
	symbolsMutex     sync.Mutex
	miscTokensMutex  sync.Mutex
}

// Runs the LSP server indefinitely.
func (s *poryscriptServer) Run() {
	<-s.connection.DisconnectNotify()
}

// Handles an incoming LSP 'initialize' request.
// https://microsoft.github.io/language-server-protocol/specifications/specification-current/#initialize
func (s *poryscriptServer) onInitialize(ctx context.Context, params lsp.InitializeParams) *lsp.InitializeResult {
	s.config.HasConfigCapability = params.Capabilities.Workspace.Configuration
	s.config.HasWorkspaceFolderCapability = params.Capabilities.Workspace.WorkspaceFolders

	return &lsp.InitializeResult{
		Capabilities: lsp.ServerCapabilities{
			TextDocumentSync: &lsp.TextDocumentSyncOptionsOrKind{
				Options: &lsp.TextDocumentSyncOptions{
					OpenClose: true,
					Change:    lsp.TDSKFull,
				},
			},
			CompletionProvider: &lsp.CompletionOptions{},
			SignatureHelpProvider: &lsp.SignatureHelpOptions{
				TriggerCharacters: []string{"(", ","},
			},
			SemanticTokensProvider: &lsp.SemanticTokensOptions{
				Full:  lsp.STPFFull,
				Range: false,
				Legend: lsp.SemanticTokensLegend{
					TokenTypes: []string{"keyword", "function", "enumMember", "variable"},
				},
			},
			DefinitionProvider: true,
		},
	}
}

// Handles an incoming LSP 'initialized' notification.
// https://microsoft.github.io/language-server-protocol/specifications/specification-current/#initialized
func (s *poryscriptServer) onInitialized(ctx context.Context) error {
	if s.config.HasConfigCapability {
		params := lsp.RegistrationParams{
			Registrations: []lsp.Registration{
				{
					ID:     "workspace/didChangeConfiguration",
					Method: "workspace/didChangeConfiguration",
				},
			},
		}
		var result interface{}
		s.connection.Call(ctx, "client/registerCapability", params, &result)
	}
	if s.config.HasWorkspaceFolderCapability {
		params := lsp.RegistrationParams{
			Registrations: []lsp.Registration{
				{
					ID:     "workspace/didChangeWorkspaceFolders",
					Method: "workspace/didChangeWorkspaceFolders",
				},
			},
		}
		var result interface{}
		s.connection.Call(ctx, "client/registerCapability", params, &result)
	}
	var filepaths []string
	if err := s.connection.Call(ctx, "poryscript/getPoryscriptFiles", nil, &filepaths); err != nil {
		os.Stderr.WriteString(err.Error())
	}
	for _, filepath := range filepaths {
		if _, err := s.getSymbolsInFile(ctx, "file://"+filepath); err != nil {
			os.Stderr.WriteString(err.Error())
		}
	}
	return nil
}

// Handles an incoming LSP 'textDocument/completion' request.
// https://microsoft.github.io/language-server-protocol/specifications/specification-current/#textDocument_completion
func (s *poryscriptServer) onCompletion(ctx context.Context, req lsp.CompletionParams) ([]lsp.CompletionItem, error) {
	commands, _ := s.getCommands(ctx, string(req.TextDocument.URI))
	constants, _ := s.getConstantsInFile(ctx, string(req.TextDocument.URI))
	miscTokens, _ := s.getMiscTokens(ctx, string(req.TextDocument.URI))

	s.getSymbolsInFile(ctx, string(req.TextDocument.URI))
	symbols := []parse.Symbol{}
	for _, fileSymbols := range s.cachedSymbols {
		for _, s := range fileSymbols {
			symbols = append(symbols, s)
		}
	}

	completionItems := []lsp.CompletionItem{}
	for _, command := range commands {
		completionItems = append(completionItems, command.ToCompletionItem())
	}
	for _, constant := range constants {
		completionItems = append(completionItems, constant.ToCompletionItem())
	}
	for _, symbol := range symbols {
		completionItems = append(completionItems, symbol.ToCompletionItem())
	}
	for _, miscToken := range miscTokens {
		completionItems = append(completionItems, miscToken.ToCompletionItem())
	}
	return completionItems, nil
}

// Handles an incoming LSP 'textDocument/definition' request.
// https://microsoft.github.io/language-server-protocol/specifications/specification-current/#textDocument_definition
func (s *poryscriptServer) onDefinition(ctx context.Context, req lsp.DefinitionParams) ([]lsp.Location, error) {
	content, err := s.getDocumentContent(ctx, string(req.TextDocument.URI))
	if err != nil {
		return []lsp.Location{}, err
	}
	token := parse.GetTokenAt(content, req.Position.Line, req.Position.Character)

	constants, _ := s.getConstantsInFile(ctx, string(req.TextDocument.URI))
	if c, ok := constants[token]; ok {
		return []lsp.Location{c.ToLocation()}, nil
	}

	s.getSymbolsInFile(ctx, string(req.TextDocument.URI))
	symbols := map[string]parse.Symbol{}
	for _, fileSymbols := range s.cachedSymbols {
		for _, s := range fileSymbols {
			symbols[s.Name] = s
		}
	}
	if s, ok := symbols[token]; ok {
		return []lsp.Location{s.ToLocation()}, nil
	}

	miscTokens, _ := s.getMiscTokens(ctx, string(req.TextDocument.URI))
	if t, ok := miscTokens[token]; ok {
		return []lsp.Location{t.ToLocation()}, nil
	}

	return []lsp.Location{}, nil
}

// Handles an incoming LSP 'textDocument/signatureHelp' request.
// https://microsoft.github.io/language-server-protocol/specifications/specification-current/#textDocument_signatureHelp
func (s *poryscriptServer) onSignatureHelp(ctx context.Context, req lsp.SignatureHelpParams) (lsp.SignatureHelp, error) {
	content, err := s.getDocumentContent(ctx, string(req.TextDocument.URI))
	if err != nil {
		return lsp.SignatureHelp{}, err
	}

	callInfo, err := parse.GetCommandCallParts(content, req.Position.Line, req.Position.Character)
	if err != nil {
		// TODO: log error?
		return lsp.SignatureHelp{}, nil
	}

	commands, _ := s.getCommands(ctx, string(req.TextDocument.URI))
	command, ok := commands[callInfo.Command]

	if !ok || len(command.Parameters) == 0 {
		return lsp.SignatureHelp{}, nil
	}
	if req.Position.Character < callInfo.OpenParen.Character+1 || req.Position.Character > callInfo.CloseParen.Character {
		return lsp.SignatureHelp{}, nil
	}

	paramId := 0
	for paramId < len(callInfo.Commas) && req.Position.Character > callInfo.Commas[paramId].Character {
		paramId++
	}

	return lsp.SignatureHelp{
		ActiveParameter: paramId,
		ActiveSignature: 0,
		Signatures: []lsp.SignatureInformation{
			{
				Label:         command.GetParamsLabel(),
				Documentation: command.Documentation,
				Parameters:    command.GetParamInformation(),
			},
		},
	}, nil
}

// Handles an incoming LSP 'textDocument/didChange' request.
// https://microsoft.github.io/language-server-protocol/specifications/specification-current/#textDocument_didChange
func (s *poryscriptServer) onTextDocumentDidChange(ctx context.Context, req lsp.DidChangeTextDocumentParams) error {
	if len(req.ContentChanges) == 0 {
		return nil
	}
	fileUri, _ := url.QueryUnescape(string(req.TextDocument.URI))
	s.clearCaches(fileUri)
	return nil
}

// Handles an incoming LSP 'textDocument/semanticTokens/full' request.
// https://microsoft.github.io/language-server-protocol/specifications/specification-current/#textDocument_semanticTokens
func (s *poryscriptServer) onSemanticTokensFull(ctx context.Context, req lsp.SemanticTokensParams) (lsp.SemanticTokens, error) {
	content, err := s.getDocumentContent(ctx, string(req.TextDocument.URI))
	if err != nil {
		return lsp.SemanticTokens{}, err
	}

	// Collect the tokens.
	l := lexer.New(content)
	tokens := []token.Token{}
	for {
		t := l.NextToken()
		if t.Type == token.EOF {
			break
		}
		tokens = append(tokens, t)
	}

	commands, _ := s.getCommands(ctx, string(req.TextDocument.URI))
	constants, _ := s.getConstantsInFile(ctx, string(req.TextDocument.URI))
	miscTokens, _ := s.getMiscTokens(ctx, string(req.TextDocument.URI))
	s.getSymbolsInFile(ctx, string(req.TextDocument.URI))
	symbols := map[string]parse.Symbol{}
	for _, fileSymbols := range s.cachedSymbols {
		for _, s := range fileSymbols {
			symbols[s.Name] = s
		}
	}

	// TODO: use strongly-typed token types for AddToken(), rather than hardcoded integers
	builder := lsp.SemanticTokenBuilder{}
	for _, t := range tokens {
		if command, ok := commands[t.Literal]; ok {
			// 'switch' and 'case' are both Poryscript keywords and scripting commands.
			if t.Literal != "switch" && t.Literal != "case" {
				switch command.Kind {
				case lsp.CIKFunction:
					builder.AddToken(t.LineNumber-1, t.StartCharIndex, t.EndCharIndex-t.StartCharIndex, 1, 0)
				case lsp.CIKConstant:
					builder.AddToken(t.LineNumber-1, t.StartCharIndex, t.EndCharIndex-t.StartCharIndex, 0, 0)
				}
			}
		}

		if constant, ok := constants[t.Literal]; ok {
			if t.LineNumber-1 != constant.Position.Line {
				builder.AddToken(t.LineNumber-1, t.StartCharIndex, t.EndCharIndex-t.StartCharIndex, 2, 0)
			}
		}

		if symbol, ok := symbols[t.Literal]; ok {
			switch symbol.Kind {
			case parse.SymbolKindScript, parse.SymbolKindMapScripts:
				builder.AddToken(t.LineNumber-1, t.StartCharIndex, t.EndCharIndex-t.StartCharIndex, 1, 0)
			case parse.SymbolKindMovementScript, parse.SymbolKindText:
				builder.AddToken(t.LineNumber-1, t.StartCharIndex, t.EndCharIndex-t.StartCharIndex, 3, 0)
			}
		}

		if miscToken, ok := miscTokens[t.Literal]; ok {
			switch miscToken.Type {
			case "special":
				builder.AddToken(t.LineNumber-1, t.StartCharIndex, t.EndCharIndex-t.StartCharIndex, 1, 0)
			case "define":
				builder.AddToken(t.LineNumber-1, t.StartCharIndex, t.EndCharIndex-t.StartCharIndex, 2, 0)
			default:
				builder.AddToken(t.LineNumber-1, t.StartCharIndex, t.EndCharIndex-t.StartCharIndex, 0, 0)
			}
		}
	}

	return lsp.SemanticTokens{Data: builder.Build()}, nil
}
