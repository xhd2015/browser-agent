package browseragent

import (
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"
)

func TestEmbedCompleteFS_placeholderOnly(t *testing.T) {
	fsys := fstest.MapFS{
		"placeholder.txt": &fstest.MapFile{Data: []byte("incomplete\n")},
	}
	if EmbedCompleteFS(fsys, AssetKindExtension) {
		t.Fatal("placeholder-only extension must be incomplete")
	}
	if EmbedCompleteFS(fsys, AssetKindSessionPage) {
		t.Fatal("placeholder-only session-page must be incomplete")
	}
}

func TestDiskPlaceholderLayout_gitKeepFiles(t *testing.T) {
	// Repo must ship placeholders at these paths (tracked).
	for _, p := range []string{
		"embedded/extension/placeholder.txt",
		"embedded/session-page/placeholder.txt",
	} {
		if _, err := os.Stat(filepath.Join(".", p)); err != nil {
			// tests run with package dir = browseragent/
			if _, err2 := os.Stat(p); err2 != nil {
				t.Fatalf("missing %s: %v / %v", p, err, err2)
			}
		}
	}
}
