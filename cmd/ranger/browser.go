package main

import (
	"fmt"
	"os/exec"
	"runtime"
)

// browserArgv builds the platform's URL-opener invocation. on windows the
// entry point is part of the argv — url.dll alone does nothing.
func browserArgv(goos, url string) ([]string, error) {
	switch goos {
	case "linux":
		return []string{"xdg-open", url}, nil
	case "windows":
		return []string{"rundll32.exe", "url.dll,FileProtocolHandler", url}, nil
	default:
		return nil, fmt.Errorf("no browser opener for %s; browse to %s by hand", goos, url)
	}
}

// openBrowser fires the operator's default browser at url and gets out of
// the way — opening is not window management; the browser does the rest.
func openBrowser(url string) error {
	argv, err := browserArgv(runtime.GOOS, url)
	if err != nil {
		return err
	}
	cmd := exec.Command(argv[0], argv[1:]...)
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("%s: %w", argv[0], err)
	}
	// reap the opener so a long-lived daemon accumulates no zombies
	go func() { _ = cmd.Wait() }()
	return nil
}
