package main

import (
	"fmt"
	"os"

	"github.com/BenjaminEPH/abyss/ui"
)

func main() {
	if err := ui.Run(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}
