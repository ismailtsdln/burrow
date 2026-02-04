package cleaner

import (
	"os"
	"path/filepath"
	"testing"
)

func TestTrashManager_MovePath(t *testing.T) {
	tm := NewTrashManager()
	tempDir, err := os.MkdirTemp("", "burrow-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	tm.TrashBaseDir = filepath.Join(tempDir, "trash")

	src := filepath.Join(tempDir, "test.txt")
	if err := os.WriteFile(src, []byte("hello"), 0644); err != nil {
		t.Fatal(err)
	}

	dst := filepath.Join(tm.TrashBaseDir, "test.txt")
	if err := os.MkdirAll(tm.TrashBaseDir, 0755); err != nil {
		t.Fatal(err)
	}

	if err := tm.movePath(src, dst); err != nil {
		t.Fatalf("movePath failed: %v", err)
	}

	if _, err := os.Stat(src); !os.IsNotExist(err) {
		t.Errorf("source file still exists")
	}

	content, err := os.ReadFile(dst)
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != "hello" {
		t.Errorf("content mismatch: got %s, want hello", string(content))
	}
}

func TestTrashManager_CopyPath_Dir(t *testing.T) {
	tm := NewTrashManager()
	tempDir, err := os.MkdirTemp("", "burrow-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	srcDir := filepath.Join(tempDir, "src")
	if err := os.MkdirAll(filepath.Join(srcDir, "sub"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(srcDir, "file1.txt"), []byte("1"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(srcDir, "sub", "file2.txt"), []byte("2"), 0644); err != nil {
		t.Fatal(err)
	}

	dstDir := filepath.Join(tempDir, "dst")
	if err := tm.copyPath(srcDir, dstDir); err != nil {
		t.Fatalf("copyPath failed: %v", err)
	}

	// Verify content
	content1, _ := os.ReadFile(filepath.Join(dstDir, "file1.txt"))
	if string(content1) != "1" {
		t.Errorf("file1 content mismatch: %s", string(content1))
	}
	content2, _ := os.ReadFile(filepath.Join(dstDir, "sub", "file2.txt"))
	if string(content2) != "2" {
		t.Errorf("file2 content mismatch: %s", string(content2))
	}
}
