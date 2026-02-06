package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/remay/mozaic-generator/pkg/imageutils"
)

// Cache is the top-level struct for our cache file.
type Cache struct {
	Signature string             `json:"signature"`
	Tiles     []imageutils.TileImage `json:"tiles"`
}

// loadOrCreateTiles orchestrates loading tiles from cache or generating them if the cache is invalid.
func loadOrCreateTiles(inputDir string) ([]imageutils.TileImage, error) {
	cacheDir := "cache"
	cacheFileName := "cache.json"
	cacheFilePath := filepath.Join(cacheDir, cacheFileName)

	// Ensure cache directory exists.
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create cache directory %s: %w", cacheDir, err)
	}

	// Generate a signature for the current state of the input directory.
	currentSignature, err := GenerateFolderSignature(inputDir)
	if err != nil {
		return nil, fmt.Errorf("failed to generate folder signature: %w", err)
	}

	// Try to read the cache file.
	cacheFileContent, err := os.ReadFile(cacheFilePath)
	if err == nil {
		// If file exists, try to use it.
		var cache Cache
		if json.Unmarshal(cacheFileContent, &cache) == nil {
			if cache.Signature == currentSignature {
				log.Printf("Cache is valid. Loading %d tiles from %s.", len(cache.Tiles), cacheFilePath)
				return cache.Tiles, nil
			}
			log.Println("Cache signature mismatch. Rebuilding cache.")
		} else {
			log.Println("Error unmarshaling cache file. Rebuilding cache.")
		}
	} else if !os.IsNotExist(err) {
		// If the error is anything other than "not found", log it but continue.
		log.Printf("Error reading cache file %s: %v. Rebuilding cache.", cacheFilePath, err)
	} else {
		log.Printf("Cache file not found. Rebuilding cache.")
	}

	// If cache was not valid, generate tiles from scratch.
	log.Println("Processing images from input directory...")
	tiles, err := processImages(inputDir)
	if err != nil {
		return nil, fmt.Errorf("failed to process images: %w", err)
	}

	// Save the newly generated tiles to the cache.
	newCache := Cache{
		Signature: currentSignature,
		Tiles:     tiles,
	}
	jsonBytes, err := json.MarshalIndent(newCache, "", "  ")
	if err != nil {
		log.Printf("Error marshaling cache to JSON: %v", err)
	} else if err := os.WriteFile(cacheFilePath, jsonBytes, 0644); err != nil {
		log.Printf("Error writing cache file %s: %v", cacheFilePath, err)
	} else {
		log.Printf("Cache successfully written to %s", cacheFilePath)
	}

	return tiles, nil
}

// GenerateFolderSignature creates a unique signature based on file names and modification times.
func GenerateFolderSignature(dirPath string) (string, error) {
	files, err := os.ReadDir(dirPath)
	if err != nil {
		return "", err
	}

	fileSignatures := make([]string, 0, len(files))
	for _, file := range files {
		if !file.IsDir() {
			info, err := file.Info()
			if err != nil {
				return "", fmt.Errorf("failed to get file info for %s: %w", file.Name(), err)
			}
			fileSignatures = append(fileSignatures, fmt.Sprintf("%s:%d", file.Name(), info.ModTime().UnixNano()))
		}
	}

	sort.Strings(fileSignatures)
	return strings.Join(fileSignatures, ","), nil
}
