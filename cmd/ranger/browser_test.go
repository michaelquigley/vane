package main

import (
	"reflect"
	"strings"
	"testing"
)

// the windows branch never runs under a green linux build, so the argv
// construction is pinned per platform here.
func TestBrowserArgv(t *testing.T) {
	linux, err := browserArgv("linux", "http://127.0.0.1:4114")
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(linux, []string{"xdg-open", "http://127.0.0.1:4114"}) {
		t.Errorf("linux argv = %v", linux)
	}

	windows, err := browserArgv("windows", "http://127.0.0.1:4114")
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(windows, []string{"rundll32.exe", "url.dll,FileProtocolHandler", "http://127.0.0.1:4114"}) {
		t.Errorf("windows argv = %v", windows)
	}

	if _, err := browserArgv("darwin", "http://127.0.0.1:4114"); err == nil || !strings.Contains(err.Error(), "http://127.0.0.1:4114") {
		t.Errorf("unsupported platform must refuse naming the url, got %v", err)
	}
}
