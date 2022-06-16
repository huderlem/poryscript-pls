package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/huderlem/poryscript-pls/server"
)

const version = "1.0.4"

func parseOptions() {
	helpPtr := flag.Bool("h", false, "show poryscript-pls help information")
	versionPtr := flag.Bool("v", false, "show version of poryscript-pls")
	flag.Parse()

	if *helpPtr {
		flag.Usage()
		os.Exit(0)
	}

	if *versionPtr {
		fmt.Printf("%s\n", version)
		os.Exit(0)
	}
}

func main() {
	parseOptions()

	s := server.New()
	s.Run()
}
