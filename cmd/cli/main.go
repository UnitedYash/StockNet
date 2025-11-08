package main

import (
	"log"

	"StockNet/internal/cli/tui"
)

func main() {
	if err := tui.StartApp(); err != nil {
		log.Fatal(err)
	}
}
