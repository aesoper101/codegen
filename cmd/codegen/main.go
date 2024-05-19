package main

import (
	"fmt"
	"os"

	"github.com/aesoper101/codegen/internal/commands"
)

func main() {
	root := commands.NewCodegenCommand()
	if err := root.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
