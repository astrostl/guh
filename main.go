package main

import (
	"context"
	"fmt"
	"os"
	"runtime/debug"
	"text/tabwriter"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"golang.org/x/term"
)

var version = "dev"

func init() {
	if version != "dev" {
		return
	}
	if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "" && info.Main.Version != "(devel)" {
		version = info.Main.Version
	}
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "guh: %v\n", err)
		os.Exit(1)
	}
}

func hasFlag(args []string, name string) bool {
	one, two := "-"+name, "--"+name
	for _, a := range args {
		if a == one || a == two {
			return true
		}
	}
	return false
}

func run() error {
	if hasFlag(os.Args[1:], "version") {
		fmt.Printf("guh %s\n", version)
		return nil
	}

	demo := hasFlag(os.Args[1:], "demo")

	if !term.IsTerminal(int(os.Stdout.Fd())) {
		if demo {
			printPlain(demoReport(""))
			return nil
		}
		ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
		defer cancel()
		report, err := loadReport(ctx, "")
		if err != nil {
			return err
		}
		printPlain(report)
		return nil
	}

	forceTrueColor()
	p := tea.NewProgram(newModelWith(demo), tea.WithAltScreen())
	_, err := p.Run()
	return err
}

func printPlain(report Report) {
	now := time.Now()
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "TYPE\tREPOSITORY\tCOMMITS\tISSUES\tPRS\tSTARS\tUPDATE\tURL")

	issueCounts := make(map[string]int)
	prCounts := make(map[string]int)
	for _, it := range report.Issues {
		issueCounts[it.Repo]++
	}
	for _, it := range report.PRs {
		prCounts[it.Repo]++
	}

	for _, repo := range report.Repos {
		updStr := formatDaysOffset(repo.UpdatedAt, now)
		fmt.Fprintf(w, "%s\t%s\t%d\t%d\t%d\t%d\t%s\t%s\n",
			repo.TypeEmoji(),
			repo.Repo,
			repo.CommitCount,
			issueCounts[repo.Repo],
			prCounts[repo.Repo],
			repo.Stars,
			updStr,
			repo.URL,
		)
	}
	w.Flush()

	if len(report.Issues) > 0 {
		fmt.Printf("\nOPEN ISSUES (%d):\n", len(report.Issues))
		for _, iss := range report.Issues {
			author := ""
			if iss.Author != "" {
				author = fmt.Sprintf("by @%s", iss.Author)
			}
			updStr := formatDaysOffset(iss.UpdatedAt, now)
			fmt.Printf("  %s %-30s #%-4d %-15s %-6s %s\t%s\n", iss.TypeEmoji(), iss.Repo, iss.Number, author, updStr, iss.Title, iss.URL)
		}
	}
	if len(report.PRs) > 0 {
		fmt.Printf("\nOPEN PULL REQUESTS (%d):\n", len(report.PRs))
		for _, pr := range report.PRs {
			author := ""
			if pr.Author != "" {
				author = fmt.Sprintf("by @%s", pr.Author)
			}
			updStr := formatDaysOffset(pr.UpdatedAt, now)
			fmt.Printf("  %s %-30s #%-4d %-15s %-6s %s\t%s\n", pr.TypeEmoji(), pr.Repo, pr.Number, author, updStr, pr.Title, pr.URL)
		}
	}
}
