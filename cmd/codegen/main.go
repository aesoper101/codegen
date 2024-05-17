package main

import (
	"fmt"
	"github.com/aesoper101/codegen/internal/commands"
	"os"
)

func main() {
	root := commands.NewCodegenCommand()
	if err := root.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
