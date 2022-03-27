package server

import (
	"context"
	"net/url"

	"github.com/huderlem/poryscript-pls/parse"
)

// Gets the aggregate list of Commands from the collection of files that define
// the Commands. The Commands are cached for the given file uri so that parsing is
// avoided in future calls.
func (s *poryscriptServer) getCommands(ctx context.Context, uri string) (map[string]parse.Command, error) {
	settings, err := s.config.GetFileSettings(ctx, s.connection, uri)
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

// Gets the list of Commands from the given file uri. The Commands
// are cached for the file so that parsing is avoided in future
// calls.
func (s *poryscriptServer) getCommandsInFile(ctx context.Context, uri string) (map[string]parse.Command, error) {
	s.commandsMutex.Lock()
	defer s.commandsMutex.Unlock()

	uri, _ = url.QueryUnescape(uri)
	if commands, ok := s.cachedCommands[uri]; ok {
		return commands, nil
	}
	return s.getAndCacheCommandsInFile(ctx, uri)
}

// Fetches and caches the Commands from the given file uri.
func (s *poryscriptServer) getAndCacheCommandsInFile(ctx context.Context, uri string) (map[string]parse.Command, error) {
	uri, _ = url.QueryUnescape(uri)
	var content string
	if err := s.connection.Call(ctx, "poryscript/readfile", uri, &content); err != nil {
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
	s.cachedCommands[uri] = commandSet
	return commandSet, nil
}

// Gets the list of poryscript constants from the given file uri. The constants
// are cached for the file so that parsing is avoided in future calls.
func (s *poryscriptServer) getConstantsInFile(ctx context.Context, uri string) (map[string]parse.ConstantSymbol, error) {
	s.constantsMutex.Lock()
	defer s.constantsMutex.Unlock()

	uri, _ = url.QueryUnescape(uri)
	if constants, ok := s.cachedConstants[uri]; ok {
		return constants, nil
	}
	return s.getAndCacheConstantsInFile(ctx, uri)
}

// Fetches and caches the poryscript constants from the given file uri.
func (s *poryscriptServer) getAndCacheConstantsInFile(ctx context.Context, uri string) (map[string]parse.ConstantSymbol, error) {
	uri, _ = url.QueryUnescape(uri)
	content, err := s.getDocumentContent(ctx, uri)
	if err != nil {
		return nil, err
	}
	constants := parse.ParseConstants(content, uri)
	constantSet := map[string]parse.ConstantSymbol{}
	for _, c := range constants {
		constantSet[c.Name] = c
	}
	s.cachedConstants[uri] = constantSet
	return constantSet, nil
}

// Gets the list of poryscript symbols from the given file uri. The symbols
// are cached for the file so that parsing is avoided in future calls.
func (s *poryscriptServer) getSymbolsInFile(ctx context.Context, uri string) (map[string]parse.Symbol, error) {
	s.symbolsMutex.Lock()
	defer s.symbolsMutex.Unlock()

	uri, _ = url.QueryUnescape(uri)
	if symbols, ok := s.cachedSymbols[uri]; ok {
		return symbols, nil
	}
	return s.getAndCacheSymbolsInFile(ctx, uri)
}

// Fetches and caches the poryscript symbols from the given file uri.
func (s *poryscriptServer) getAndCacheSymbolsInFile(ctx context.Context, uri string) (map[string]parse.Symbol, error) {
	uri, _ = url.QueryUnescape(uri)
	content, err := s.getDocumentContent(ctx, uri)
	if err != nil {
		return nil, err
	}
	symbols := parse.ParseSymbols(content, uri)
	symbolSet := map[string]parse.Symbol{}
	for _, s := range symbols {
		symbolSet[s.Name] = s
	}
	s.cachedSymbols[uri] = symbolSet
	return symbolSet, nil
}

// Gets the aggregate list of miscellaneous tokens from the collection of files
// specified in the settings.
func (s *poryscriptServer) getMiscTokens(ctx context.Context, uri string) (map[string]parse.MiscToken, error) {
	uri, _ = url.QueryUnescape(uri)
	settings, err := s.config.GetFileSettings(ctx, s.connection, uri)
	if err != nil {
		return nil, err
	}
	miscTokens := map[string]parse.MiscToken{}
	for _, includeSetting := range settings.SymbolIncludes {
		s.miscTokensMutex.Lock()
		tokens, err := s.getMiscTokensInFile(ctx, includeSetting.Expression, includeSetting.Type, includeSetting.File)
		s.miscTokensMutex.Unlock()
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

// Gets the list of miscellaneous tokens from the given file uri. The tokens
// are cached for the file so that parsing is avoided in future calls.
func (s *poryscriptServer) getMiscTokensInFile(ctx context.Context, expression, tokenType, uri string) (map[string]parse.MiscToken, error) {
	uri, _ = url.QueryUnescape(uri)
	if tokens, ok := s.cachedMiscTokens[uri+expression]; ok {
		return tokens, nil
	}
	return s.getAndCacheMiscTokensInFile(ctx, expression, tokenType, uri)
}

// Fetches and caches the miscellaneous tokens from the given file uri.
func (s *poryscriptServer) getAndCacheMiscTokensInFile(ctx context.Context, expression, tokenType, uri string) (map[string]parse.MiscToken, error) {
	uri, _ = url.QueryUnescape(uri)
	var content string
	if err := s.connection.Call(ctx, "poryscript/readfile", uri, &content); err != nil {
		return nil, err
	}
	var fileUri string
	if err := s.connection.Call(ctx, "poryscript/getfileuri", uri, &fileUri); err != nil {
		return nil, err
	}
	tokens := parse.ParseMiscTokens(content, expression, tokenType, fileUri)
	tokenSet := map[string]parse.MiscToken{}
	for _, t := range tokens {
		tokenSet[t.Name] = t
	}
	s.cachedMiscTokens[uri+expression] = tokenSet
	return tokenSet, nil
}

// Gets the content for the given file uri. The content is cached
// for the file so that parsing is avoided in future calls.
func (s *poryscriptServer) getDocumentContent(ctx context.Context, uri string) (string, error) {
	s.documentsMutex.Lock()
	defer s.documentsMutex.Unlock()

	uri, _ = url.QueryUnescape(uri)
	if content, ok := s.cachedDocuments[uri]; ok {
		return content, nil
	}
	return s.getAndCacheDocumentContent(ctx, uri)
}

// Fetches and caches the content for the given file uri.
func (s *poryscriptServer) getAndCacheDocumentContent(ctx context.Context, uri string) (string, error) {
	uri, _ = url.QueryUnescape(uri)
	var content string
	if err := s.connection.Call(ctx, "poryscript/readfs", uri, &content); err != nil {
		return "", err
	}
	s.cachedDocuments[uri] = content
	return content, nil
}

// Clears the various cached artifacts for the given file uri.
func (s *poryscriptServer) clearCaches(uri string) {
	s.documentsMutex.Lock()
	defer s.documentsMutex.Unlock()
	s.commandsMutex.Lock()
	defer s.commandsMutex.Unlock()
	s.constantsMutex.Lock()
	defer s.constantsMutex.Unlock()
	s.symbolsMutex.Lock()
	defer s.symbolsMutex.Unlock()
	s.miscTokensMutex.Lock()
	defer s.miscTokensMutex.Unlock()
	delete(s.cachedDocuments, uri)
	delete(s.cachedCommands, uri)
	delete(s.cachedConstants, uri)
	delete(s.cachedSymbols, uri)
	delete(s.cachedMiscTokens, uri)
}

// Clears the various cached artifacts for watched files (.inc and .h files).
func (s *poryscriptServer) clearWatchedFileCaches() {
	s.commandsMutex.Lock()
	defer s.commandsMutex.Unlock()
	s.miscTokensMutex.Lock()
	defer s.miscTokensMutex.Unlock()
	s.cachedCommands = map[string]map[string]parse.Command{}
	s.cachedMiscTokens = map[string]map[string]parse.MiscToken{}
}
