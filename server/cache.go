package server

import (
	"context"
	"net/url"

	"github.com/huderlem/poryscript-pls/parse"
)

// Gets the aggregate list of Commands from the collection of files that define
// the Commands. The Commands are cached for the given file so that parsing is avoided
// in future calls.
func (s *poryscriptServer) getCommands(ctx context.Context, file string) ([]parse.Command, error) {
	settings, err := s.config.GetFileSettings(ctx, s.connection, file)
	if err != nil {
		return []parse.Command{}, err
	}
	// Aggregate a slice of Commands from all of the relevant files.
	commands := []parse.Command{}
	for _, includeFilepath := range settings.CommandIncludes {
		fileCommands, err := s.getCommandsInFile(ctx, includeFilepath)
		if err != nil {
			// TODO: log error? we don't want to fail if a single file resulted in an error.
			continue
		}
		commands = append(commands, fileCommands...)
	}
	commands = append(commands, parse.KeywordCommands...)
	return commands, nil
}

// Gets the list of Commands from the given file. The Commands
// are cached for the given file so that parsing is avoided in future
// calls.
func (s *poryscriptServer) getCommandsInFile(ctx context.Context, file string) ([]parse.Command, error) {
	file, _ = url.QueryUnescape(file)
	if commands, ok := s.cachedCommands[file]; ok {
		return commands, nil
	}
	return s.getAndCacheCommandsInFile(ctx, file)
}

// Fetches and caches the Commands from the given file.
func (s *poryscriptServer) getAndCacheCommandsInFile(ctx context.Context, file string) ([]parse.Command, error) {
	file, _ = url.QueryUnescape(file)
	var content string
	if err := s.connection.Call(ctx, "poryscript/readfile", file, &content); err != nil {
		return []parse.Command{}, err
	}
	if !s.config.HasWorkspaceFolderCapability {
		return []parse.Command{}, nil
	}
	commands := parse.ParseCommands(content)
	s.cachedCommands[file] = commands
	return commands, nil
}

// Gets the list of poryscript constants from the given file. The constants
// are cached for the given file so that parsing is avoided in future calls.
func (s *poryscriptServer) getConstantsInFile(ctx context.Context, file string) ([]parse.ConstantSymbol, error) {
	file, _ = url.QueryUnescape(file)
	if constants, ok := s.cachedConstants[file]; ok {
		return constants, nil
	}
	return s.getAndCacheConstantsInFile(ctx, file)
}

// Fetches and caches the poryscript constants from the given file.
func (s *poryscriptServer) getAndCacheConstantsInFile(ctx context.Context, file string) ([]parse.ConstantSymbol, error) {
	file, _ = url.QueryUnescape(file)
	content, err := s.getDocumentContent(ctx, file)
	if err != nil {
		return []parse.ConstantSymbol{}, err
	}
	constants := parse.ParseConstants(content, file)
	s.cachedConstants[file] = constants
	return constants, nil
}

// Gets the list of poryscript symbols from the given file. The symbols
// are cached for the given file so that parsing is avoided in future calls.
func (s *poryscriptServer) getSymbolsInFile(ctx context.Context, file string) ([]parse.Symbol, error) {
	file, _ = url.QueryUnescape(file)
	if symbols, ok := s.cachedSymbols[file]; ok {
		return symbols, nil
	}
	return s.getAndCacheSymbolsInFile(ctx, file)
}

// Fetches and caches the poryscript symbols from the given file.
func (s *poryscriptServer) getAndCacheSymbolsInFile(ctx context.Context, file string) ([]parse.Symbol, error) {
	file, _ = url.QueryUnescape(file)
	content, err := s.getDocumentContent(ctx, file)
	if err != nil {
		return []parse.Symbol{}, err
	}
	symbols := parse.ParseSymbols(content, file)
	s.cachedSymbols[file] = symbols
	return symbols, nil
}

// Gets the aggregate list of miscellaneous tokens from the collection of files
// specified in the settings.
func (s *poryscriptServer) getMiscTokens(ctx context.Context, file string) ([]parse.MiscToken, error) {
	file, _ = url.QueryUnescape(file)
	settings, err := s.config.GetFileSettings(ctx, s.connection, file)
	if err != nil {
		return []parse.MiscToken{}, err
	}
	miscTokens := []parse.MiscToken{}
	for _, includeSetting := range settings.SymbolIncludes {
		tokens, err := s.getMiscTokensInFile(ctx, includeSetting.Expression, includeSetting.Type, includeSetting.File)
		if err != nil {
			// TODO: log error?
			continue
		}
		miscTokens = append(miscTokens, tokens...)
	}
	return miscTokens, nil
}

// Gets the list of miscellaneous tokens from the given file. The tokens
// are cached for the given file so that parsing is avoided in future calls.
func (s *poryscriptServer) getMiscTokensInFile(ctx context.Context, expression, tokenType, file string) ([]parse.MiscToken, error) {
	file, _ = url.QueryUnescape(file)
	if tokens, ok := s.cachedMiscTokens[file+expression]; ok {
		return tokens, nil
	}
	return s.getAndCacheMiscTokensInFile(ctx, expression, tokenType, file)
}

// Fetches and caches the miscellaneous tokens from the given file.
func (s *poryscriptServer) getAndCacheMiscTokensInFile(ctx context.Context, expression, tokenType, file string) ([]parse.MiscToken, error) {
	file, _ = url.QueryUnescape(file)
	var content string
	if err := s.connection.Call(ctx, "poryscript/readfile", file, &content); err != nil {
		return []parse.MiscToken{}, err
	}
	var fileUri string
	if err := s.connection.Call(ctx, "poryscript/getfileuri", file, &fileUri); err != nil {
		return []parse.MiscToken{}, err
	}
	tokens := parse.ParseMiscTokens(content, expression, tokenType, fileUri)
	s.cachedMiscTokens[file+expression] = tokens
	return tokens, nil
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
