package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadOrDefaultMissingFile(t *testing.T) {
	t.Parallel()

	cfg, err := LoadOrDefault(filepath.Join(t.TempDir(), "does-not-exist.json"))
	if err != nil {
		t.Fatalf("LoadOrDefault() error = %v", err)
	}
	if cfg.DefaultTemplate == "" {
		t.Fatalf("DefaultTemplate is empty")
	}
}

func TestLoad(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	content := []byte(`{
  "default_template":"standard",
  "templates":{
    "standard":{
      "name":"Custom",
      "layout":{"width":900,"height":1400,"qr_size":560},
      "style":{"accent_color":"#112233"}
    }
  }
}`)
	if err := os.WriteFile(path, content, 0o644); err != nil {
		t.Fatalf("os.WriteFile() error = %v", err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if got, want := cfg.Templates["standard"].Name, "Custom"; got != want {
		t.Fatalf("template name = %q, want %q", got, want)
	}
	if got, want := cfg.Templates["standard"].Style.AccentHex, "#112233"; got != want {
		t.Fatalf("accent = %q, want %q", got, want)
	}
}
