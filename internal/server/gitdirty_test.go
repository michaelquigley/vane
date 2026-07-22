package server

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/michaelquigley/ranger/internal/api"
)

// gitify seeds a repository over the fixture root and commits everything;
// tests needing git skip when the binary is absent.
func gitify(t *testing.T, root string) {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not installed")
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
}

func cardIn(t *testing.T, lane api.Lane, filename string) api.Card {
	t.Helper()
	for _, c := range lane.Cards {
		if c.Filename == filename {
			return c
		}
	}
	t.Fatalf("no card %s in %s", filename, lane.State)
	return api.Card{}
}

// TestDirtyAbsentWithoutGit pins the unknown tier on every surface: a
// git-less root answers no dirty verdict anywhere — absent, never false.
func TestDirtyAbsentWithoutGit(t *testing.T) {
	s, _ := fixture(t, true)

	res, err := s.GetBoard(context.Background(), api.GetBoardParams{Project: "test"})
	board := mustBoard(t, res, err)
	if board.Dirty.Set {
		t.Errorf("board dirty must be absent without git: %+v", board.Dirty)
	}
	if c := cardIn(t, laneOf(t, board, api.StateResearching), "retry-semantics.md"); c.Dirty.Set {
		t.Errorf("card dirty must be absent without git: %+v", c.Dirty)
	}

	idx, idxErr := s.GetProjects(context.Background())
	if idxErr != nil {
		t.Fatal(idxErr)
	}
	if idx.Projects[0].Dirty.Set {
		t.Errorf("index dirty must be absent without git: %+v", idx.Projects[0])
	}
}

// TestDirtySurfaces walks the verdict across the wire: a committed board
// reports clean everywhere, then one modified item flags its own card, the
// board, and the project index — and the untouched sibling stays clean.
func TestDirtySurfaces(t *testing.T) {
	s, w := fixture(t, true)
	gitify(t, w.Root())

	res, err := s.GetBoard(context.Background(), api.GetBoardParams{Project: "test"})
	board := mustBoard(t, res, err)
	if v, ok := board.Dirty.Get(); !ok || v {
		t.Errorf("committed board must report clean: %+v", board.Dirty)
	}

	path := filepath.Join(w.Root(), "docs", "future", "roadmap", "retry-semantics.md")
	if err := os.WriteFile(path, []byte(retryItem+"\nedited.\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	res, err = s.GetBoard(context.Background(), api.GetBoardParams{Project: "test"})
	board = mustBoard(t, res, err)
	if v, ok := board.Dirty.Get(); !ok || !v {
		t.Errorf("modified item must dirty the board: %+v", board.Dirty)
	}
	if c := cardIn(t, laneOf(t, board, api.StateResearching), "retry-semantics.md"); !c.Dirty.Value {
		t.Errorf("modified card must be dirty: %+v", c.Dirty)
	}
	if c := cardIn(t, laneOf(t, board, api.StateInbox), "board-capture.md"); !c.Dirty.Set || c.Dirty.Value {
		t.Errorf("untouched card must be present-and-clean: %+v", c.Dirty)
	}

	idx, idxErr := s.GetProjects(context.Background())
	if idxErr != nil {
		t.Fatal(idxErr)
	}
	if v, ok := idx.Projects[0].Dirty.Get(); !ok || !v {
		t.Errorf("index must carry the dirty verdict: %+v", idx.Projects[0])
	}

	item, itemErr := s.GetItem(context.Background(), api.GetItemParams{Project: "test", Filename: "retry-semantics.md"})
	if itemErr != nil {
		t.Fatal(itemErr)
	}
	if ok := item.(*api.GetItemOK); !ok.Card.Dirty.Value {
		t.Errorf("item card must carry the verdict: %+v", ok.Card.Dirty)
	}
}
