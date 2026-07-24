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
	// Each embed root must exist: either placeholder (incomplete) or a real staged tree.
	// Production bundle scripts replace placeholders with built assets.
	for _, dir := range []string{
		"embedded/extension",
		"embedded/session-page",
	} {
		base := dir
		if _, err := os.Stat(base); err != nil {
			base = filepath.Join(".", dir)
		}
		st, err := os.Stat(base)
		if err != nil || !st.IsDir() {
			t.Fatalf("missing embed dir %s: %v", dir, err)
		}
		// Accept placeholder OR real payload (manifest.json / index.html).
		ph := filepath.Join(base, "placeholder.txt")
		mani := filepath.Join(base, "manifest.json")
		idx := filepath.Join(base, "index.html")
		_, e1 := os.Stat(ph)
		_, e2 := os.Stat(mani)
		_, e3 := os.Stat(idx)
		if e1 != nil && e2 != nil && e3 != nil {
			t.Fatalf("%s: need placeholder.txt and/or staged payload (manifest.json/index.html)", dir)
		}
	}
}
