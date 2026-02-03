package main

import (
	"image"
	"image/color"
	_ "image/jpeg"
	_ "image/png"
	"log"
	"os"

	"github.com/disintegration/imaging"
)

func main() {
	inputDir := "input"
	files, err := os.ReadDir(inputDir)

	if err != nil {
		panic(err)
	}

	averagedImages := make([]TileImage, 0, len(files))

	for _, file := range files {

		filePath := inputDir + "/" + file.Name()

		reader, err := os.Open(filePath)
		if err != nil {
			log.Printf("Error opening file %s: %v", filePath, err)
			continue
		}

		defer reader.Close()

		_, _, configErr := image.DecodeConfig(reader)

		if configErr != nil {
			log.Printf("Skipping %s: not a valid image format or corrupted: %v", filePath, configErr) // More descriptive log
			continue
		}

		_, err = reader.Seek(0, 0)
		if err != nil {
			log.Printf("Error seeking file %s: %v", filePath, err)
		}

		img, err := imaging.Decode(reader)

		if err != nil {
			log.Printf("Error decoding image %s: %v", filePath, err)
		}

		r, g, b := getAverageColor(img)

		averagedImages = append(averagedImages, TileImage{
			R:    r,
			G:    g,
			B:    b,
			path: filePath,
			used: false,
		})
	}
}

// Return the image's average of R,G,B
func getAverageColor(img image.Image) (int, int, int) {
	var r, g, b, count float64

	bounds := img.Bounds()
	width, height := bounds.Max.X, bounds.Max.Y
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			c := img.At(x, y)
			rgba := color.RGBAModel.Convert(c).(color.RGBA)
			r += float64(rgba.R)
			g += float64(rgba.G)
			b += float64(rgba.B)
			count++
		}
	}
	return int(r / count), int(g / count), int(b / count)

}

// todo: rename, could it be merged with RGB struct?
type TileImage struct {
	R int
	G int
	B int

	path string
	used bool
}
