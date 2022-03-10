package config

import (
	"context"
	"fmt"

	"github.com/huderlem/poryscript-pls/lsp"
	"github.com/sourcegraph/jsonrpc2"
)

// Configuration for the Poryscript language server.
type Config struct {
	ResourceSettings                   map[string]PoryscriptSettings
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
}

type TokenIncludeSetting struct {
	Expression string `json:"expression"`
	Type       string `json:"type"`
	File       string `json:"file"`
}

var defaultPoryscriptSettings = PoryscriptSettings{
	CommandIncludes: []string{"asm/macros/event.inc", "asm/macros/movement.inc"},
	SymbolIncludes:  []TokenIncludeSetting{},
}

func New() Config {
	return Config{
		ResourceSettings:                   map[string]PoryscriptSettings{},
		HasConfigCapability:                false,
		HasWorkspaceFolderCapability:       false,
		HasDiagnosticRelatedInfoCapability: false,
	}
}

// GetResourceSettings retrieves the PoryscriptSettings associated with the given resource.
// If the client doesn't support configuration, it returns the default settings.
// Settings are cached on a resource-by-resource basis..
func (c *Config) GetResourceSettings(ctx context.Context, conn jsonrpc2.JSONRPC2, resource string) (PoryscriptSettings, error) {
	if !c.HasConfigCapability {
		return defaultPoryscriptSettings, nil
	}
	if settings, ok := c.ResourceSettings[resource]; ok {
		return settings, nil
	}
	settings, err := c.fetchResourceSettings(ctx, conn, resource)
	if err != nil {
		return PoryscriptSettings{}, err
	}
	c.ResourceSettings[resource] = settings
	return settings, nil
}

// Fetches the configuration from the client for the given resource.
func (c *Config) fetchResourceSettings(ctx context.Context, conn jsonrpc2.JSONRPC2, resource string) (PoryscriptSettings, error) {
	params := lsp.ConfigurationParams{
		Items: []lsp.ConfigurationItem{
			{
				// ScopeURI: resource,
				Section: "languageServerPoryscript",
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
