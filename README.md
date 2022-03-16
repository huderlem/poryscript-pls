# poryscript-pls

The language server for [Poryscript](https://github.com/huderlem/poryscript). **This is in very early development.**

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

In `client/src/extension.ts`, replace the `serverOptions` with a direct call to the Poryscript language server binary (`poryscript-pls`).
```ts
const serverOptions: ServerOptions = async () => {
    const binPath = "path\\to\\poryscript-pls.exe";
    if (!binPath) {
        throw new Error("Couldn't fetch poryscript-pls binary");
    }
    return child_process.spawn(binPath);
};
```

Also in `client/src/extension.ts`, replace the `poryscript/getfileuri` handler with the following:
```ts
client.onRequest("poryscript/getfileuri", file => {
    return pathToFileURL(path.join(workspace.workspaceFolders[0].uri.fsPath, file)).toString();
});
```

This will cause the VS Code extension client to talk to the `poryscript-pls` server via `stdout` and `stdin`, rather than communicating via IPC with the Node server.

Launch the extension as usual to test the `poryscript-pls` server.

## Notes

This project's `lsp/` directory is a modified vendor copy of Sourcegraph's LSP bindings for Go. It provides all of the Go structs for the [LSP specification](https://microsoft.github.io/language-server-protocol/specifications/specification-current/). Sourcegraph stopped maintaining their library for some reason, but it still seems to be the best library for it.

It also makes use of Sourcegraph's `jsonrpc2` library to facilitate the JSON RPC communications.
