package main

import (
	"fmt"
	"os"

	"github.com/Haruko386/Gogit/cmd"
)

func main() {
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Gogit: %v\n", err)
		os.Exit(1)
	}
}
