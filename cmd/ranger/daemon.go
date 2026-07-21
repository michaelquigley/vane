package main

import (
	_ "embed"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/michaelquigley/df/dl"
	"github.com/michaelquigley/dfw/tray"
	"github.com/spf13/cobra"
	"github.com/michaelquigley/ranger/internal/config"
	"github.com/michaelquigley/ranger/internal/server"
)

// the binoculars mark, rendered from favicon.svg for the tray.
//
//go:embed tray-icon.png
var trayIconPNG []byte

// newDaemonCmd runs ranger as the tray-resident daemon: every configured
// root served from one process, one gesture from any board. fail-fast
// covers only the daemon's own preconditions — an unreadable config, an
// unbindable port; project roots are not bootstrap, and a broken one
// degrades per request instead of keeping the daemon down.
func newDaemonCmd() *cobra.Command {
	var portFlag int
	cmd := &cobra.Command{
		Use:   "daemon",
		Short: "run the tray-resident daemon over the configured projects",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			path, err := daemonConfigPath()
			if err != nil {
				return err
			}
			cfg, mux, err := daemonBootstrap(path)
			if err != nil {
				return err
			}
			port := resolvePort(cfg, portFlag, cmd.Flags().Changed("port"))

			// the board URL comes from the bound listener's address, never
			// from a config re-read: the port is bootstrap-fixed (the one
			// carve-out from the fresh-config promise), so a later port:
			// edit must not change where the tray points.
			var boardURL string
			listen := func() (*http.Server, net.Listener, error) {
				addr := net.JoinHostPort("127.0.0.1", strconv.Itoa(port))
				listener, err := net.Listen("tcp", addr)
				if err != nil {
					return nil, nil, err
				}
				boardURL = "http://" + listener.Addr().String()
				dl.Infof("daemon serving %d projects at %s", len(cfg.Projects), boardURL)
				return &http.Server{Handler: mux}, listener, nil
			}

			// SpawnWindow stays unset — it is dfw's webview-window hook and
			// ranger opens no webview, ever; the browser is the window
			// manager. the menu is open board plus dfw's built-in quit.
			return tray.Daemon(tray.DaemonApp{
				AppID:   "com.michaelquigley.ranger",
				Title:   "ranger",
				IconPNG: trayIconPNG,
				Listen:  listen,
				TrayItems: []tray.TrayMenuItem{
					{Label: "open board", Tooltip: "open the board in your browser", OnClick: func() {
						if err := openBrowser(boardURL); err != nil {
							dl.Errorf("open board: %v", err)
						}
					}},
				},
			})
		},
	}
	cmd.Flags().IntVar(&portFlag, "port", config.DefaultPort, "listen port on 127.0.0.1 (overrides the config)")
	return cmd
}

// resolvePort applies the daemon's port precedence: the config's port
// unless --port was explicitly supplied — cobra's Changed, never a
// zero-value check, so an explicit --port 4114 still wins over a config
// naming another port.
func resolvePort(cfg *config.Config, flagPort int, flagChanged bool) int {
	if flagChanged {
		return flagPort
	}
	return cfg.Port
}

// daemonConfigPath is ~/.config/ranger/config.yaml — the hand-edited
// project set; the operator's editor is the config surface.
func daemonConfigPath() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, "ranger", "config.yaml"), nil
}

// daemonBootstrap validates the config once — fail-fast, naming the path
// and the fix — then assembles the shared mux over a file-backed source
// that re-reads per request, so an edit takes effect on the next request
// and a mid-flight parse failure surfaces as that request's plain error.
// deliberately no per-root load gate anywhere: degradation is the point.
func daemonBootstrap(path string) (*config.Config, http.Handler, error) {
	cfg, err := config.Load(path)
	if err != nil {
		// name the fix the failure actually has: a missing file needs
		// creating; every other load error already names its own repair.
		if _, statErr := os.Stat(path); os.IsNotExist(statErr) {
			return nil, nil, fmt.Errorf("%v\nthe daemon starts only on a valid config: create %s naming at least one project root", err, path)
		}
		return nil, nil, fmt.Errorf("%v\nthe daemon starts only on a valid config: fix %s and restart", err, path)
	}
	mux, err := buildMux(server.NewProjects(func() (*config.Config, error) { return config.Load(path) }))
	if err != nil {
		return nil, nil, err
	}
	return cfg, mux, nil
}
