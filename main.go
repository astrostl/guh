package main

import (
	"context"
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"golang.org/x/term"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "guh: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	if !term.IsTerminal(int(os.Stdout.Fd())) {
		ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
		defer cancel()
		report, err := loadReport(ctx, "")
		if err != nil {
			return err
		}
		printPlain(report)
		return nil
	}

	p := tea.NewProgram(newModel(), tea.WithAltScreen())
	_, err := p.Run()
	return err
}

func printPlain(report Report) {
	now := time.Now()
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "TYPE\tREPOSITORY\tISSUES\tPRS\tSTARS\tUPDATE\tURL")

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
		fmt.Fprintf(w, "%s\t%s\t%d\t%d\t%d\t%s\t%s\n",
			repo.TypeEmoji(),
			repo.Repo,
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
