package main

import (
	"github.com/huderlem/poryscript-pls/server"
)

func main() {
	lspServer := server.New()
	lspServer.Run()
}
