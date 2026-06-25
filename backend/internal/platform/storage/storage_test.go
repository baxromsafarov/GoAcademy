package storage

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLocalStorage_SaveAndDelete(t *testing.T) {
	dir := t.TempDir()
	st, err := NewLocalStorage(dir, "http://localhost:8080/static/")
	if err != nil {
		t.Fatalf("NewLocalStorage: %v", err)
	}

	url, err := st.Save(context.Background(), "avatars/u1.png", strings.NewReader("PNGDATA"))
	if err != nil {
		t.Fatalf("Save: %v", err)
	}
	if url != "http://localhost:8080/static/avatars/u1.png" {
		t.Errorf("url = %q", url)
	}

	got, err := os.ReadFile(filepath.Join(dir, "avatars", "u1.png"))
	if err != nil {
		t.Fatalf("file not written: %v", err)
	}
	if string(got) != "PNGDATA" {
		t.Errorf("contents = %q", got)
	}

	if err := st.Delete(context.Background(), "avatars/u1.png"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "avatars", "u1.png")); !os.IsNotExist(err) {
		t.Error("file should be deleted")
	}
	// Deleting a missing object is not an error.
	if err := st.Delete(context.Background(), "avatars/u1.png"); err != nil {
		t.Errorf("Delete (missing) = %v, want nil", err)
	}
}

func TestLocalStorage_RejectsTraversal(t *testing.T) {
	st, _ := NewLocalStorage(t.TempDir(), "http://x/static")
	if _, err := st.Save(context.Background(), "../escape.txt", strings.NewReader("x")); err == nil {
		t.Error("expected traversal key to be rejected")
	}
}
