package main

import (
	"encoding/json"
	"image"
	"image/color"
	_ "image/jpeg"
	_ "image/png"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/disintegration/imaging"
)

type TileImage struct {
	R int
	G int
	B int

	path string
	used bool
}

type Cache struct {
	Signature string      `json:"signature"`
	Tiles     []TileImage `json:"tiles"`
}

func main() {
	inputDir := "input"
	files, err := os.ReadDir(inputDir)

	if err != nil {
		panic(err)
	}

	signature := GenerateFolderSignature(files)

	cacheDir := "cache"
	if _, err := os.Stat(cacheDir); os.IsNotExist(err) {
		os.Mkdir(cacheDir, 0755)
	}

	cacheFileName := "cache.json"
	cacheFilePath := filepath.Join(cacheDir, cacheFileName)

	var cache Cache

	cacheFileContent, err := os.ReadFile(cacheFilePath)
	cacheValid := false

	if err == nil { // File read successfully
		jsonErr := json.Unmarshal(cacheFileContent, &cache)
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

func GenerateFolderSignature(files []os.DirEntry) string {
	fileNames := []string{}
	for _, file := range files {
		fileNames = append(fileNames, file.Name())
	}

	sort.Strings(fileNames)
	return strings.Join(fileNames, "")
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
