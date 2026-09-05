package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

func openURL(rawURL string) error {
	if rawURL == "" {
		return fmt.Errorf("empty url")
	}
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", rawURL)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", rawURL)
	default:
		cmd = exec.Command("xdg-open", rawURL)
	}
	cmd.Stdout = nil
	cmd.Stderr = nil
	return cmd.Start()
}

func defaultSrcReposDir() string {
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return ""
	}
	return filepath.Join(home, "src")
}

var srcReposDir = defaultSrcReposDir

func repoShortName(repo string) string {
	repo = strings.TrimSpace(repo)
	if repo == "" {
		return ""
	}
	name := repo
	if i := strings.LastIndex(repo, "/"); i >= 0 {
		name = repo[i+1:]
	}
	if name == "" || name == "." || name == ".." {
		return ""
	}
	if strings.ContainsAny(name, `/\`) {
		return ""
	}
	return name
}

func localRepoDir(repo string) (string, bool) {
	name := repoShortName(repo)
	root := srcReposDir()
	if name == "" || root == "" {
		return "", false
	}
	dir := filepath.Join(root, name)
	st, err := os.Stat(dir)
	if err != nil || !st.IsDir() {
		return "", false
	}
	return dir, true
}

func defaultShell() (string, []string) {
	if runtime.GOOS == "windows" {
		if s := os.Getenv("COMSPEC"); s != "" {
			return s, nil
		}
		return "cmd.exe", nil
	}
	if s := os.Getenv("SHELL"); s != "" {
		return s, nil
	}
	return "/bin/sh", nil
}

type shellDoneMsg struct {
	dir string
	err error
}

func (msg shellDoneMsg) statusError() error {
	if msg.err == nil {
		return nil
	}
	if _, isExit := msg.err.(*exec.ExitError); isExit {
		return nil
	}
	return msg.err
}

func localShellCmd(dir string) *exec.Cmd {
	shell, args := defaultShell()
	cmd := exec.Command(shell, args...)
	cmd.Dir = dir
	return cmd
}

func execLocalShell(dir string) tea.Cmd {
	return tea.ExecProcess(localShellCmd(dir), func(err error) tea.Msg {
		return shellDoneMsg{dir: dir, err: err}
	})
}

var startLocalShell = execLocalShell
