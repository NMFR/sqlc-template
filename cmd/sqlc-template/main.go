package main

import (
	"os"

	"github.com/NMFR/sqlc-template/internal/generator"
)

func main() {
	if err := generator.GenerateFromReader(os.Stdin, os.Stdout); err != nil {
		panic(err)
	}
}
