package server

import (
	"context"

	"github.com/huderlem/poryscript-pls/parse"
)

// Gets the list of Commands from the given file. The Commands
// are cached for given file so that parsing is avoided in future
// calls.
func (s *poryscriptServer) getCommands(ctx context.Context, file string) ([]parse.Command, error) {
	if commands, ok := s.cachedCommands[file]; ok {
		return commands, nil
	}
	return s.getAndCacheCommands(ctx, file)
}

// Fetches and caches the Commands from the given file.
func (s *poryscriptServer) getAndCacheCommands(ctx context.Context, file string) ([]parse.Command, error) {
	var content string
	if err := s.connection.Call(ctx, "poryscript/readfile", file, &content); err != nil {
		return []parse.Command{}, err
	}
	if !s.config.HasWorkspaceFolderCapability {
		return []parse.Command{}, nil
	}
	return parse.ParseCommands(content), nil
}
