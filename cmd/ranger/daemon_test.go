package main

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/michaelquigley/ranger/internal/config"
)

func writeConfig(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

// TestDaemonBootstrapFailFast covers the daemon's own preconditions: an
// unreadable, invalid, or missing config refuses at startup with a message
// naming the path.
func TestDaemonBootstrapFailFast(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "config.yaml")
	if _, _, err := daemonBootstrap(missing); err == nil || !strings.Contains(err.Error(), missing) || !strings.Contains(err.Error(), "create") {
		t.Errorf("missing config: err = %v", err)
	}
	// an existing-but-broken config gets "fix", never the create hint —
	// the create wording would point wrong when the failure is a duplicate
	// name or a bad override.
	invalid := writeConfig(t, "projects: [\n")
	if _, _, err := daemonBootstrap(invalid); err == nil || !strings.Contains(err.Error(), invalid) || strings.Contains(err.Error(), "at least one project root") {
		t.Errorf("invalid config: err = %v", err)
	}
	empty := writeConfig(t, "projects: []\n")
	if _, _, err := daemonBootstrap(empty); err == nil || !strings.Contains(err.Error(), "at least one") {
		t.Errorf("empty projects: err = %v", err)
	}
}

// TestDaemonBootstrapDegradedRoot is the stage 2 degradation test's
// bootstrap-side mirror: a valid config holding one broken root constructs
// successfully — root health never creeps into bootstrap validation — and
// the project is subsequently reported unavailable.
func TestDaemonBootstrapDegradedRoot(t *testing.T) {
	broken := t.TempDir() // no roadmap directory
	path := writeConfig(t, "projects:\n  - root: "+broken+"\n    name: broken\n")

	cfg, mux, err := daemonBootstrap(path)
	if err != nil {
		t.Fatalf("a broken root must not gate bootstrap: %v", err)
	}
	if cfg.Default != "broken" || cfg.Port != config.DefaultPort {
		t.Errorf("cfg = %+v", cfg)
	}

	srv := httptest.NewServer(mux)
	defer srv.Close()
	resp, err := http.Get(srv.URL + "/api/v1/projects")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	var idx struct {
		Projects []struct {
			Name      string `json:"name"`
			Available bool   `json:"available"`
			Error     string `json:"error"`
		} `json:"projects"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&idx); err != nil {
		t.Fatal(err)
	}
	if len(idx.Projects) != 1 || idx.Projects[0].Available || idx.Projects[0].Error == "" {
		t.Errorf("index = %+v", idx)
	}
}

// TestDaemonSourceReadsFresh pins the live-config surface: an edit to the
// file is visible on the next request, no restart.
func TestDaemonSourceReadsFresh(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "docs", "future", "roadmap"), 0o755); err != nil {
		t.Fatal(err)
	}
	path := writeConfig(t, "projects:\n  - root: "+root+"\n    name: before\n")
	_, mux, err := daemonBootstrap(path)
	if err != nil {
		t.Fatal(err)
	}
	srv := httptest.NewServer(mux)
	defer srv.Close()

	get := func() string {
		t.Helper()
		resp, err := http.Get(srv.URL + "/api/v1/projects")
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatal(err)
		}
		return string(body)
	}

	if body := get(); !strings.Contains(body, "before") {
		t.Fatalf("index = %s", body)
	}
	if err := os.WriteFile(path, []byte("projects:\n  - root: "+root+"\n    name: after\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if body := get(); !strings.Contains(body, "after") || strings.Contains(body, "before") {
		t.Errorf("config edit must land on the next request: %s", body)
	}
}

func TestResolvePort(t *testing.T) {
	cfg := &config.Config{Port: 4200}
	if got := resolvePort(cfg, config.DefaultPort, false); got != 4200 {
		t.Errorf("unchanged flag must yield the config's port, got %d", got)
	}
	if got := resolvePort(cfg, 4300, true); got != 4300 {
		t.Errorf("explicit --port must win, got %d", got)
	}
	// cobra's Changed, never a zero-value check: an explicit --port that
	// happens to equal the flag default still wins over the config.
	if got := resolvePort(cfg, config.DefaultPort, true); got != config.DefaultPort {
		t.Errorf("explicit default-valued --port must win, got %d", got)
	}
}
