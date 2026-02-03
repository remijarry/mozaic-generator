package main

import (
	"fmt"
	"image"
	"image/color"
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

	averagedImages := make([]AveragedImage, 0, len(files))

	for _, file := range files {
		filePath := inputDir + "/" + file.Name()
		img, err := imaging.Open(filePath)

		if err != nil {
			log.Fatalf("failed to open image: %v", err)
			os.Exit(1)
		}

		avgColor := getAverageColor(img)

		averagedImages = append(averagedImages, AveragedImage{
			R:    avgColor.R,
			G:    avgColor.G,
			B:    avgColor.B,
			path: filePath,
			used: false,
		})

		fmt.Printf("Average Color: %+v", avgColor)
	}
}

func getAverageColor(img image.Image) RGB {
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
	return RGB{
		R: int(r / count),
		G: int(g / count),
		B: int(b / count),
	}
}

type AveragedImage struct {
	R int
	G int
	B int

	path string
	used bool
}

type RGB struct {
	R int
	G int
	B int
}
