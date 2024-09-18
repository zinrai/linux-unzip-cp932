package main

import (
	"archive/zip"
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
)

func TestExtractZip(t *testing.T) {
	// テスト用のファイル名（CP932でエンコード）
	filename := encodeCP932("テスト文書.txt")
	content := []byte("これはテスト文書です。")

	// ZIP ファイルの作成
	buf := new(bytes.Buffer)
	w := zip.NewWriter(buf)
	f, err := w.Create(string(filename))
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

	// 一時ディレクトリの作成
	tempDir, err := os.MkdirTemp("", "cp932test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// ZIP ファイルの保存
	zipPath := filepath.Join(tempDir, "test.zip")
	err = os.WriteFile(zipPath, buf.Bytes(), 0644)
	if err != nil {
		t.Fatalf("Failed to write zip file: %v", err)
	}

	// ZIP ファイルの解凍
	extractDir := filepath.Join(tempDir, "extracted")
	err = extractZip(zipPath, extractDir)
	if err != nil {
		t.Fatalf("Failed to extract zip: %v", err)
	}

	// 解凍されたファイルの確認
	extractedPath := filepath.Join(extractDir, "テスト文書.txt")
	if _, err := os.Stat(extractedPath); os.IsNotExist(err) {
		t.Fatalf("Extracted file does not exist: %v", err)
	}

	// ファイル内容の確認
	extractedContent, err := os.ReadFile(extractedPath)
	if err != nil {
		t.Fatalf("Failed to read extracted file: %v", err)
	}

	if string(extractedContent) != string(content) {
		t.Fatalf("Extracted content does not match original. Got %s, want %s", extractedContent, content)
	}

	t.Log("Test passed successfully!")
}

func encodeCP932(s string) []byte {
	encoder := japanese.ShiftJIS.NewEncoder()
	encoded, _, _ := transform.Bytes(encoder, []byte(s))
	return encoded
}
