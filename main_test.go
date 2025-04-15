package main

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/alexmullins/zip"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
)

func TestExtractZip(t *testing.T) {
	runTest(t, false, "")
}

func TestExtractEncryptedZip(t *testing.T) {
	runTest(t, true, "testpassword")
}

func TestExtractZipWithGPB11(t *testing.T) {
	runTestWithGPB11(t, false, "")
}

func TestExtractEncryptedZipWithGPB11(t *testing.T) {
	runTestWithGPB11(t, true, "testpassword")
}

func runTest(t *testing.T, encrypted bool, password string) {
	// Japanese filename encoded in CP932 (without GPB11 flag)
	filename := encodeCP932("テスト文書.txt")
	content := []byte("This is a test document.")

	// Create ZIP file
	buf := new(bytes.Buffer)
	w := zip.NewWriter(buf)
	var f io.Writer
	var err error
	if encrypted {
		f, err = w.Encrypt(string(filename), password)
	} else {
		f, err = w.Create(string(filename))
	}
	if err != nil {
		t.Fatalf("Failed to create file in zip: %v", err)
	}
	_, err = f.Write(content)
	if err != nil {
		t.Fatalf("Failed to write content to zip: %v", err)
	}
	err = w.Close()
	if err != nil {
		t.Fatalf("Failed to close zip writer: %v", err)
	}

	// Create temp directory
	tempDir, err := os.MkdirTemp("", "cp932test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Save ZIP file
	zipPath := filepath.Join(tempDir, "test.zip")
	err = os.WriteFile(zipPath, buf.Bytes(), 0644)
	if err != nil {
		t.Fatalf("Failed to write zip file: %v", err)
	}

	// Extract ZIP file
	extractDir := filepath.Join(tempDir, "extracted")
	err = extractZip(zipPath, extractDir, password)
	if err != nil {
		t.Fatalf("Failed to extract zip: %v", err)
	}

	// Check extracted file
	extractedPath := filepath.Join(extractDir, "テスト文書.txt")
	if _, err := os.Stat(extractedPath); os.IsNotExist(err) {
		t.Fatalf("Extracted file does not exist: %v", err)
	}

	// Check file content
	extractedContent, err := os.ReadFile(extractedPath)
	if err != nil {
		t.Fatalf("Failed to read extracted file: %v", err)
	}

	if string(extractedContent) != string(content) {
		t.Fatalf("Extracted content does not match original. Got %s, want %s", extractedContent, content)
	}

	t.Log("Test passed successfully for CP932 encoded filename without GPB11 flag")
}

func runTestWithGPB11(t *testing.T, encrypted bool, password string) {
	// Japanese filename in UTF-8 (with GPB11 flag set)
	filename := "テスト文書_UTF8.txt"
	content := []byte("This is a test document with UTF-8 filename.")

	// Create ZIP file
	buf := new(bytes.Buffer)
	w := zip.NewWriter(buf)

	// To set GPB11 flag, we need to modify the internal structure
	// Here we use a custom header to set GPB11 flag
	fh := &zip.FileHeader{
		Name:   filename,
		Method: zip.Deflate,
	}

	// Set GPB11 flag (0x0800)
	fh.Flags |= 0x0800

	var f io.Writer
	var err error

	if encrypted {
		fh.SetPassword(password)
		f, err = w.CreateHeader(fh)
	} else {
		f, err = w.CreateHeader(fh)
	}

	if err != nil {
		t.Fatalf("Failed to create file in zip with GPB11: %v", err)
	}

	_, err = f.Write(content)
	if err != nil {
		t.Fatalf("Failed to write content to zip with GPB11: %v", err)
	}

	err = w.Close()
	if err != nil {
		t.Fatalf("Failed to close zip writer with GPB11: %v", err)
	}

	// Create temp directory
	tempDir, err := os.MkdirTemp("", "utf8test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Save ZIP file
	zipPath := filepath.Join(tempDir, "test_utf8.zip")
	err = os.WriteFile(zipPath, buf.Bytes(), 0644)
	if err != nil {
		t.Fatalf("Failed to write zip file with GPB11: %v", err)
	}

	// Extract ZIP file
	extractDir := filepath.Join(tempDir, "extracted_utf8")
	err = extractZip(zipPath, extractDir, password)
	if err != nil {
		t.Fatalf("Failed to extract zip with GPB11: %v", err)
	}

	// Check extracted file
	extractedPath := filepath.Join(extractDir, filename)
	if _, err := os.Stat(extractedPath); os.IsNotExist(err) {
		t.Fatalf("Extracted file with GPB11 does not exist: %v", err)
	}

	// Check file content
	extractedContent, err := os.ReadFile(extractedPath)
	if err != nil {
		t.Fatalf("Failed to read extracted file with GPB11: %v", err)
	}

	if string(extractedContent) != string(content) {
		t.Fatalf("Extracted content with GPB11 does not match original. Got %s, want %s", extractedContent, content)
	}

	t.Log("Test passed successfully for UTF-8 encoded filename with GPB11 flag")
}

func TestMixedEncodingZip(t *testing.T) {
	// Create temp directory
	tempDir, err := os.MkdirTemp("", "mixedtest")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Setup files for mixed encoding test
	cp932Filename := encodeCP932("日本語ファイル.txt")
	utf8Filename := "UTF8ファイル.txt"
	asciiFilename := "english_file.txt"

	content1 := []byte("CP932 encoded filename content")
	content2 := []byte("UTF-8 encoded filename content")
	content3 := []byte("ASCII filename content")

	// Create ZIP file with mixed encodings
	buf := new(bytes.Buffer)
	w := zip.NewWriter(buf)

	// Add CP932 encoded filename (no GPB11)
	f1, err := w.Create(string(cp932Filename))
	if err != nil {
		t.Fatalf("Failed to create CP932 file in zip: %v", err)
	}
	_, err = f1.Write(content1)
	if err != nil {
		t.Fatalf("Failed to write CP932 content: %v", err)
	}

	// Add UTF-8 encoded filename (with GPB11)
	fh := &zip.FileHeader{
		Name:   utf8Filename,
		Method: zip.Deflate,
	}
	fh.Flags |= 0x0800 // Set GPB11 flag

	f2, err := w.CreateHeader(fh)
	if err != nil {
		t.Fatalf("Failed to create UTF-8 file in zip: %v", err)
	}
	_, err = f2.Write(content2)
	if err != nil {
		t.Fatalf("Failed to write UTF-8 content: %v", err)
	}

	// Add ASCII filename
	f3, err := w.Create(asciiFilename)
	if err != nil {
		t.Fatalf("Failed to create ASCII file in zip: %v", err)
	}
	_, err = f3.Write(content3)
	if err != nil {
		t.Fatalf("Failed to write ASCII content: %v", err)
	}

	err = w.Close()
	if err != nil {
		t.Fatalf("Failed to close zip writer: %v", err)
	}

	// Save ZIP file
	zipPath := filepath.Join(tempDir, "mixed_encoding.zip")
	err = os.WriteFile(zipPath, buf.Bytes(), 0644)
	if err != nil {
		t.Fatalf("Failed to write mixed encoding zip file: %v", err)
	}

	// Extract ZIP file
	extractDir := filepath.Join(tempDir, "extracted_mixed")
	err = extractZip(zipPath, extractDir, "")
	if err != nil {
		t.Fatalf("Failed to extract mixed encoding zip: %v", err)
	}

	// Check all extracted files
	extractedCP932Path := filepath.Join(extractDir, "日本語ファイル.txt")
	extractedUTF8Path := filepath.Join(extractDir, utf8Filename)
	extractedASCIIPath := filepath.Join(extractDir, asciiFilename)

	// Check CP932 file
	if _, err := os.Stat(extractedCP932Path); os.IsNotExist(err) {
		t.Fatalf("Extracted CP932 file does not exist: %v", err)
	}
	extractedContent1, err := os.ReadFile(extractedCP932Path)
	if err != nil {
		t.Fatalf("Failed to read extracted CP932 file: %v", err)
	}
	if string(extractedContent1) != string(content1) {
		t.Fatalf("Extracted CP932 content does not match. Got %s, want %s", extractedContent1, content1)
	}

	// Check UTF-8 file
	if _, err := os.Stat(extractedUTF8Path); os.IsNotExist(err) {
		t.Fatalf("Extracted UTF-8 file does not exist: %v", err)
	}
	extractedContent2, err := os.ReadFile(extractedUTF8Path)
	if err != nil {
		t.Fatalf("Failed to read extracted UTF-8 file: %v", err)
	}
	if string(extractedContent2) != string(content2) {
		t.Fatalf("Extracted UTF-8 content does not match. Got %s, want %s", extractedContent2, content2)
	}

	// Check ASCII file
	if _, err := os.Stat(extractedASCIIPath); os.IsNotExist(err) {
		t.Fatalf("Extracted ASCII file does not exist: %v", err)
	}
	extractedContent3, err := os.ReadFile(extractedASCIIPath)
	if err != nil {
		t.Fatalf("Failed to read extracted ASCII file: %v", err)
	}
	if string(extractedContent3) != string(content3) {
		t.Fatalf("Extracted ASCII content does not match. Got %s, want %s", extractedContent3, content3)
	}

	t.Log("Mixed encoding test passed successfully")
}

// Helper function to encode string in CP932
func encodeCP932(s string) []byte {
	encoder := japanese.ShiftJIS.NewEncoder()
	encoded, _, _ := transform.Bytes(encoder, []byte(s))
	return encoded
}

func TestGetDecodedFileNameWithRealZipFiles(t *testing.T) {
	// Test cases for getDecodedFileName function using real ZIP files
	testCases := []struct {
		name         string
		filename     string
		setGPB11Flag bool
		expected     string
	}{
		{
			name:         "ASCII filename without GPB11",
			filename:     "test.txt",
			setGPB11Flag: false,
			expected:     "test.txt",
		},
		{
			name:         "ASCII filename with GPB11",
			filename:     "test.txt",
			setGPB11Flag: true,
			expected:     "test.txt",
		},
		{
			name:         "Japanese filename without GPB11",
			filename:     string(encodeCP932("テスト.txt")),
			setGPB11Flag: false,
			expected:     "テスト.txt",
		},
		{
			name:         "Japanese filename with GPB11",
			filename:     "テスト.txt",
			setGPB11Flag: true,
			expected:     "テスト.txt",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a real ZIP file for testing
			buf := new(bytes.Buffer)
			w := zip.NewWriter(buf)

			var f io.Writer
			var err error

			if tc.setGPB11Flag {
				// Create with GPB11 flag
				fh := &zip.FileHeader{
					Name:   tc.filename,
					Method: zip.Deflate,
				}
				fh.Flags |= 0x0800
				f, err = w.CreateHeader(fh)
			} else {
				// Create without GPB11 flag
				f, err = w.Create(tc.filename)
			}

			if err != nil {
				t.Fatalf("Failed to create test file in zip: %v", err)
			}

			// Write some dummy content
			_, err = f.Write([]byte("test content"))
			if err != nil {
				t.Fatalf("Failed to write test content: %v", err)
			}

			err = w.Close()
			if err != nil {
				t.Fatalf("Failed to close zip writer: %v", err)
			}

			// Create a temp file for the ZIP
			tempFile, err := os.CreateTemp("", "ziptest*.zip")
			if err != nil {
				t.Fatalf("Failed to create temp file: %v", err)
			}
			tempPath := tempFile.Name()
			defer os.Remove(tempPath)
			tempFile.Close()

			// Write the ZIP data to the temp file
			err = os.WriteFile(tempPath, buf.Bytes(), 0644)
			if err != nil {
				t.Fatalf("Failed to write zip file: %v", err)
			}

			// Open the ZIP file to get the real zip.File
			reader, err := zip.OpenReader(tempPath)
			if err != nil {
				t.Fatalf("Failed to open zip: %v", err)
			}
			defer reader.Close()

			if len(reader.File) != 1 {
				t.Fatalf("Expected 1 file in ZIP, got %d", len(reader.File))
			}

			// Get the actual file from the ZIP
			zipFile := reader.File[0]

			// Call the function being tested
			result := getDecodedFileName(zipFile, 0, 1)

			// Check if the result matches expected
			if result != tc.expected {
				t.Errorf("Expected decoded filename %q, got %q", tc.expected, result)
			}
		})
	}
}

func TestIsPathSafe(t *testing.T) {
	testCases := []struct {
		path     string
		expected bool
	}{
		{"normal/path.txt", true},
		{"../parent/path.txt", false},
		{"/absolute/path.txt", false},
		{"../../dangerous/path.txt", false},
		{"./relative/path.txt", true},
		{"just_filename.txt", true},
	}

	for _, tc := range testCases {
		result := isPathSafe(tc.path)
		if result != tc.expected {
			t.Errorf("isPathSafe(%q) = %v, expected %v", tc.path, result, tc.expected)
		}
	}
}

func TestContainsJapaneseChars(t *testing.T) {
	testCases := []struct {
		str      string
		expected bool
	}{
		{"ASCII only", false},
		{"123456", false},
		{"テスト", true},
		{"Test　with　Japanese　space", true}, // Full-width space is Japanese
		{"漢字も日本語です", true},
		{"ｶﾀｶﾅ", false}, // Half-width katakana is not in the ranges we check
		{"１２３", true},   // Full-width digits
	}

	for _, tc := range testCases {
		result := containsJapaneseChars(tc.str)
		if result != tc.expected {
			t.Errorf("containsJapaneseChars(%q) = %v, expected %v", tc.str, result, tc.expected)
		}
	}
}
