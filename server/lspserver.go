package server

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/huderlem/poryscript-pls/lsp"
	"github.com/sourcegraph/jsonrpc2"
)

type LspServer interface {
	Run()
}

func New() LspServer {
	server := poryscriptServer{
		hasConfigCapability:                false,
		hasWorkspaceFolderCapability:       false,
		hasDiagnosticRelatedInfoCapability: false,
	}

	handler := jsonrpc2.HandlerWithError(buildRequestHandler(&server))
	server.connection = jsonrpc2.NewConn(context.Background(), jsonrpc2.NewBufferedStream(StdIoConn{}, jsonrpc2.VSCodeObjectCodec{}), handler)
	return server
}

func buildRequestHandler(server *poryscriptServer) func(ctx context.Context, connection *jsonrpc2.Conn, request *jsonrpc2.Request) (interface{}, error) {
	return func(ctx context.Context, connection *jsonrpc2.Conn, request *jsonrpc2.Request) (interface{}, error) {
		os.Stderr.WriteString(fmt.Sprintf("Handling request: %s\n", request.Method))
		switch request.Method {
		case "initialize":
			params := lsp.InitializeParams{}
			if err := json.Unmarshal(*request.Params, &params); err != nil {
				return nil, err
			}
			return server.onInitialize(ctx, params), nil
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
}

// poryscriptServer is the main handler for the Poryscript LSP server. It implements the
// LspServer interface.
type poryscriptServer struct {
	connection                         *jsonrpc2.Conn
	hasConfigCapability                bool
	hasWorkspaceFolderCapability       bool
	hasDiagnosticRelatedInfoCapability bool
}

// Runs the LSP server indefinitely.
func (s poryscriptServer) Run() {
	<-s.connection.DisconnectNotify()
}

// Handles an incoming LSP 'initialize' request.
// https://microsoft.github.io/language-server-protocol/specifications/lsp/3.17/specification/#initialize
func (s *poryscriptServer) onInitialize(ctx context.Context, params lsp.InitializeParams) *lsp.InitializeResult {
	s.hasConfigCapability = params.Capabilities.Workspace.Configuration
	s.hasWorkspaceFolderCapability = params.Capabilities.Workspace.WorkspaceFolders

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

// Handles an incoming LSP 'textDocument/completion' request.
// https://microsoft.github.io/language-server-protocol/specifications/lsp/3.17/specification/#textDocument_completion
func (s *poryscriptServer) onCompletion(ctx context.Context, req lsp.CompletionParams) (*[]lsp.CompletionItem, error) {
	return &[]lsp.CompletionItem{{
		Label:  "first",
		Kind:   lsp.CIKText,
		Detail: "Testing first...",
		Data:   1,
	}, {
		Label:  "second",
		Kind:   lsp.CIKText,
		Detail: "Testing second...",
		Data:   2,
	}, {
		Label:  "third",
		Kind:   lsp.CIKText,
		Detail: "Testing third...",
		Data:   3,
	}}, nil
}
