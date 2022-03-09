package server

import "os"

// StdIoConn implements the ReadWriteCloser interface to facilitate
// JSON RPC communication over stdout and stdin channels.
// Implementation borrowed from https://github.com/tawasprache/kompilierer
type StdIoConn struct{}

func (StdIoConn) Read(p []byte) (int, error) {
	return os.Stdin.Read(p)
}

func (StdIoConn) Write(p []byte) (int, error) {
	return os.Stdout.Write(p)
}

func (StdIoConn) Close() error {
	if err := os.Stdin.Close(); err != nil {
		return err
	}
	return os.Stdout.Close()
}
