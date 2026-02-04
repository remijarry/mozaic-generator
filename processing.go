package main

import (
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"log"
	"os"
	"path/filepath"

	"github.com/disintegration/imaging"
)

// RGB represents a simple RGB color.
type RGB struct {
	R int `json:"r"`
	G int `json:"g"`
	B int `json:"b"`
}

// TileImage holds the data for a single tile, including its average color and file path.
type TileImage struct {
	RGB  `json:"rgb"`
	Path string `json:"path"`
	Used bool   `json:"used"`
}

// processImages reads all valid images from a directory and calculates their average color.
func processImages(inputDir string) ([]TileImage, error) {
	files, err := os.ReadDir(inputDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read input directory %s: %w", inputDir, err)
	}

	var tiles []TileImage
	for _, file := range files {
		filePath := filepath.Join(inputDir, file.Name())

		reader, err := os.Open(filePath)
		if err != nil {
			log.Printf("Could not open %s, skipping: %v", filePath, err)
			continue
		}
		defer reader.Close()

		// Check if it's a valid image format without decoding the whole file.
		if _, _, err := image.DecodeConfig(reader); err != nil {
			// Not a valid image, or not a format we have a decoder for.
			continue
		}

		// Rewind the file to decode the full image.
		if _, err := reader.Seek(0, 0); err != nil {
			log.Printf("Could not seek in %s, skipping: %v", filePath, err)
			continue
		}

		img, err := imaging.Decode(reader)
		if err != nil {
			log.Printf("Could not decode image %s, skipping: %v", filePath, err)
			continue
		}

		tiles = append(tiles, TileImage{
			RGB:  getAverageColor(img),
			Path: filePath,
			Used: false,
		})
	}
	return tiles, nil
}

// getAverageColor calculates the average RGB values for a given image.
func getAverageColor(img image.Image) RGB {
	bounds := img.Bounds()
	var r, g, b, count uint64
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			pr, pg, pb, _ := img.At(x, y).RGBA()
			r += uint64(pr >> 8)
			g += uint64(pg >> 8)
			b += uint64(pb >> 8)
			count++
		}
	}
	return RGB{
		R: int(r / count),
		G: int(g / count),
		B: int(b / count),
	}
}
