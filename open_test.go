package main

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestRepoShortName(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{"astrostl/foo-bar", "foo-bar"},
		{"foo-bar", "foo-bar"},
		{"  astrostl/foo-bar  ", "foo-bar"},
		{"", ""},
		{"astrostl/", ""},
		{".", ""},
		{"..", ""},
		{"owner/..", ""},
	}
	for _, tt := range tests {
		if got := repoShortName(tt.in); got != tt.want {
			t.Errorf("repoShortName(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestLocalRepoDir(t *testing.T) {
	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, "foo-bar"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "not-a-dir"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	orig := srcReposDir
	srcReposDir = func() string { return root }
	t.Cleanup(func() { srcReposDir = orig })

	dir, ok := localRepoDir("astrostl/foo-bar")
	if !ok || dir != filepath.Join(root, "foo-bar") {
		t.Fatalf("foo-bar: dir=%q ok=%v", dir, ok)
	}

	if _, ok := localRepoDir("astrostl/missing"); ok {
		t.Fatal("expected missing repo to fail")
	}
	if _, ok := localRepoDir("astrostl/not-a-dir"); ok {
		t.Fatal("expected file to fail")
	}
	if _, ok := localRepoDir(""); ok {
		t.Fatal("expected empty repo to fail")
	}
}

func TestShellDoneStatusError(t *testing.T) {
	if err := (shellDoneMsg{}).statusError(); err != nil {
		t.Fatalf("nil err: %v", err)
	}
	if err := (shellDoneMsg{err: &exec.ExitError{}}).statusError(); err != nil {
		t.Fatalf("exit error should be ignored: %v", err)
	}
	want := errors.New("boom")
	if err := (shellDoneMsg{err: want}).statusError(); err != want {
		t.Fatalf("got %v, want boom", err)
	}
}
