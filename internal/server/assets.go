package server

import (
	"net/http"
	"os"
	"path"
	"strings"
)

// Assets serves the roadmap directory's files read-only — the relative
// images and attachments item bodies reference. fresh disk read per
// request, files only: a directory request 404s rather than listing. no
// dot-prefixed component is ever served and no symlink component is
// followed — assets are real files physically under the roadmap
// directory, so the no-git boundary holds on this route without the code
// ever naming .git. the root itself is verified non-symlink per request,
// which closes the last door: nothing this handler opens can be git
// metadata, categorically.
func Assets(dir string) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			http.Error(rw, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		name := strings.TrimPrefix(path.Clean("/"+r.URL.Path), "/")
		if name == "" {
			http.NotFound(rw, r)
			return
		}
		parts := strings.Split(name, "/")
		for _, part := range parts {
			if strings.HasPrefix(part, ".") {
				http.NotFound(rw, r)
				return
			}
		}
		if info, err := os.Lstat(dir); err != nil || info.Mode()&os.ModeSymlink != 0 {
			http.NotFound(rw, r)
			return
		}
		root, err := os.OpenRoot(dir)
		if err != nil {
			http.NotFound(rw, r)
			return
		}
		defer root.Close()
		var info os.FileInfo
		for i := range parts {
			info, err = root.Lstat(path.Join(parts[:i+1]...))
			if err != nil || info.Mode()&os.ModeSymlink != 0 {
				http.NotFound(rw, r)
				return
			}
		}
		if !info.Mode().IsRegular() {
			http.NotFound(rw, r)
			return
		}
		f, err := root.Open(name)
		if err != nil {
			http.NotFound(rw, r)
			return
		}
		defer f.Close()
		http.ServeContent(rw, r, name, info.ModTime(), f)
	})
}
