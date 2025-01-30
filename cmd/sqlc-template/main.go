package main

import (
	"os"

	"github.com/NMFR/sqlc-template/internal/code"
)

func main() {
	if err := code.GenerateFromReader(os.Stdin, os.Stdout); err != nil {
		panic(err)
	}
}
