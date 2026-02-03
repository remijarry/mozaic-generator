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

	for _, file := range files {
		img, err := imaging.Open(inputDir + "/" + file.Name())

		if err != nil {
			log.Fatalf("failed to open image: %v", err)
			os.Exit(1)
		}

		avgColor := getAverageColor(img)
		fmt.Printf("Average Color: %+v", avgColor)
	}
}

func getAverageColor(img image.Image) color.Color {
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
	return color.RGBA{
		R: uint8(r / count),
		G: uint8(g / count),
		B: uint8(b / count),
		A: 255,
	}
}
