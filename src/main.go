package main

import (
	"fmt"
	"image"
	"image/color"
	"log"
	"math/rand"
	"os"
	"sort" // New import

	"github.com/disintegration/imaging"
	"github.com/remay/mozaic-generator/pkg/imageutils"
)

const (
	inputDir      = "/home/tgrx/programming/go/mozaic-generator/input"
	mainImagePath = "/home/tgrx/programming/go/mozaic-generator/main-picture/main.jpg"
	gridWidth     = 60
	outputDir     = "/home/tgrx/programming/go/mozaic-generator/output"
	outputFile    = "/home/tgrx/programming/go/mozaic-generator/output/mosaic.jpg"

	topNMatches  = 5 // Define a constant for the number of top matches to consider
	MaxTileUsage = 5
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

	// Create a struct for example type PerlinNoise { noise, flattenedIndex }
	// Allocate PerlinNoise array with its final size (width x height)
	// Iterate over grid, compute a flattened index (2D to 1D) idx = i + j * width
	// Compute perlin noise and add it at perlinNoise[idx]
	// Sort perlin noise
	// var perlinNoise []float32

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
func preloadAndResizeTiles(tiles []imageutils.TileImage, cellWidth, cellHeight int) (map[string]image.Image, error) {
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
func assembleMosaic(mosaic [][]imageutils.TileImage, cellWidth, cellHeight int, preloadedTiles map[string]image.Image) image.Image {
	// Calculate the dimensions of the final image.
	outputWidth := len(mosaic[0]) * cellWidth
	outputHeight := len(mosaic) * cellHeight

	// Create a new blank canvas.
	canvas := imaging.New(outputWidth, outputHeight, color.NRGBA{0, 0, 0, 255})

	fmt.Println("Assembling final mosaic image from pre-loaded tiles...")
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

			pastePoint := image.Pt(x*cellWidth, y*cellHeight)

			canvas = imaging.Paste(canvas, resized, pastePoint)
		}
	}
	fmt.Println("Assembly complete.")
	return canvas
}

// buildMosaic creates a 2D map of TileImages that corresponds to the grid of the main image.
func buildMosaic(grid [][]imageutils.RGB, tiles []imageutils.TileImage) [][]imageutils.TileImage {
	// Create the 2D slice that will hold the final arrangement of tile images.
	mosaic := make([][]imageutils.TileImage, len(grid))
	for i := range grid {
		mosaic[i] = make([]imageutils.TileImage, len(grid[0]))
	}

	usedTilesCount := 0
	totalTiles := len(tiles)

	for y, row := range grid {
		for x, cellColor := range row {
			// If we have used all available tiles, reset the `UsageCount` flag on all of them.
			if usedTilesCount == totalTiles {
				fmt.Println("Resetting tile usage, all tiles have been used once.") // Commenting out to reduce log spam.
				for i := range tiles {
					tiles[i].UsageCount = 0
				}
				usedTilesCount = 0
			}

			// Find the best matching tiles for the current cell color.
			bestMatchIndices := findBestMatchIndex(cellColor, tiles)
			if len(bestMatchIndices) > 0 {
				// Randomly select one tile from the best matches.
				randomIndex := rand.Intn(len(bestMatchIndices))
				chosenTileIndex := bestMatchIndices[randomIndex]
				mosaic[y][x] = tiles[chosenTileIndex]
			}
		}
	}
	return mosaic
}

// findBestMatchIndex finds the indices of the best-matching tiles for a given target color.
// It returns up to `topNMatches` indices of the closest tiles.
func findBestMatchIndex(targetColor imageutils.RGB, tiles []imageutils.TileImage) []int {
	type tileDistance struct {
		Distance float64
		Index    int
	}

	var distances []tileDistance
	for i, tile := range tiles {
		dist := imageutils.SquaredEuclideanDistance(targetColor, tile.RGB)
		distances = append(distances, tileDistance{Distance: float64(dist), Index: i})
	}

	// Sort the distances in ascending order.
	sort.Slice(distances, func(i, j int) bool {
		return distances[i].Distance < distances[j].Distance
	})

	// Collect the indices of the top N matches.
	var bestMatchIndices []int
	for i := 0; i < len(distances) && i < topNMatches; i++ {
		distanceIndex := distances[i].Index
		if tiles[distanceIndex].UsageCount > MaxTileUsage {
			continue
		}

		bestMatchIndices = append(bestMatchIndices, distances[i].Index)
	}

	return bestMatchIndices
}

// generateMosaicGrid loads the main image and divides it into a grid, calculating the average color for each cell.
func generateMosaicGrid(path string, width int) ([][]imageutils.RGB, int, int, error) {
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
	grid := make([][]imageutils.RGB, height)
	for i := range grid {
		grid[i] = make([]imageutils.RGB, width)
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
