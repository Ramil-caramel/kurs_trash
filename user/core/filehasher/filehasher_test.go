package filehasher_test

import (
	"crypto/sha1"
	"os"
	"testing"

	"user/core/filehasher"
)

// Вспомогательная функция: min
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Тест с реальным файлом из testdata
func TestHashFile_RealFile(t *testing.T) {
	hasher := &filehasher.FileHasher{}
	pieceSize := 1024 // 1 KB
	path := "./testdata/avidreaders.ru__oblomov.txt"

	if _, err := os.Stat(path); err != nil {
		t.Fatalf("test file not found: %v", err)
	}

	pieces, err := hasher.HashFile(path, pieceSize)
	if err != nil {
		t.Fatalf("HashFile returned error: %v", err)
	}

	if len(pieces) == 0 {
		t.Fatal("expected non-empty pieces")
	}

	// Проверяем первый кусок вручную
	data, _ := os.ReadFile(path)
	expected := sha1.Sum(data[:min(pieceSize, len(data))])
	if pieces[0].Hash != expected {
		t.Fatalf("invalid hash for first piece:\n got:  %x\n want: %x", pieces[0].Hash, expected)
	}
}

// Тест с несуществующим файлом
func TestHashFile_FileNotExist(t *testing.T) {
	hasher := &filehasher.FileHasher{}
	_, err := hasher.HashFile("no_such_file.txt", 1024)
	if err == nil {
		t.Fatal("expected error for non-existing file")
	}
}

// 3. Тест с маленьким файлом (меньше pieceSize)
func TestHashFile_SmallFile(t *testing.T) {
	hasher := &filehasher.FileHasher{}
	content := []byte("Hello, world!") // всего 13 байт
	tmpFile := createTempFile(t, content)
	defer os.Remove(tmpFile)

	pieceSize := 1024 // больше, чем размер файла
	pieces, err := hasher.HashFile(tmpFile, pieceSize)
	if err != nil {
		t.Fatalf("HashFile error: %v", err)
	}

	if len(pieces) != 1 {
		t.Fatalf("expected 1 piece, got %d", len(pieces))
	}

	expected := sha1.Sum(content)
	if pieces[0].Hash != expected {
		t.Fatalf("invalid hash:\n got:  %x\n want: %x", pieces[0].Hash, expected)
	}
}

// Тест с файлом, размер которого кратен pieceSize
func TestHashFile_ExactMultiple(t *testing.T) {
	hasher := &filehasher.FileHasher{}
	content := make([]byte, 4096) // 4 KB
	for i := range content {
		content[i] = byte(i % 256)
	}
	tmpFile := createTempFile(t, content)
	defer os.Remove(tmpFile)

	pieceSize := 1024 // 1 KB
	pieces, err := hasher.HashFile(tmpFile, pieceSize)
	if err != nil {
		t.Fatalf("HashFile error: %v", err)
	}

	if len(pieces) != 4 {
		t.Fatalf("expected 4 pieces, got %d", len(pieces))
	}

	// Проверяем каждый кусок
	for i := 0; i < 4; i++ {
		start := i * pieceSize
		end := start + pieceSize
		expected := sha1.Sum(content[start:end])
		if pieces[i].Hash != expected {
			t.Fatalf("invalid hash for piece %d", i)
		}
	}
}

// Тест с файлом, размер которого не кратен pieceSize
func TestHashFile_NotMultiple(t *testing.T) {
	hasher := &filehasher.FileHasher{}
	content := make([]byte, 2500) // не кратно 1024
	for i := range content {
		content[i] = byte(i % 256)
	}
	tmpFile := createTempFile(t, content)
	defer os.Remove(tmpFile)

	pieceSize := 1024
	pieces, err := hasher.HashFile(tmpFile, pieceSize)
	if err != nil {
		t.Fatalf("HashFile error: %v", err)
	}

	expectedPieces := 3 // 1024 + 1024 + 452
	if len(pieces) != expectedPieces {
		t.Fatalf("expected %d pieces, got %d", expectedPieces, len(pieces))
	}

	// Проверяем последний кусок
	lastStart := 2 * pieceSize
	expected := sha1.Sum(content[lastStart:])
	if pieces[2].Hash != expected {
		t.Fatal("invalid hash for last piece")
	}
}

// 6. Тест с pieceSize = 1
func TestHashFile_PieceSizeOne(t *testing.T) {
	hasher := &filehasher.FileHasher{}
	content := []byte("abc")
	tmpFile := createTempFile(t, content)
	defer os.Remove(tmpFile)

	pieceSize := 1
	pieces, err := hasher.HashFile(tmpFile, pieceSize)
	if err != nil {
		t.Fatalf("HashFile error: %v", err)
	}

	if len(pieces) != 3 {
		t.Fatalf("expected 3 pieces, got %d", len(pieces))
	}

	for i := range content {
		expected := sha1.Sum([]byte{content[i]})
		if pieces[i].Hash != expected {
			t.Fatalf("invalid hash for piece %d", i)
		}
	}
}

// Вспомогательная функция: создание временного файла с данными
func createTempFile(t *testing.T, content []byte) string {
	tmpFile, err := os.CreateTemp("", "filehasher_test_")
	if err != nil {
		t.Fatalf("cannot create temp file: %v", err)
	}
	if _, err := tmpFile.Write(content); err != nil {
		tmpFile.Close()
		t.Fatalf("cannot write to temp file: %v", err)
	}
	tmpFile.Close()
	return tmpFile.Name()
}

