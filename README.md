# poryscript-pls

[![Actions Status](https://github.com/huderlem/poryscript-pls/workflows/Go/badge.svg)](https://github.com/huderlem/poryscript-pls/actions) [![codecov](https://codecov.io/gh/huderlem/poryscript-pls/branch/master/graph/badge.svg)](https://codecov.io/gh/huderlem/poryscript-pls)

The language server for [Poryscript](https://github.com/huderlem/poryscript).

## Local Development Setup

First, install [Go](https://go.dev/doc/install).

Then, build it!
```
go build
```

## Testing with the Poryscript VS Code Extension

Clone the Poryscript Language Extension repository.
```
git clone https://github.com/SBird1337/poryscript-language
```

In `client/src/extension.ts`, replace the executable path with hardcoded paths to your Poryscript language server binary (`poryscript-pls`).
```ts
const debugPlsPath = "your\\path\\to\\poryscript-pls.exe";
const releasePlsPath = "your\\path\\to\\poryscript-pls.exe";
```

Launch the extension as usual (e.g. pressing `F5`) to test the `poryscript-pls` server.  **Windows Note**: It doesn't seem to load properly if the project you load in the `Extension Development Host` is located in the WSL filesystem, so make sure you're testing in a normal Windows environment.

## Notes

This project's `lsp/` directory is a modified vendor copy of Sourcegraph's LSP bindings for Go. It provides all of the Go structs for the [LSP specification](https://microsoft.github.io/language-server-protocol/specifications/specification-current/). Sourcegraph stopped maintaining their library for some reason, but it still seems to be the best library for it.

It also makes use of Sourcegraph's `jsonrpc2` library to facilitate the JSON RPC communications.
