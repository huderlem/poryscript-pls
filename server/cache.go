package server

import (
	"context"

	"github.com/huderlem/poryscript-pls/parse"
)

// Gets the aggregate list of Commands from the collection of files that define
// the Commands. The Commands are cached for given file so that parsing is avoided
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
// are cached for given file so that parsing is avoided in future
// calls.
func (s *poryscriptServer) getCommandsInFile(ctx context.Context, file string) ([]parse.Command, error) {
	if commands, ok := s.cachedCommands[file]; ok {
		return commands, nil
	}
	return s.getAndCacheCommandsInFile(ctx, file)
}

// Fetches and caches the Commands from the given file.
func (s *poryscriptServer) getAndCacheCommandsInFile(ctx context.Context, file string) ([]parse.Command, error) {
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
// are cached for given file so that parsing is avoided in future calls.
func (s *poryscriptServer) getConstantsInFile(ctx context.Context, file string) ([]parse.ConstantSymbol, error) {
	if constants, ok := s.cachedConstants[file]; ok {
		return constants, nil
	}
	return s.getAndCacheConstantsInFile(ctx, file)
}

// Fetches and caches the poryscript constants from the given file.
func (s *poryscriptServer) getAndCacheConstantsInFile(ctx context.Context, file string) ([]parse.ConstantSymbol, error) {
	var content string
	if err := s.connection.Call(ctx, "poryscript/readfs", file, &content); err != nil {
		return []parse.ConstantSymbol{}, err
	}
	constants := parse.ParseConstants(content)
	s.cachedConstants[file] = constants
	return constants, nil
}
