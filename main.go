package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"

	"github.com/yeka/zip"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
)

func main() {
	inputFile := flag.String("input", "", "Input ZIP file (required)")
	outputDir := flag.String("output", ".", "Output directory (default: current directory)")
	password := flag.String("password", "", "Password for encrypted ZIP (optional)")
	showVersion := flag.Bool("version", false, "Print version information and exit")
	flag.Parse()

	if *showVersion {
		printVersion()
		os.Exit(0)
	}

	if *inputFile == "" {
		fmt.Println("Error: Input ZIP file is required")
		fmt.Println("Usage: linux-unzip-cp932 -input <zip_file> [-output <output_directory>] [-password <zip_password>]")
		flag.PrintDefaults()
		os.Exit(1)
	}

	err := extractZip(*inputFile, *outputDir, *password)
	if err != nil {
		fmt.Printf("Error extracting ZIP file: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Extraction completed successfully.")
}

func extractZip(zipPath, destPath, password string) error {
	// Open the ZIP file
	reader, err := zip.OpenReader(zipPath)
	if err != nil {
		return fmt.Errorf("failed to open zip: %v", err)
	}
	defer reader.Close()

	// Create output directory
	if err := os.MkdirAll(destPath, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %v", err)
	}

	// Print start information
	printStartInfo(zipPath, destPath, len(reader.File))

	// Process each file
	for i, file := range reader.File {
		// Decode filename
		decodedFileName := getDecodedFileName(file, i, len(reader.File))

		// Check path safety
		if !isPathSafe(decodedFileName) {
			fmt.Printf("Warning: Skipping potentially unsafe path: %s\n", decodedFileName)
			continue
		}

		// Extract file
		if err := extractFile(file, destPath, password, decodedFileName); err != nil {
			return err
		}
	}

	return nil
}

// Print information at the start of extraction
func printStartInfo(zipPath, destPath string, totalFiles int) {
	fmt.Printf("Opening ZIP file: %s\n", zipPath)
	fmt.Printf("Extracting to: %s\n", destPath)
	fmt.Printf("Total files: %d\n\n", totalFiles)
}

// Check if the path is safe (not absolute or traversing parent directories)
func isPathSafe(path string) bool {
	cleanPath := filepath.Clean(path)
	return !strings.HasPrefix(cleanPath, "../") && !strings.HasPrefix(cleanPath, "/")
}

// Get decoded filename based on encoding detection
func getDecodedFileName(file *zip.File, index, total int) string {
	// Check GPB 11 flag (UTF-8 encoding flag)
	gpb11Set := (file.Flags & 0x0800) != 0

	if gpb11Set {
		// If GPB 11 is set, filename should already be UTF-8 encoded
		fmt.Printf("[%d/%d] GPB 11 is SET - Using filename as UTF-8: %s\n",
			index+1, total, file.Name)
		return file.Name
	}

	// If GPB 11 is not set, check if valid as UTF-8
	if utf8.ValidString(file.Name) && !containsJapaneseChars(file.Name) {
		// ASCII-only, use as-is
		fmt.Printf("[%d/%d] Using filename as-is (ASCII): %s\n",
			index+1, total, file.Name)
		return file.Name
	}

	// Try decoding from CP932
	decoded, err := decodeCP932(file.Name)
	if err != nil {
		fmt.Printf("[%d/%d] Warning: CP932 decoding failed for '%s', using as-is\n",
			index+1, total, file.Name)
		return file.Name
	}

	fmt.Printf("[%d/%d] Decoded from CP932: %s -> %s\n",
		index+1, total, file.Name, decoded)
	return decoded
}

// Extract a single file from the ZIP archive
func extractFile(file *zip.File, destPath, password string, decodedFileName string) error {
	path := filepath.Join(destPath, decodedFileName)

	// Handle directory
	if file.FileInfo().IsDir() {
		fmt.Printf("Creating directory: %s\n", path)
		return os.MkdirAll(path, file.Mode())
	}

	// Ensure parent directory exists
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("failed to create parent directory for '%s': %v", path, err)
	}

	// Open the file from ZIP
	fileReader, err := openFileWithPassword(file, decodedFileName, password)
	if err != nil {
		return err
	}
	defer fileReader.Close()

	// Create target file
	fmt.Printf("Creating file: %s\n", path)
	targetFile, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
	if err != nil {
		return fmt.Errorf("failed to create file '%s': %v", path, err)
	}
	defer targetFile.Close()

	// Copy file contents
	if _, err = io.Copy(targetFile, fileReader); err != nil {
		return fmt.Errorf("failed to write file '%s': %v", path, err)
	}

	fmt.Printf("Successfully extracted: %s\n\n", path)
	return nil
}

// Open a file from ZIP with password if needed
func openFileWithPassword(file *zip.File, decodedFileName, password string) (io.ReadCloser, error) {
	// Encryption is detected up front from the entry header, not by opening
	// and inspecting the error. Covers both traditional ZipCrypto and WinZip AES.
	if file.IsEncrypted() {
		if password == "" {
			return nil, fmt.Errorf("file '%s' is encrypted but no password provided", decodedFileName)
		}
		file.SetPassword(password)
	}

	fileReader, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open file '%s': %v", decodedFileName, err)
	}
	return fileReader, nil
}

// Check if string contains Japanese characters
func containsJapaneseChars(s string) bool {
	for _, r := range s {
		// Japanese character code ranges
		if (r >= 0x3000 && r <= 0x30FF) || // Fullwidth space, Hiragana, Katakana
			(r >= 0x4E00 && r <= 0x9FFF) || // Kanji (CJK Unified Ideographs)
			(r >= 0xFF01 && r <= 0xFF60) || // Fullwidth forms (excluding halfwidth katakana)
			(r >= 0xFFE0 && r <= 0xFFEF) { // Fullwidth symbols
			return true
		}
	}
	return false
}

// Decode CP932 encoded string to UTF-8
func decodeCP932(s string) (string, error) {
	decoder := japanese.ShiftJIS.NewDecoder()
	decodedBytes, _, err := transform.Bytes(decoder, []byte(s))
	if err != nil {
		return "", err
	}
	return string(decodedBytes), nil
}
