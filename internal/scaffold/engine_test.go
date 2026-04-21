package scaffold

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRenderTemplate(t *testing.T) {
	tmpl := "Hello {{.ProjectName}} at {{.Module}}"
	params := Params{ProjectName: "myapp", Module: "github.com/test/myapp"}
	result, err := RenderTemplate("test", tmpl, params)
	if err != nil {
		t.Fatalf("RenderTemplate: %v", err)
	}
	want := "Hello myapp at github.com/test/myapp"
	if result != want {
		t.Errorf("got %q, want %q", result, want)
	}
}

func TestWriteFile_CreatesNewFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sub", "file.txt")
	written, err := WriteFile(path, "content", 0o644)
	if err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	if !written {
		t.Error("WriteFile returned false, want true")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "content" {
		t.Errorf("file content = %q, want %q", string(data), "content")
	}
}

func TestWriteFile_SkipsExisting(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "file.txt")
	_ = os.WriteFile(path, []byte("original"), 0o644)
	written, err := WriteFile(path, "new content", 0o644)
	if err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	if written {
		t.Error("WriteFile returned true for existing file, want false")
	}
	data, _ := os.ReadFile(path)
	if string(data) != "original" {
		t.Errorf("file was overwritten: got %q", string(data))
	}
}
