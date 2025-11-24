package meta_test

import (
	"encoding/base64"
	"fmt"
	"os"
	"testing"

	"user/core/coretype"
	"user/core/meta"
)

// Моковый хешер
type mockHasher struct{}

func (m *mockHasher) HashFile(filePath string, pieceSize int) ([]coretype.Piece, error) {
	// Возвращаем фиктивные куски
	return []coretype.Piece{
		{Index: 0, Hash: [20]byte{1, 2, 3}, Size: 10},
		{Index: 1, Hash: [20]byte{4, 5, 6}, Size: 10},
	}, nil
}

// Вспомогательная функция: создание временного файла
func createTempFile(t *testing.T, content []byte) string {
	tmpFile, err := os.CreateTemp("", "meta_test_")
	if err != nil {
		t.Fatalf("cannot create temp file: %v", err)
	}
	if _, err := tmpFile.Write(content); err != nil {
		tmpFile.Close()
		t.Fatalf("cannot write to temp file: %v", err)
	}
	tmpFile.Close()
	/*
	tmpFileInfo, _ :=tmpFile.Stat()
	fmt.Println(tmpFileInfo.Name())
	*/
	return tmpFile.Name()
}


// Проверка генерации .mytorrent
func TestGenerateMyTorrent(t *testing.T) {
	content := []byte("Hello Meta!")
	tmpFile := createTempFile(t, content)
	defer os.Remove(tmpFile)
	mytorrentPath := tmpFile[5:] + ".mytorrent"
	defer os.Remove(mytorrentPath ) 

	generator := &meta.MetaGenerator{Hasher: &mockHasher{}}
	fmt.Println(tmpFile,mytorrentPath)
	err := generator.GenerateMyTorrent(tmpFile, 10, "127.0.0.1")
	if err != nil {
		t.Fatalf("GenerateMyTorrent returned error: %v", err)
	}

	if _, err := os.Stat(mytorrentPath); err != nil {
		t.Fatalf(".mytorrent file was not created: %v", err)
	}

	// Проверяем содержимое файла
	data, err := os.ReadFile(mytorrentPath)
	if err != nil {
		t.Fatalf("cannot read .mytorrent: %v", err)
	}

	if len(data) == 0 {
		t.Fatal(".mytorrent file is empty")
	}
}


// Проверка ConvertPiecesToBase64
func TestConvertPiecesToBase64(t *testing.T) {
	pieces := []coretype.Piece{
		{Index: 0, Hash: [20]byte{0, 1, 2}},
		{Index: 1, Hash: [20]byte{3, 4, 5}},
	}
	base64Strings := meta.ConvertPiecesToBase64(pieces)

	if len(base64Strings) != len(pieces) {
		t.Fatalf("expected %d strings, got %d", len(pieces), len(base64Strings))
	}

	expected := base64.StdEncoding.EncodeToString(pieces[0].Hash[:])
	if base64Strings[0] != expected {
		t.Fatalf("unexpected base64 for first piece: got %s, want %s", base64Strings[0], expected)
	}
}


// ConvertPiecesToBase64 с пустым слайсом
func TestConvertPiecesToBase64_Empty(t *testing.T) {
	result := meta.ConvertPiecesToBase64([]coretype.Piece{})
	if result != nil {
		t.Fatalf("expected nil for empty input, got %v", result)
	}
}

