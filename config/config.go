package config

import (
	"context"
	"fmt"
	"sync"

	"github.com/huderlem/poryscript-pls/lsp"
	"github.com/sourcegraph/jsonrpc2"
)

var lock = sync.Mutex{}

// Configuration for the Poryscript language server.
type Config struct {
	FileSettings                       map[string]PoryscriptSettings
	HasConfigCapability                bool
	HasWorkspaceFolderCapability       bool
	HasDiagnosticRelatedInfoCapability bool
}

// Settings for the Poryscript language server. These are controlled
// by the client.
type PoryscriptSettings struct {
	// Filepaths for script macro definitions.
	CommandIncludes []string `json:"commandIncludes"`
	// Filepaths for constant and symbol definitions.
	SymbolIncludes []TokenIncludeSetting `json:"symbolIncludes"`
	// Filepath for command config.
	CommandConfigFilepath string `json:"commandConfigFilepath"`
}

type TokenIncludeSetting struct {
	Expression string `json:"expression"`
	Type       string `json:"type"`
	File       string `json:"file"`
}

var defaultPoryscriptSettings = PoryscriptSettings{
	CommandIncludes:       []string{"asm/macros/event.inc", "asm/macros/movement.inc"},
	SymbolIncludes:        []TokenIncludeSetting{},
	CommandConfigFilepath: "tools/poryscript/command_config.json",
}

func New() Config {
	return Config{
		FileSettings:                       map[string]PoryscriptSettings{},
		HasConfigCapability:                false,
		HasWorkspaceFolderCapability:       false,
		HasDiagnosticRelatedInfoCapability: false,
	}
}

// Completely clears the cached settings.
func (c *Config) ClearSettings() {
	c.FileSettings = map[string]PoryscriptSettings{}
}

// GetFileSettings retrieves the PoryscriptSettings associated with the given file.
// If the client doesn't support configuration, it returns the default settings.
// Settings are cached on a filepath-by-filepath basis..
func (c *Config) GetFileSettings(ctx context.Context, conn jsonrpc2.JSONRPC2, filepath string) (PoryscriptSettings, error) {
	if !c.HasConfigCapability {
		return defaultPoryscriptSettings, nil
	}
	lock.Lock()
	defer lock.Unlock()
	if settings, ok := c.FileSettings[filepath]; ok {
		return settings, nil
	}
	settings, err := c.fetchFileSettings(ctx, conn, filepath)
	if err != nil {
		return PoryscriptSettings{}, err
	}
	c.FileSettings[filepath] = settings
	return settings, nil
}

// Fetches the configuration from the client for the given file.
func (c *Config) fetchFileSettings(ctx context.Context, conn jsonrpc2.JSONRPC2, filepath string) (PoryscriptSettings, error) {
	params := lsp.ConfigurationParams{
		Items: []lsp.ConfigurationItem{
			{
				ScopeURI: filepath,
				Section:  "languageServerPoryscript",
			},
		},
	}
	result := &[]PoryscriptSettings{}
	if err := conn.Call(ctx, "workspace/configuration", params, result); err != nil {
		return PoryscriptSettings{}, err
	}
	if len(*result) != 1 {
		return PoryscriptSettings{}, fmt.Errorf("failed to fetch config settings. Expected result arry to be one element, but received %d elements instead", len(*result))
	}

	return (*result)[0], nil
}
