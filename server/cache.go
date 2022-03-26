package server

import (
	"context"
	"net/url"

	"github.com/huderlem/poryscript-pls/parse"
)

// Gets the aggregate list of Commands from the collection of files that define
// the Commands. The Commands are cached for the given file so that parsing is avoided
// in future calls.
func (s *poryscriptServer) getCommands(ctx context.Context, file string) (map[string]parse.Command, error) {
	settings, err := s.config.GetFileSettings(ctx, s.connection, file)
	if err != nil {
		return nil, err
	}
	// Aggregate a slice of Commands from all of the relevant files.
	commands := map[string]parse.Command{}
	for _, includeFilepath := range settings.CommandIncludes {
		fileCommands, err := s.getCommandsInFile(ctx, includeFilepath)
		if err != nil {
			// TODO: log error? we don't want to fail if a single file resulted in an error.
			continue
		}
		for _, command := range fileCommands {
			commands[command.Name] = command
		}
	}
	for _, command := range parse.KeywordCommands {
		commands[command.Name] = command
	}
	return commands, nil
}

// Gets the list of Commands from the given file. The Commands
// are cached for the given file so that parsing is avoided in future
// calls.
func (s *poryscriptServer) getCommandsInFile(ctx context.Context, file string) (map[string]parse.Command, error) {
	file, _ = url.QueryUnescape(file)
	if commands, ok := s.cachedCommands[file]; ok {
		return commands, nil
	}
	return s.getAndCacheCommandsInFile(ctx, file)
}

// Fetches and caches the Commands from the given file.
func (s *poryscriptServer) getAndCacheCommandsInFile(ctx context.Context, file string) (map[string]parse.Command, error) {
	file, _ = url.QueryUnescape(file)
	var content string
	if err := s.connection.Call(ctx, "poryscript/readfile", file, &content); err != nil {
		return nil, err
	}
	if !s.config.HasWorkspaceFolderCapability {
		return nil, nil
	}
	commands := parse.ParseCommands(content)
	commandSet := map[string]parse.Command{}
	for _, c := range commands {
		commandSet[c.Name] = c
	}
	s.cachedCommands[file] = commandSet
	return commandSet, nil
}

// Gets the list of poryscript constants from the given file. The constants
// are cached for the given file so that parsing is avoided in future calls.
func (s *poryscriptServer) getConstantsInFile(ctx context.Context, file string) (map[string]parse.ConstantSymbol, error) {
	file, _ = url.QueryUnescape(file)
	if constants, ok := s.cachedConstants[file]; ok {
		return constants, nil
	}
	return s.getAndCacheConstantsInFile(ctx, file)
}

// Fetches and caches the poryscript constants from the given file.
func (s *poryscriptServer) getAndCacheConstantsInFile(ctx context.Context, file string) (map[string]parse.ConstantSymbol, error) {
	file, _ = url.QueryUnescape(file)
	content, err := s.getDocumentContent(ctx, file)
	if err != nil {
		return nil, err
	}
	constants := parse.ParseConstants(content, file)
	constantSet := map[string]parse.ConstantSymbol{}
	for _, c := range constants {
		constantSet[c.Name] = c
	}
	s.cachedConstants[file] = constantSet
	return constantSet, nil
}

// Gets the list of poryscript symbols from the given file. The symbols
// are cached for the given file so that parsing is avoided in future calls.
func (s *poryscriptServer) getSymbolsInFile(ctx context.Context, file string) (map[string]parse.Symbol, error) {
	file, _ = url.QueryUnescape(file)
	if symbols, ok := s.cachedSymbols[file]; ok {
		return symbols, nil
	}
	return s.getAndCacheSymbolsInFile(ctx, file)
}

// Fetches and caches the poryscript symbols from the given file.
func (s *poryscriptServer) getAndCacheSymbolsInFile(ctx context.Context, file string) (map[string]parse.Symbol, error) {
	file, _ = url.QueryUnescape(file)
	content, err := s.getDocumentContent(ctx, file)
	if err != nil {
		return nil, err
	}
	symbols := parse.ParseSymbols(content, file)
	symbolSet := map[string]parse.Symbol{}
	for _, s := range symbols {
		symbolSet[s.Name] = s
	}
	s.cachedSymbols[file] = symbolSet
	return symbolSet, nil
}

// Gets the aggregate list of miscellaneous tokens from the collection of files
// specified in the settings.
func (s *poryscriptServer) getMiscTokens(ctx context.Context, file string) (map[string]parse.MiscToken, error) {
	file, _ = url.QueryUnescape(file)
	settings, err := s.config.GetFileSettings(ctx, s.connection, file)
	if err != nil {
		return nil, err
	}
	miscTokens := map[string]parse.MiscToken{}
	for _, includeSetting := range settings.SymbolIncludes {
		tokens, err := s.getMiscTokensInFile(ctx, includeSetting.Expression, includeSetting.Type, includeSetting.File)
		if err != nil {
			// TODO: log error?
			continue
		}
		for _, t := range tokens {
			miscTokens[t.Name] = t
		}
	}
	return miscTokens, nil
}

// Gets the list of miscellaneous tokens from the given file. The tokens
// are cached for the given file so that parsing is avoided in future calls.
func (s *poryscriptServer) getMiscTokensInFile(ctx context.Context, expression, tokenType, file string) (map[string]parse.MiscToken, error) {
	file, _ = url.QueryUnescape(file)
	if tokens, ok := s.cachedMiscTokens[file+expression]; ok {
		return tokens, nil
	}
	return s.getAndCacheMiscTokensInFile(ctx, expression, tokenType, file)
}

// Fetches and caches the miscellaneous tokens from the given file.
func (s *poryscriptServer) getAndCacheMiscTokensInFile(ctx context.Context, expression, tokenType, file string) (map[string]parse.MiscToken, error) {
	file, _ = url.QueryUnescape(file)
	var content string
	if err := s.connection.Call(ctx, "poryscript/readfile", file, &content); err != nil {
		return nil, err
	}
	var fileUri string
	if err := s.connection.Call(ctx, "poryscript/getfileuri", file, &fileUri); err != nil {
		return nil, err
	}
	tokens := parse.ParseMiscTokens(content, expression, tokenType, fileUri)
	tokenSet := map[string]parse.MiscToken{}
	for _, t := range tokens {
		tokenSet[t.Name] = t
	}
	s.cachedMiscTokens[file+expression] = tokenSet
	return tokenSet, nil
}

// Gets the content for the given file. The content is cached
// for the given file so that parsing is avoided in future calls.//
func (s *poryscriptServer) getDocumentContent(ctx context.Context, file string) (string, error) {
	file, _ = url.QueryUnescape(file)
	if content, ok := s.cachedDocuments[file]; ok {
		return content, nil
	}
	return s.getAndCacheDocumentContent(ctx, file)
}

// Fetches and caches the content for the given file.
func (s *poryscriptServer) getAndCacheDocumentContent(ctx context.Context, file string) (string, error) {
	file, _ = url.QueryUnescape(file)
	var content string
	if err := s.connection.Call(ctx, "poryscript/readfs", file, &content); err != nil {
		return "", err
	}
	s.cachedDocuments[file] = content
	return content, nil
}
