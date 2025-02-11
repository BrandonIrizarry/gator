package main

import (
	"fmt"
	"github.com/BrandonIrizarry/gator/internal/config"
	"os"
)

func main() {
	cfg, err := config.Read()

	if err != nil {
		fmt.Fprintf(os.Stderr, "%v", err)
	}

	cfg.SetUser("bci")

	cfg, err = config.Read()

	fmt.Printf("%v", cfg)
}
