package main

import (
	"fmt"
	"os"

	"github.com/a2gx/weather/internal/infra/config"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "config load error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("%+v\n", cfg)

	fmt.Println("Hello World!")
}
