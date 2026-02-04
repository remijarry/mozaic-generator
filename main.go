package main

import (
	"fmt"
	"log"
)

func main() {
	inputDir := "input"

	// 1. Load tile images, either from cache or by processing the input directory.
	tiles, err := loadOrCreateTiles(inputDir)
	if err != nil {
		log.Fatalf("Fatal error: could not load or create tile data: %v", err)
	}

	fmt.Printf("Successfully loaded %d tile images.\n", len(tiles))

	// Next steps will go here:
	// - Load the main image
	// - Divide it into pieces
	// - Match pieces to tiles
	// - Assemble the mosaic
}