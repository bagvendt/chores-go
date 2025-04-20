package utils

import (
	"log"
	"os"
	"path/filepath"
)

// GetImageFiles reads the static/img directory and returns a list of image filenames.
func GetImageFiles() ([]string, error) {
	var files []string
	imgDir := "./static/img"
	items, err := os.ReadDir(imgDir)
	if err != nil {
		log.Printf("Error reading image directory %s: %v", imgDir, err)
		return nil, err
	}

	for _, item := range items {
		if !item.IsDir() {
			ext := filepath.Ext(item.Name())
			if ext == ".png" || ext == ".jpg" || ext == ".jpeg" || ext == ".gif" || ext == ".svg" {
				files = append(files, item.Name())
			}
		}
	}
	return files, nil
}
