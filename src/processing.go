package main

import (
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/remay/mozaic-generator/pkg/imageutils"

	"github.com/disintegration/imaging"
)

// processImages reads all valid images from a directory and calculates their average color.
func processImages(inputDir string) ([]imageutils.TileImage, error) {
	files, err := os.ReadDir(inputDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read input directory %s: %w", inputDir, err)
	}

	var tiles []imageutils.TileImage
	for _, file := range files {
		filePath := filepath.Join(inputDir, file.Name())

		reader, err := os.Open(filePath)
		if err != nil {
			log.Printf("Could not open %s, skipping: %v", filePath, err)
			continue
		}
		defer reader.Close()

		// Check if it's a valid image format without decoding the whole file.
		if !isValidFormat(reader) {
			// Not a valid image, or not a format we have a decoder for.
			continue
		}

		// Rewind the file to decode the full image.
		if err := rewindFilePointer(reader); err != nil {
			log.Printf("Could not seek in %s, skipping: %v", filePath, err)
			continue
		}

		img, err := imaging.Decode(reader)
		if err != nil {
			log.Printf("Could not decode image %s, skipping: %v", filePath, err)
			continue
		}

		tiles = append(tiles, imageutils.TileImage{
			RGB:  getAverageColor(img),
			Path: filePath,
		})
	}
	return tiles, nil
}

// getAverageColor calculates the average RGB values for a given image.
func getAverageColor(img image.Image) imageutils.RGB {
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
	return imageutils.RGB{
		R: int(r / count),
		G: int(g / count),
		B: int(b / count),
	}
}

// isValidFormat checks if a reader contains a valid, decodable image.
func isValidFormat(reader io.Reader) bool {
	_, _, err := image.DecodeConfig(reader)
	return err == nil
}

// rewindFilePointer moves the seeker's offset to the beginning of the file.
func rewindFilePointer(seeker io.Seeker) error {
	_, err := seeker.Seek(0, 0)
	return err
}
