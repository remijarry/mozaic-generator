package main

import (
	"fmt"
	"image"
	"image/color"
	"log"
	"math"
	"os"

	"github.com/disintegration/imaging"
)

const (
	inputDir      = "input"
	mainImagePath = "main-picture/main.jpg"
	gridWidth     = 150
	outputDir     = "output"
	outputFile    = "output/mosaic.jpg"
)

func main() {
	// 1. Load tile images, either from cache or by processing the input directory.
	tiles, err := loadOrCreateTiles(inputDir)
	if err != nil {
		log.Fatalf("Fatal error: could not load or create tile data: %v", err)
	}
	fmt.Printf("Successfully loaded %d tile images.\n", len(tiles))

	// 2. Process the main image into a grid of average colors.
	grid, cellWidth, cellHeight, err := generateMosaicGrid(mainImagePath, gridWidth)
	if err != nil {
		log.Fatalf("Fatal error: could not process main image: %v", err)
	}
	fmt.Printf("Successfully generated a %dx%d mosaic grid from the main image.\n", len(grid[0]), len(grid))

	// 3. Match tile images to the grid cells.
	mosaic := buildMosaic(grid, tiles)
	fmt.Printf("Successfully built a %dx%d mosaic map.\n", len(mosaic[0]), len(mosaic))

	// 4. (OPTIMIZATION) Pre-load and resize all unique tile images into memory.
	preloadedTiles, err := preloadAndResizeTiles(tiles, cellWidth, cellHeight)
	if err != nil {
		log.Fatalf("Fatal error: could not preload tiles: %v", err)
	}
	fmt.Printf("Successfully preloaded and resized %d unique tiles.\n", len(preloadedTiles))


	// 5. Assemble the final mosaic image from the pre-loaded tiles.
	finalImage := assembleMosaic(mosaic, cellWidth, cellHeight, preloadedTiles)

	// 6. Save the final image to a file.
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		log.Fatalf("Fatal error: could not create output directory: %v", err)
	}
	if err := imaging.Save(finalImage, outputFile); err != nil {
		log.Fatalf("Fatal error: could not save final image: %v", err)
	}
	fmt.Printf("Successfully saved final mosaic to %s\n", outputFile)
}

// preloadAndResizeTiles loads all unique tile images and resizes them once.
func preloadAndResizeTiles(tiles []TileImage, cellWidth, cellHeight int) (map[string]image.Image, error) {
	// This map will hold the resized image for each unique tile path.
	preloaded := make(map[string]image.Image)

	for _, tile := range tiles {
		// If we have already loaded this image path, skip it.
		if _, ok := preloaded[tile.Path]; ok {
			continue
		}

		img, err := imaging.Open(tile.Path)
		if err != nil {
			// Return an error to fail fast if a tile image is missing or corrupt.
			return nil, fmt.Errorf("failed to open tile %s: %w", tile.Path, err)
		}
		// Resize the image and store it in the map.
		resized := imaging.Resize(img, cellWidth, cellHeight, imaging.Lanczos)
		preloaded[tile.Path] = resized
	}
	return preloaded, nil
}


// assembleMosaic creates the final image by pasting the chosen tiles onto a canvas.
func assembleMosaic(mosaic [][]TileImage, cellWidth, cellHeight int, preloadedTiles map[string]image.Image) image.Image {
	// Calculate the dimensions of the final image.
	outputWidth := len(mosaic[0]) * cellWidth
	outputHeight := len(mosaic) * cellHeight

	// Create a new blank canvas.
	canvas := imaging.New(outputWidth, outputHeight, color.NRGBA{0, 0, 0, 255})

	fmt.Println("Assembling final mosaic image from pre-loaded tiles...")
	// Loop through the mosaic map and paste each tile.
	for y, row := range mosaic {
		for x, tile := range row {
			if tile.Path == "" {
				continue
			}
			// Perform a fast lookup to get the pre-resized tile.
			resized, ok := preloadedTiles[tile.Path]
			if !ok {
				// This should not happen if preloading was successful, but good to have a check.
				log.Printf("Warning: preloaded tile not found for path %s", tile.Path)
				continue
			}
			
			// Define the point where the tile will be pasted.
			pastePoint := image.Pt(x*cellWidth, y*cellHeight)

			// Paste the resized tile onto the canvas.
			canvas = imaging.Paste(canvas, resized, pastePoint)
		}
	}
	fmt.Println("Assembly complete.")
	return canvas
}

// buildMosaic creates a 2D map of TileImages that corresponds to the grid of the main image.
func buildMosaic(grid [][]RGB, tiles []TileImage) [][]TileImage {
	// Create the 2D slice that will hold the final arrangement of tile images.
	mosaic := make([][]TileImage, len(grid))
	for i := range grid {
		mosaic[i] = make([]TileImage, len(grid[0]))
	}

	usedTilesCount := 0
	totalTiles := len(tiles)

	for y, row := range grid {
		for x, cellColor := range row {
			// If we have used all available tiles, reset the `Used` flag on all of them.
			if usedTilesCount == totalTiles {
				// fmt.Println("Resetting tile usage, all tiles have been used once.") // Commenting out to reduce log spam.
				for i := range tiles {
					tiles[i].Used = false
				}
				usedTilesCount = 0
			}

			// Find the best matching unused tile for the current cell color.
			bestMatchIndex := findBestMatchIndex(cellColor, tiles)
			if bestMatchIndex != -1 {
				// Assign the chosen tile to the mosaic map and mark it as used.
				mosaic[y][x] = tiles[bestMatchIndex]
				tiles[bestMatchIndex].Used = true
				usedTilesCount++
			}
		}
	}
	return mosaic
}

// findBestMatchIndex finds the index of the best-matching, unused tile image for a given target color.
func findBestMatchIndex(targetColor RGB, tiles []TileImage) int {
	bestIndex := -1
	minDist := math.Inf(1)

	for i, tile := range tiles {
		// Skip tiles that have already been used in the current cycle.
		if tile.Used {
			continue
		}

		dist := squaredEuclideanDistance(targetColor, tile.RGB)
		if float64(dist) < minDist {
			minDist = float64(dist)
			bestIndex = i
		}
	}

	return bestIndex
}

// squaredEuclideanDistance calculates the squared distance between two RGB colors.
// We use squared distance as it's faster than taking the square root and is sufficient for comparison.
func squaredEuclideanDistance(c1, c2 RGB) int {
	dr := c1.R - c2.R
	dg := c1.G - c2.G
	db := c1.B - c2.B
	return dr*dr + dg*dg + db*db
}

// generateMosaicGrid loads the main image and divides it into a grid, calculating the average color for each cell.
func generateMosaicGrid(path string, width int) ([][]RGB, int, int, error) {
	// Open the main image. using imaging.Open is a robust way to handle various formats.
	img, err := imaging.Open(path)
	if err != nil {
		return nil, 0, 0, fmt.Errorf("failed to open main image %s: %w", path, err)
	}

	// Calculate the grid dimensions while preserving aspect ratio.
	bounds := img.Bounds()
	imgWidth := bounds.Dx()
	imgHeight := bounds.Dy()
	aspectRatio := float64(imgHeight) / float64(imgWidth)
	height := int(float64(width) * aspectRatio)

	// Calculate the size of each cell in pixels.
	cellWidth := imgWidth / width
	cellHeight := imgHeight / height

	fmt.Printf("Main image dimensions: %dx%d. Mosaic grid dimensions: %dx%d. Cell dimensions: %dx%d.\n", imgWidth, imgHeight, width, height, cellWidth, cellHeight)

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

	return grid, cellWidth, cellHeight, nil
}
