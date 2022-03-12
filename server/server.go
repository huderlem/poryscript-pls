package server

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/huderlem/poryscript-pls/config"
	"github.com/huderlem/poryscript-pls/lsp"
	"github.com/huderlem/poryscript-pls/parse"
	"github.com/sourcegraph/jsonrpc2"
)

type LspServer interface {
	Run()
}

func New() LspServer {
	server := poryscriptServer{
		config:          config.New(),
		cachedCommands:  map[string][]parse.Command{},
		cachedConstants: map[string][]parse.ConstantSymbol{},
		cachedSymbols:   map[string][]parse.Symbol{},
	}

	// Wrap with AsyncHandler to allow for calling client requests in the middle of
	// handling a request. Otherwise, a channel deadlock will occur and cause a panic.
	handler := jsonrpc2.AsyncHandler(jsonrpc2.HandlerWithError(server.handle))
	server.connection = jsonrpc2.NewConn(context.Background(), jsonrpc2.NewBufferedStream(StdioRWC{}, jsonrpc2.VSCodeObjectCodec{}), handler)
	return server
}

func (server *poryscriptServer) handle(ctx context.Context, conn *jsonrpc2.Conn, request *jsonrpc2.Request) (interface{}, error) {
	os.Stderr.WriteString(fmt.Sprintf("Handling request: %s\n", request.Method))
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
	default:
		return nil, fmt.Errorf("unsupported request method '%s'", request.Method)
	}
}

// poryscriptServer is the main handler for the Poryscript LSP server. It implements the
// LspServer interface.
type poryscriptServer struct {
	connection      *jsonrpc2.Conn
	config          config.Config
	cachedCommands  map[string][]parse.Command
	cachedConstants map[string][]parse.ConstantSymbol
	cachedSymbols   map[string][]parse.Symbol
}

// Runs the LSP server indefinitely.
func (s poryscriptServer) Run() {
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
					Change:    lsp.TDSKIncremental,
				},
			},
			CompletionProvider: &lsp.CompletionOptions{},
			SignatureHelpProvider: &lsp.SignatureHelpOptions{
				TriggerCharacters: []string{"(", ","},
			},
			SemanticTokensProvider: &lsp.SemanticTokensOptions{
				Full:  lsp.STPFFull,
				Range: true,
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
	return nil
}

// Handles an incoming LSP 'textDocument/completion' request.
// https://microsoft.github.io/language-server-protocol/specifications/specification-current/#textDocument_completion
func (s *poryscriptServer) onCompletion(ctx context.Context, req lsp.CompletionParams) ([]lsp.CompletionItem, error) {
	commands, _ := s.getCommands(ctx, string(req.TextDocument.URI))
	constants, _ := s.getConstantsInFile(ctx, string(req.TextDocument.URI))
	symbols, _ := s.getSymbolsInFile(ctx, string(req.TextDocument.URI))

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
	return completionItems, nil
}
