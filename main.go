package main

import (
	"fmt"
	"image"
	"log"

	"github.com/disintegration/imaging"
)

const (
	inputDir      = "input"
	mainImagePath = "main-picture/main.jpg"
	gridWidth     = 150
)

func main() {
	// 1. Load tile images, either from cache or by processing the input directory.
	tiles, err := loadOrCreateTiles(inputDir)
	if err != nil {
		log.Fatalf("Fatal error: could not load or create tile data: %v", err)
	}
	fmt.Printf("Successfully loaded %d tile images.\n", len(tiles))

	// 2. Process the main image into a grid of average colors.
	grid, err := generateMosaicGrid(mainImagePath, gridWidth)
	if err != nil {
		log.Fatalf("Fatal error: could not process main image: %v", err)
	}
	fmt.Printf("Successfully generated a %dx%d mosaic grid from the main image.\n", len(grid[0]), len(grid))

	// Next steps:
	// - Match pieces to tiles
	// - Assemble the mosaic
}

// generateMosaicGrid loads the main image and divides it into a grid, calculating the average color for each cell.
func generateMosaicGrid(path string, width int) ([][]RGB, error) {
	// Open the main image. using imaging.Open is a robust way to handle various formats.
	img, err := imaging.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open main image %s: %w", path, err)
	}

	// Calculate the grid dimensions while preserving aspect ratio.
	bounds := img.Bounds()
	imgWidth := bounds.Dx()
	imgHeight := bounds.Dy()
	aspectRatio := float64(imgHeight) / float64(imgWidth)
	height := int(float64(width) * aspectRatio)

	fmt.Printf("Main image dimensions: %dx%d. Mosaic grid dimensions: %dx%d.\n", imgWidth, imgHeight, width, height)

	// Calculate the size of each cell in pixels.
	cellWidth := imgWidth / width
	cellHeight := imgHeight / height

	// Create the 2D slice to store the average color of each grid cell.
	grid := make([][]RGB, height)
	for i := range grid {
		grid[i] = make([]RGB, width)
	}

	// Loop through each cell of the grid.
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			// Define the rectangle for the current cell.
			rect := image.Rect(x*cellWidth, y*cellHeight, (x+1)*cellWidth, (y+1)*cellHeight)

			// Crop the sub-image for the cell and get its average color.
			subImg := imaging.Crop(img, rect)
			grid[y][x] = getAverageColor(subImg)
		}
	}

	return grid, nil
}
