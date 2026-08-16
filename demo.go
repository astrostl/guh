package main

import (
	"fmt"
	"strings"
	"time"
)

const demoLogin = "mira"

var demoNow = time.Date(2026, 8, 16, 12, 30, 0, 0, time.FixedZone("CDT", -5*3600))

var demoOrgs = []string{"lantern", "northwind", "rivulet"}

func demoAgo(days, hours int) time.Time {
	return demoNow.Add(-time.Duration(days)*24*time.Hour - time.Duration(hours)*time.Hour)
}

func demoOwners() (login string, orgs []string) {
	return demoLogin, append([]string{}, demoOrgs...)
}

func demoReport(owner string) Report {
	owner = strings.TrimSpace(owner)
	if owner == "" || strings.EqualFold(owner, demoLogin) {
		return demoPersonal()
	}
	switch strings.ToLower(owner) {
	case "lantern":
		return demoLantern()
	case "northwind":
		return demoNorthwind()
	case "rivulet":
		return demoRivulet()
	default:
		return Report{}
	}
}

func demoPersonal() Report {
	return assemble(
		dRepo("mira/cartograph", 0, 3, 12, false, false, false, "Go", "Offline maps from OSM extracts",
			[]Item{
				dIssue("mira/cartograph", 12, "Render tiles at 2x without blur", "priya", 0, 5),
				dIssue("mira/cartograph", 9, "Missing contour labels above 60N", "jules", 1, 2),
				dIssue("mira/cartograph", 7, "GPX import drops timezone", "mira", 2, 4),
			},
			[]Item{
				dPR("mira/cartograph", 15, "Cache DEM tiles on disk", "jules", 0, 2),
			},
		),
		dRepo("mira/letters", 0, 18, 0, true, false, false, "Swift", "A small letterpress tracker",
			[]Item{
				dIssue("mira/letters", 4, "Press hours wrap past midnight", "mira", 0, 20),
			},
			nil,
		),
		dRepo("mira/lumen", 1, 4, 184, false, false, false, "TypeScript", "Small notes, local-first",
			[]Item{
				dIssue("mira/lumen", 41, "Remember last opened note on launch", "priya", 0, 8),
				dIssue("mira/lumen", 38, "Sync conflict when two devices edit offline", "jules", 1, 6),
				dIssue("mira/lumen", 36, "Search ignores diacritics", "mira", 2, 1),
				dIssue("mira/lumen", 29, "Export to markdown strips images", "nico", 4, 3),
			},
			[]Item{
				dPR("mira/lumen", 44, "Draft: sqlite FTS tokenizer", "jules", 0, 6),
				dPR("mira/lumen", 42, "Collapse empty notebooks in sidebar", "mira", 1, 2),
			},
		),
		dRepo("mira/timecard", 1, 9, 2, true, false, false, "Go", "Hours, nothing else",
			nil, nil,
		),
		dRepo("mira/weather-radar", 3, 2, 41, false, false, false, "Rust", "NEXRAD overlays for a paper map",
			[]Item{
				dIssue("mira/weather-radar", 18, "Loop stutters after 50 frames", "owen", 2, 8),
				dIssue("mira/weather-radar", 14, "Color scale unreadable at night", "mira", 3, 1),
			},
			[]Item{
				dPR("mira/weather-radar", 21, "Decode Level III without the SDK", "kate", 1, 5),
			},
		),
		dRepo("mira/workshop", 4, 6, 8, false, false, false, "Go", "One-off tools that earned a home",
			nil,
			[]Item{
				dPR("mira/workshop", 6, "Add a csv-to-sqlite helper", "nico", 3, 4),
				dPR("mira/workshop", 5, "Quiet the JSON pretty-printer", "mira", 4, 7),
			},
		),
		dRepo("mira/dotfiles", 5, 1, 6, false, false, false, "Shell", "",
			nil, nil,
		),
		dRepo("mira/inkwell", 6, 11, 22, false, false, false, "Python", "A pen plotter driver that stays out of the way",
			[]Item{
				dIssue("mira/inkwell", 11, "Hatch fill overshoots on closed paths", "priya", 5, 2),
			},
			nil,
		),
		dRepo("mira/sketchbook", 8, 0, 3, false, true, false, "TypeScript", "Fork of atelier/sketchbook",
			nil, nil,
		),
		dRepo("mira/recipes", 12, 3, 0, true, true, false, "Markdown", "Family recipes, private fork",
			nil, nil,
		),
		dRepo("mira/harbor", 18, 7, 67, false, false, false, "Go", "A tiny static site that feels like a desk",
			[]Item{
				dIssue("mira/harbor", 8, "Drafts leak into the RSS feed", "mira", 10, 4),
			},
			nil,
		),
		dRepo("mira/old-notes", 90, 0, 1, false, false, true, "Markdown", "Pre-lumen notes. Leave them.",
			nil, nil,
		),
	)
}

func demoLantern() Report {
	return assemble(
		dRepo("lantern/beacon", 0, 6, 31, false, false, false, "Go", "Status page that does not page you",
			[]Item{
				dIssue("lantern/beacon", 22, "Mute window ignores weekends", "owen", 0, 9),
				dIssue("lantern/beacon", 19, "Incident timeline jumps a day", "mira", 1, 3),
			},
			[]Item{
				dPR("lantern/beacon", 24, "Store mute rules as a list, not a blob", "jules", 0, 4),
			},
		),
		dRepo("lantern/keep", 2, 5, 14, false, false, false, "TypeScript", "Shared bookmarks, no accounts",
			[]Item{
				dIssue("lantern/keep", 7, "Tags collide after a rename", "priya", 2, 2),
			},
			nil,
		),
		dRepo("lantern/handbook", 4, 0, 0, true, false, false, "Markdown", "How we do things",
			[]Item{
				dIssue("lantern/handbook", 3, "On-call page is a year out of date", "kate", 3, 6),
			},
			nil,
		),
		dRepo("lantern/site", 9, 8, 5, false, false, false, "HTML", "lantern.example",
			nil,
			[]Item{
				dPR("lantern/site", 12, "Drop the stock hero image", "mira", 8, 1),
			},
		),
	)
}

func demoNorthwind() Report {
	return assemble(
		dRepo("northwind/ledger", 0, 14, 9, false, false, false, "Go", "Invoices that print on one page",
			[]Item{
				dIssue("northwind/ledger", 31, "Tax line rounds the wrong way in JPY", "nico", 0, 16),
				dIssue("northwind/ledger", 28, "PDF footer clips on A4", "mira", 1, 8),
			},
			[]Item{
				dPR("northwind/ledger", 33, "Let a line item span two pages", "owen", 0, 11),
			},
		),
		dRepo("northwind/catalog", 3, 9, 4, false, false, false, "TypeScript", "What is in the warehouse",
			[]Item{
				dIssue("northwind/catalog", 16, "SKU search is case-sensitive", "kate", 3, 4),
			},
			nil,
		),
		dRepo("northwind/ops-scripts", 7, 2, 0, true, false, false, "Shell", "The glue",
			nil, nil,
		),
	)
}

func demoRivulet() Report {
	return assemble(
		dRepo("rivulet/stream", 1, 1, 53, false, false, false, "Rust", "Append-only logs over a cheap VPS",
			[]Item{
				dIssue("rivulet/stream", 40, "Compaction holds the write lock too long", "jules", 0, 22),
				dIssue("rivulet/stream", 37, "Document the retention flag", "mira", 2, 0),
			},
			[]Item{
				dPR("rivulet/stream", 41, "Background compaction", "jules", 0, 3),
			},
		),
		dRepo("rivulet/weir", 6, 4, 11, false, false, false, "Go", "A small dam in front of stream",
			nil,
			[]Item{
				dPR("rivulet/weir", 8, "Rate-limit by token, not IP", "priya", 5, 6),
			},
		),
	)
}

func assemble(parts ...struct {
	repo   Item
	issues []Item
	prs    []Item
}) Report {
	var r Report
	for _, p := range parts {
		p.repo.IssueCount = len(p.issues)
		p.repo.PRCount = len(p.prs)
		r.Repos = append(r.Repos, p.repo)
		r.Issues = append(r.Issues, p.issues...)
		r.PRs = append(r.PRs, p.prs...)
	}
	return r
}

func dRepo(name string, days, hours, stars int, priv, fork, archived bool, lang, desc string, issues, prs []Item) struct {
	repo   Item
	issues []Item
	prs    []Item
} {
	return struct {
		repo   Item
		issues []Item
		prs    []Item
	}{
		repo: Item{
			Kind:        KindRepo,
			Repo:        name,
			Title:       name,
			URL:         "https://github.com/" + name,
			UpdatedAt:   demoAgo(days, hours),
			Private:     priv,
			Fork:        fork,
			Archived:    archived,
			Description: desc,
			Stars:       stars,
			Language:    lang,
		},
		issues: issues,
		prs:    prs,
	}
}

func dIssue(repo string, n int, title, author string, days, hours int) Item {
	return Item{
		Kind:      KindIssue,
		Repo:      repo,
		Number:    n,
		Title:     title,
		URL:       fmt.Sprintf("https://github.com/%s/issues/%d", repo, n),
		UpdatedAt: demoAgo(days, hours),
		Author:    author,
	}
}

func dPR(repo string, n int, title, author string, days, hours int) Item {
	return Item{
		Kind:      KindPR,
		Repo:      repo,
		Number:    n,
		Title:     title,
		URL:       fmt.Sprintf("https://github.com/%s/pull/%d", repo, n),
		UpdatedAt: demoAgo(days, hours),
		Author:    author,
	}
}

func demoCommits(repo string) []Commit {
	if repo == "" {
		return nil
	}
	if c, ok := demoCommitBank[repo]; ok {
		return c
	}
	return []Commit{
		dCommit(repo, "4e8c1a2b9f0d3c7e5a1b8d4c6f2e0a9b7c3d5e1f", "Touch the readme", "mira", 0, 4),
		dCommit(repo, "9b2f7d1c4a8e6b0d3f5c7a1e9b4d2c8f0a6e3b5d", "Fix a typo in the usage text", "mira", 1, 2),
		dCommit(repo, "1c5a8e3b7d0f2a4c6e8b1d3f5a7c9e0b2d4f6a8c", "Quiet a warning on empty input", "jules", 3, 6),
		dCommit(repo, "7d4b0e2c9a1f5d8b3c6e0a4f2d7b9c1e5a8f3d6b", "Pin the toolchain", "nico", 6, 1),
		dCommit(repo, "2a9f6c1e4b8d0a3f5c7e9b2d4a6c8e0f1b3d5a7c", "Drop an unused helper", "mira", 9, 8),
		dCommit(repo, "8c3e1b7a5d0f4c2e9a6b8d1f3c5e7a0b2d4f6c8e", "Add a smoke test", "priya", 14, 3),
		dCommit(repo, "5f1d8a3c6e0b4d7a9c2e5b8f1d4a7c0e3b6f9a2d", "Initial commit", "mira", 40, 0),
	}
}

func dCommit(repo, sha, title, author string, days, hours int) Commit {
	when := demoAgo(days, hours)
	return Commit{
		SHA:    sha,
		Title:  title,
		Author: author,
		Date:   when,
		URL:    fmt.Sprintf("https://github.com/%s/commit/%s", repo, sha),
	}
}

var demoCommitBank = map[string][]Commit{
	"mira/lumen": {
		dCommit("mira/lumen", "c7e2a91b4d5f803e1a6c9b2d4e7f0a3c5b8d1e6f", "Remember the last notebook in local storage", "mira", 0, 3),
		dCommit("mira/lumen", "a1d4f8c2e6b0a3d5f7c9e1b4d6a8c0e2f4b6d8a0", "FTS: fold diacritics before indexing", "jules", 0, 8),
		dCommit("mira/lumen", "b8c3e0a5d1f7a2c4e6b8d0a3f5c7e9b1d4a6c8e2", "Don't export the clipboard scratch pad", "priya", 1, 5),
		dCommit("mira/lumen", "d2f6a0c4e8b1d5f7a3c9e0b2d6a8c1e4f7b0d3a5", "Sidebar: hide empty stacks", "mira", 2, 2),
		dCommit("mira/lumen", "e5a9c1d3f7b0a4c6e8d2b5f1a7c0e3d6b9a2c4e8", "Conflict: keep both notes, mark the loser", "jules", 4, 7),
		dCommit("mira/lumen", "f0b4d8a2c6e1f5a7c3e9b0d4a8c2e6f1b5d7a3c9", "Drop the leftover electron preload", "nico", 8, 4),
		dCommit("mira/lumen", "91c5e2a7d0b4f8c1e6a3d9b5f2c7e0a4d8b1f6c3", "Notes that stay on this machine", "mira", 60, 0),
	},
	"mira/cartograph": {
		dCommit("mira/cartograph", "3a7c1e5b9d2f6a0c4e8b1d5f7a3c9e0b4d6f2a8c", "Cache DEM tiles next to the OSM extract", "jules", 0, 2),
		dCommit("mira/cartograph", "6d0f4a8c2e7b1d5a9c3e6b0f4d8a2c5e9b1d7f3a", "2x tiles: pick the @2x pack when present", "mira", 0, 6),
		dCommit("mira/cartograph", "8e2b6d0a4c9f1e5a7c3b8d2f6a0c4e9b1d5f7a3c", "GPX: keep the original timezone offset", "priya", 1, 4),
		dCommit("mira/cartograph", "1f5c9a3e7b0d4a8c2e6b1d5f9a3c7e0b4d8f2a6c", "Contours: skip labels that collide", "jules", 2, 9),
		dCommit("mira/cartograph", "4b8d2f6a0c5e9b1d3f7a2c6e0b4d8f1a5c9e3b7d", "Trim the coastline shapefile", "mira", 5, 1),
		dCommit("mira/cartograph", "7c1e5a9b3d6f0c4a8e2b5d9f1c7a3e6b0d4f8a2c", "Render a single extract without a server", "mira", 20, 3),
		dCommit("mira/cartograph", "0a4c8e2b6d1f5a9c3e7b0d4a8c2e6f1b5d9a3c7e", "Paper maps, offline", "mira", 80, 0),
	},
	"lantern/beacon": {
		dCommit("lantern/beacon", "2c6e0a4d8b1f5c9a3e7b0d4f8a2c6e1b5d9f3a7c", "Mute rules as a list", "jules", 0, 4),
		dCommit("lantern/beacon", "5f9a3c7e1b4d8a0c6e2b5d9f3a7c1e4b8d0f6a2c", "Weekend mute: Saturday 00:00 to Monday 00:00", "owen", 1, 1),
		dCommit("lantern/beacon", "8a2d6f0c4e9b1a5c7e3b8d2f6a0c4e9b1d5f7a3c", "Timeline: pin the day in the local zone", "mira", 2, 8),
		dCommit("lantern/beacon", "1d5f9a3c7e0b4d8a2c6e1b5d9f3a7c0e4b8d2f6a", "Don't page for a blip under 90s", "kate", 4, 2),
		dCommit("lantern/beacon", "4e8b2d6f0a3c7e1b5d9a2c6e0b4d8f1a5c9e3b7d", "Strip the vendor status iframe", "mira", 9, 6),
		dCommit("lantern/beacon", "7b1d5f9a3c6e0b4d8a2c5e9b1d7f3a6c0e4b8d2f", "A status page that does not page you", "mira", 45, 0),
		dCommit("lantern/beacon", "0c4e8a2d6f1b5a9c3e7b0d4f8a2c6e9b1d5f7a3c", "Initial commit", "mira", 100, 0),
	},
}
