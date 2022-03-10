package server

import "os"

// StdioRWC implements the ReadWriteCloser interface to facilitate
// JSON RPC communication over stdout and stdin channels.
// Implementation borrowed from https://github.com/tawasprache/kompilierer
type StdioRWC struct{}

func (StdioRWC) Read(p []byte) (int, error) {
	return os.Stdin.Read(p)
}

func (StdioRWC) Write(p []byte) (int, error) {
	return os.Stdout.Write(p)
}

func (StdioRWC) Close() error {
	if err := os.Stdin.Close(); err != nil {
		return err
	}
	return os.Stdout.Close()
}
