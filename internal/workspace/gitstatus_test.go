package workspace

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// gitRepo builds a repository whose roadmap holds one committed item,
// returning the root. tests needing git skip when the binary is absent.
func gitRepo(t *testing.T) string {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not installed")
	}
	root := t.TempDir()
	roadmap := filepath.Join(root, filepath.FromSlash(RoadmapRel))
	if err := os.MkdirAll(roadmap, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(roadmap, "committed.md"), []byte("---\ntitle: committed\nstate: inbox\n---\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	for _, args := range [][]string{
		{"init"},
		{"config", "user.email", "test@example.com"},
		{"config", "user.name", "test"},
		{"add", "."},
		{"commit", "-m", "seed"},
	} {
		cmd := exec.Command("git", append([]string{"-C", root}, args...)...)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	return root
}

func TestGitStatusCleanRepo(t *testing.T) {
	w := New(gitRepo(t))
	status := w.GitStatus()
	if !status.Known || status.Dirty || len(status.Files) != 0 {
		t.Errorf("clean repo: %+v", status)
	}
}

// TestGitStatusDirtyKinds pins that modified, untracked, and staged files
// all count as dirty, keyed by their roadmap-relative paths — an asset in
// a subdirectory dirties the board without naming any item.
func TestGitStatusDirtyKinds(t *testing.T) {
	root := gitRepo(t)
	roadmap := filepath.Join(root, filepath.FromSlash(RoadmapRel))

	if err := os.WriteFile(filepath.Join(roadmap, "committed.md"), []byte("---\ntitle: committed\nstate: horizon\n---\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(roadmap, "untracked.md"), []byte("---\ntitle: untracked\nstate: inbox\n---\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(roadmap, "images"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(roadmap, "images", "sketch.png"), []byte("png"), 0o644); err != nil {
		t.Fatal(err)
	}
	staged := filepath.Join(roadmap, "staged.md")
	if err := os.WriteFile(staged, []byte("---\ntitle: staged\nstate: inbox\n---\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if out, err := exec.Command("git", "-C", root, "add", staged).CombinedOutput(); err != nil {
		t.Fatalf("git add: %v\n%s", err, out)
	}

	status := New(root).GitStatus()
	if !status.Known || !status.Dirty {
		t.Fatalf("dirty repo: %+v", status)
	}
	for _, want := range []string{"committed.md", "untracked.md", "staged.md", "images/sketch.png"} {
		if !status.Files[want] {
			t.Errorf("missing %s in %v", want, status.Files)
		}
	}
	if len(status.Files) != 4 {
		t.Errorf("files = %v", status.Files)
	}
}

// TestGitStatusScopedToRoadmap pins the pathspec: dirt elsewhere in the
// repository never flags the roadmap.
func TestGitStatusScopedToRoadmap(t *testing.T) {
	root := gitRepo(t)
	if err := os.WriteFile(filepath.Join(root, "elsewhere.txt"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	status := New(root).GitStatus()
	if !status.Known || status.Dirty {
		t.Errorf("out-of-roadmap dirt must not flag: %+v", status)
	}
}

// TestGitStatusUnknown covers the degradations: a root that is no git
// repository answers unknown, never an error.
func TestGitStatusUnknown(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not installed")
	}
	status := New(t.TempDir()).GitStatus()
	if status.Known || status.Dirty {
		t.Errorf("non-repo root: %+v", status)
	}
}
