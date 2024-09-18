package main

import (
	"archive/zip"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
)

func main() {
	inputFile := flag.String("input", "", "Input ZIP file (required)")
	outputDir := flag.String("output", ".", "Output directory (default: current directory)")
	flag.Parse()

	if *inputFile == "" {
		fmt.Println("Error: Input ZIP file is required")
		fmt.Println("Usage: linux-unzip-cp932 -input <zip_file> [-output <output_directory>]")
		flag.PrintDefaults()
		os.Exit(1)
	}

	err := extractZip(*inputFile, *outputDir)
	if err != nil {
		fmt.Printf("Error extracting ZIP file: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Extraction completed successfully.")
}

func extractZip(zipPath, destPath string) error {
	reader, err := zip.OpenReader(zipPath)
	if err != nil {
		return fmt.Errorf("failed to open zip: %v", err)
	}
	defer reader.Close()

	err = os.MkdirAll(destPath, 0755)
	if err != nil {
		return fmt.Errorf("failed to create output directory: %v", err)
	}

	for _, file := range reader.File {
		decodedFileName, err := decodeCP932(file.Name)
		if err != nil {
			return fmt.Errorf("failed to decode filename '%s': %v", file.Name, err)
		}

		path := filepath.Join(destPath, decodedFileName)

		if file.FileInfo().IsDir() {
			os.MkdirAll(path, file.Mode())
			continue
		}

		fileReader, err := file.Open()
		if err != nil {
			return fmt.Errorf("failed to open file '%s': %v", decodedFileName, err)
		}
		defer fileReader.Close()

		targetFile, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
		if err != nil {
			return fmt.Errorf("failed to create file '%s': %v", path, err)
		}
		defer targetFile.Close()

		_, err = io.Copy(targetFile, fileReader)
		if err != nil {
			return fmt.Errorf("failed to write file '%s': %v", path, err)
		}
	}

	return nil
}

func decodeCP932(s string) (string, error) {
	decoder := japanese.ShiftJIS.NewDecoder()
	decodedBytes, _, err := transform.Bytes(decoder, []byte(s))
	if err != nil {
		return "", err
	}
	return string(decodedBytes), nil
}
