package main

import (
	"os"

	"github.com/mberwanger/dockerfiles/tool/cmd"
)

func main() {
	cmd.Execute(os.Args[1:])
}
