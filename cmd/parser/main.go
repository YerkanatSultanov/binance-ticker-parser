package main

import (
	"binance-ticker-parser/internal/applicator"
	"log"
)

func main() {
	if err := app.Run(); err != nil {
		log.Fatalf("error: %v", err)
	}
}
