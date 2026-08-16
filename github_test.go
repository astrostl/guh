package main

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestItemLabel(t *testing.T) {
	tests := []struct {
		name string
		item Item
		want string
	}{
		{
			name: "repo",
			item: Item{Kind: KindRepo, Repo: "astrostl/guh"},
			want: "astrostl/guh",
		},
		{
			name: "private fork",
			item: Item{Kind: KindRepo, Repo: "astrostl/guh", Private: true, Fork: true},
			want: "astrostl/guh  🔒 🍴",
		},
		{
			name: "issue",
			item: Item{Kind: KindIssue, Repo: "astrostl/guh", Number: 3, Title: "Fix the thing"},
			want: "astrostl/guh#3  Fix the thing",
		},
		{
			name: "pr no title",
			item: Item{Kind: KindPR, Repo: "astrostl/guh", Number: 9},
			want: "astrostl/guh#9",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.item.Label(); got != tt.want {
				t.Fatalf("Label() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestTypeEmoji(t *testing.T) {
	tests := []struct {
		name string
		item Item
		want string
	}{
		{
			name: "plain public repo",
			item: Item{Kind: KindRepo, Repo: "astrostl/guh"},
			want: "",
		},
		{
			name: "private repo",
			item: Item{Kind: KindRepo, Repo: "astrostl/guh", Private: true},
			want: "🔒",
		},
		{
			name: "fork repo",
			item: Item{Kind: KindRepo, Repo: "astrostl/guh", Fork: true},
			want: "🍴",
		},
		{
			name: "private fork repo",
			item: Item{Kind: KindRepo, Repo: "astrostl/guh", Private: true, Fork: true},
			want: "🔒🍴",
		},
		{
			name: "issue",
			item: Item{Kind: KindIssue, Number: 1},
			want: "",
		},
		{
			name: "pr",
			item: Item{Kind: KindPR, Number: 2},
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.item.TypeEmoji(); got != tt.want {
				t.Fatalf("TypeEmoji() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestChildLabel(t *testing.T) {
	got := Item{Kind: KindIssue, Number: 86, Title: "IntelliBrite Color Mode Change"}.ChildLabel()
	if got != "#86  IntelliBrite Color Mode Change" {
		t.Fatalf("got %q", got)
	}
}

func TestFormatDaysOffset(t *testing.T) {
	now := time.Date(2026, 8, 16, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		t    time.Time
		want string
	}{
		{time.Time{}, "-"},
		{now.Add(-2 * time.Hour), "0"},
		{now.Add(-23 * time.Hour), "0"},
		{now.Add(-25 * time.Hour), "-1"},
		{now.Add(-33 * 24 * time.Hour), "-33"},
		{now.Add(-100 * 24 * time.Hour), "-100"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := formatDaysOffset(tt.t, now); got != tt.want {
				t.Errorf("formatDaysOffset() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFormatLocalDateTime(t *testing.T) {
	if formatLocalDateTime(time.Time{}) != "-" {
		t.Fatal("expected dash for zero time")
	}
	const layout = "Mon Jan 2, 2006 3:04 PM MST"
	loc := time.FixedZone("CST", -6*60*60)
	t0 := time.Date(2026, 8, 15, 18, 30, 0, 0, time.UTC).In(loc)
	if got := t0.Format(layout); got != "Sat Aug 15, 2026 12:30 PM CST" {
		t.Fatalf("setup format = %q", got)
	}
	utc := time.Date(2026, 8, 15, 18, 30, 0, 0, time.UTC)
	want := utc.Local().Format(layout)
	if formatLocalDateTime(utc) != want {
		t.Fatalf("formatLocalDateTime = %q, want %q", formatLocalDateTime(utc), want)
	}
}

func TestFormatRelativeTime(t *testing.T) {
	now := time.Date(2026, 8, 16, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		t    time.Time
		want string
	}{
		{time.Time{}, "-"},
		{now.Add(-30 * time.Second), "just now"},
		{now.Add(-1 * time.Minute), "1m ago"},
		{now.Add(-15 * time.Minute), "15m ago"},
		{now.Add(-1 * time.Hour), "1h ago"},
		{now.Add(-4 * time.Hour), "4h ago"},
		{now.Add(-24 * time.Hour), "1d ago"},
		{now.Add(-5 * 24 * time.Hour), "5d ago"},
		{now.Add(-45 * 24 * time.Hour), "1mo ago"},
		{now.Add(-120 * 24 * time.Hour), "4mo ago"},
		{now.Add(-400 * 24 * time.Hour), "1y ago"},
		{now.Add(-800 * 24 * time.Hour), "2y ago"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := formatRelativeTime(tt.t, now); got != tt.want {
				t.Errorf("formatRelativeTime() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestLoadReportWith(t *testing.T) {
	run := func(_ context.Context, name string, args ...string) ([]byte, error) {
		if name != "gh" {
			t.Fatalf("unexpected command %s", name)
		}
		joined := strings.Join(args, " ")
		switch {
		case strings.HasPrefix(joined, "repo list"):
			return []byte(`[
				{"nameWithOwner":"astrostl/old","url":"https://github.com/astrostl/old","updatedAt":"2024-01-01T00:00:00Z"},
				{"nameWithOwner":"astrostl/new","url":"https://github.com/astrostl/new","isPrivate":true,"updatedAt":"2026-01-01T00:00:00Z","description":"fresh","stargazerCount":5,"primaryLanguage":{"name":"Go"}}
			]`), nil
		case strings.HasPrefix(joined, "api graphql"):
			return []byte(`{
				"data": {
					"viewer": {
						"repositories": {
							"pageInfo": {"hasNextPage": false, "endCursor": ""},
							"nodes": [
								{
									"nameWithOwner": "astrostl/new",
									"issues": {
										"totalCount": 2,
										"nodes": [
											{"number":2,"title":"Later","url":"https://github.com/astrostl/new/issues/2","updatedAt":"2026-02-01T00:00:00Z","author":{"login":"alice"}},
											{"number":1,"title":"Earlier","url":"https://github.com/astrostl/new/issues/1","updatedAt":"2026-01-01T00:00:00Z"}
										]
									},
									"pullRequests": {
										"totalCount": 1,
										"nodes": [
											{"number":7,"title":"Ship it","url":"https://github.com/astrostl/new/pull/7","updatedAt":"2026-03-01T00:00:00Z","author":{"login":"bob"}}
										]
									}
								}
							]
						}
					}
				}
			}`), nil
		default:
			return nil, fmt.Errorf("unexpected args %q", joined)
		}
	}

	report, err := loadReportWith(context.Background(), run, "")
	if err != nil {
		t.Fatal(err)
	}
	if len(report.Repos) != 2 || report.Repos[0].Repo != "astrostl/new" {
		t.Fatalf("repos = %+v", report.Repos)
	}
	if !report.Repos[0].Private || report.Repos[0].Description != "fresh" || report.Repos[0].Stars != 5 || report.Repos[0].Language != "Go" {
		t.Fatalf("repo fields = %+v", report.Repos[0])
	}
	if report.Repos[0].IssueCount != 2 || report.Repos[0].PRCount != 1 {
		t.Fatalf("repo counts = issues %d prs %d", report.Repos[0].IssueCount, report.Repos[0].PRCount)
	}
	if len(report.Issues) != 2 || report.Issues[0].Number != 2 || report.Issues[0].Author != "alice" {
		t.Fatalf("issues = %+v", report.Issues)
	}
	if len(report.PRs) != 1 || report.PRs[0].Title != "Ship it" || report.PRs[0].Author != "bob" {
		t.Fatalf("prs = %+v", report.PRs)
	}
}

func TestLoadReportWithOwner(t *testing.T) {
	var mu sync.Mutex
	var seen []string
	run := func(_ context.Context, name string, args ...string) ([]byte, error) {
		if name != "gh" {
			t.Fatalf("unexpected command %s", name)
		}
		joined := strings.Join(args, " ")
		mu.Lock()
		seen = append(seen, joined)
		mu.Unlock()
		if strings.HasPrefix(joined, "api graphql") {
			return []byte(`{"data":{"organization":{"repositories":{"pageInfo":{"hasNextPage":false},"nodes":[]}}}}`), nil
		}
		return []byte("[]"), nil
	}
	if _, err := loadReportWith(context.Background(), run, "acme"); err != nil {
		t.Fatal(err)
	}
	joined := strings.Join(seen, "\n")
	if !strings.Contains(joined, "repo list acme ") {
		t.Fatalf("expected org repo list, got %q", joined)
	}
	if !strings.Contains(joined, "api graphql") || !strings.Contains(joined, "login=acme") {
		t.Fatalf("expected org graphql activity, got %q", joined)
	}
}

func TestActivityPageUsesRepoTotals(t *testing.T) {
	run := func(_ context.Context, name string, args ...string) ([]byte, error) {
		if !strings.Contains(strings.Join(args, " "), "api graphql") {
			t.Fatalf("unexpected args %v", args)
		}
		return []byte(`{
			"data": {
				"organization": {
					"repositories": {
						"pageInfo": {"hasNextPage": false, "endCursor": "c1"},
						"nodes": [{
							"nameWithOwner": "acme/widgets",
							"issues": {"totalCount": 0, "nodes": []},
							"pullRequests": {
								"totalCount": 2,
								"nodes": [
									{"number":2,"title":"Later change","url":"https://github.com/acme/widgets/pull/2","updatedAt":"2026-03-01T00:00:00Z","author":{"login":"dev"}},
									{"number":1,"title":"Earlier change","url":"https://github.com/acme/widgets/pull/1","updatedAt":"2024-01-01T00:00:00Z","author":{"login":"dev"}}
								]
							}
						}]
					}
				}
			}
		}`), nil
	}
	page, err := fetchActivityPage(context.Background(), run, "acme", "")
	if err != nil {
		t.Fatal(err)
	}
	if len(page.Repos) != 1 || page.Repos[0].Repo != "acme/widgets" || page.Repos[0].PRCount != 2 {
		t.Fatalf("repo counts = %+v", page.Repos)
	}
	if len(page.PRs) != 2 || page.PRs[0].Number != 2 || page.PRs[1].Number != 1 {
		t.Fatalf("prs = %+v", page.PRs)
	}
}

func TestFetchOwnerChoices(t *testing.T) {
	run := func(_ context.Context, name string, args ...string) ([]byte, error) {
		joined := strings.Join(args, " ")
		switch {
		case strings.HasPrefix(joined, "api user"):
			return []byte("astrostl\n"), nil
		case strings.HasPrefix(joined, "org list"):
			return []byte("acme\nastrostl\nwidgets-inc\nShowing 2 organizations\n"), nil
		default:
			return nil, fmt.Errorf("unexpected args %q", joined)
		}
	}
	login, orgs, err := fetchOwnerChoices(context.Background(), run)
	if err != nil {
		t.Fatal(err)
	}
	if login != "astrostl" {
		t.Fatalf("login = %q, want astrostl", login)
	}
	if len(orgs) != 2 || orgs[0] != "acme" || orgs[1] != "widgets-inc" {
		t.Fatalf("orgs = %v", orgs)
	}
}

func TestFetchRecentCommits(t *testing.T) {
	run := func(_ context.Context, name string, args ...string) ([]byte, error) {
		joined := strings.Join(args, " ")
		if !strings.Contains(joined, "repos/astrostl/guh/commits?per_page=7") {
			t.Fatalf("unexpected args %q", joined)
		}
		return []byte(`[
			{"sha":"abcdef1234567890","html_url":"https://github.com/astrostl/guh/commit/abcdef1234567890","commit":{"message":"Fix the thing\n\nDetails.","author":{"name":"Ada","date":"2026-08-15T12:00:00Z"}}},
			{"sha":"1234567deadbeef","html_url":"https://github.com/astrostl/guh/commit/1234567deadbeef","commit":{"message":"Start project","author":{"name":"Bob","date":"2026-08-01T00:00:00Z"}}}
		]`), nil
	}
	got, err := fetchRecentCommits(context.Background(), run, "astrostl/guh", 7)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 {
		t.Fatalf("len = %d, want 2", len(got))
	}
	if got[0].ShortSHA() != "abcdef1" || got[0].Title != "Fix the thing" || got[0].Author != "Ada" {
		t.Fatalf("commit 0 = %+v", got[0])
	}
	if got[0].URL != "https://github.com/astrostl/guh/commit/abcdef1234567890" {
		t.Fatalf("url = %q", got[0].URL)
	}
	if got[1].Title != "Start project" || got[1].Author != "Bob" {
		t.Fatalf("commit 1 = %+v", got[1])
	}
}

func TestParseTime(t *testing.T) {
	got := parseTime("2026-01-02T03:04:05Z")
	want := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	if !got.Equal(want) {
		t.Fatalf("parseTime = %v, want %v", got, want)
	}
	if !parseTime("").IsZero() || !parseTime("nope").IsZero() {
		t.Fatal("expected zero time for invalid input")
	}
}
