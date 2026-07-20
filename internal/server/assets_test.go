package server

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestAssets(t *testing.T) {
	parent := t.TempDir()
	dir := filepath.Join(parent, "roadmap")
	if err := os.MkdirAll(filepath.Join(dir, "images"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "images", "pic.png"), []byte("png-bytes"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(parent, "outside.txt"), []byte("outside"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(parent, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(parent, ".git", "config"), []byte("git-bytes"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(parent, ".git", "config"), filepath.Join(dir, "escape.txt")); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(dir, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".git", "config"), []byte("nested-git"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(".git/config", filepath.Join(dir, "inroot-link")); err != nil {
		t.Fatal(err)
	}
	h := Assets(dir)

	get := func(target string) *httptest.ResponseRecorder {
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, target, nil))
		return rec
	}

	if rec := get("/images/pic.png"); rec.Code != http.StatusOK || rec.Body.String() != "png-bytes" {
		t.Fatalf("file: got %d %q", rec.Code, rec.Body.String())
	}
	if rec := get("/../outside.txt"); rec.Code == http.StatusOK {
		t.Fatalf("traversal: got %d %q", rec.Code, rec.Body.String())
	}
	if rec := get("/images"); rec.Code != http.StatusNotFound {
		t.Fatalf("directory: got %d", rec.Code)
	}
	if rec := get("/escape.txt"); rec.Code != http.StatusNotFound {
		t.Fatalf("symlink escape: got %d %q", rec.Code, rec.Body.String())
	}
	if rec := get("/.git/config"); rec.Code != http.StatusNotFound {
		t.Fatalf("git component: got %d %q", rec.Code, rec.Body.String())
	}
	if rec := get("/inroot-link"); rec.Code != http.StatusNotFound {
		t.Fatalf("in-root symlink: got %d %q", rec.Code, rec.Body.String())
	}
	if rec := get("/images/missing.png"); rec.Code != http.StatusNotFound {
		t.Fatalf("missing: got %d", rec.Code)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/images/pic.png", nil))
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("post: got %d", rec.Code)
	}

	linkRoot := filepath.Join(parent, "roadmap-link")
	if err := os.Symlink(filepath.Join(parent, ".git"), linkRoot); err != nil {
		t.Fatal(err)
	}
	linked := Assets(linkRoot)
	rec = httptest.NewRecorder()
	linked.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/config", nil))
	if rec.Code != http.StatusNotFound {
		t.Fatalf("symlinked root: got %d %q", rec.Code, rec.Body.String())
	}
}
