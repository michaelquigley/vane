package workspace

import (
	"os/exec"
	"strings"
)

// GitStatus is one fresh read of the roadmap directory's uncommitted
// state. read is the operative word: the 2026-07-21 design change opened
// read-only status inspection through the git CLI and nothing more — the
// tool still never writes git state; commit, push, and history remain the
// operator's judgment gate.
type GitStatus struct {
	// Known reports whether git answered: false without a git binary,
	// outside a repository, or on any git failure. unknown is the absence
	// of information, never an error and never cleanliness.
	Known bool
	// Dirty is true when anything under the roadmap directory is
	// uncommitted — items, order.yaml, and assets alike.
	Dirty bool
	// Files holds the dirty paths relative to the roadmap directory, so
	// flat item filenames key directly.
	Files map[string]bool
}

// GitStatus shells out to `git status --porcelain` scoped to the roadmap
// directory — read-only, fresh per call, and modified, staged, and
// untracked all count: every one is work the operator hasn't committed.
func (w *Workspace) GitStatus() GitStatus {
	out, err := exec.Command("git", "-C", w.root, "status", "--porcelain", "-z", "--untracked-files=all", "--", RoadmapRel).Output()
	if err != nil {
		return GitStatus{}
	}
	status := GitStatus{Known: true, Files: map[string]bool{}}
	fields := strings.Split(string(out), "\x00")
	for i := 0; i < len(fields); i++ {
		entry := fields[i]
		if len(entry) < 4 {
			continue
		}
		xy, path := entry[:2], entry[3:]
		// a rename or copy entry carries the original path as its own
		// NUL-separated field; consume it without recording — the file it
		// names no longer exists to highlight.
		if strings.ContainsAny(xy, "RC") {
			i++
		}
		if rel, ok := strings.CutPrefix(path, RoadmapRel+"/"); ok {
			status.Files[rel] = true
			status.Dirty = true
		}
	}
	return status
}
