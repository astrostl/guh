package main

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

func sampleReport() Report {
	t1 := time.Date(2026, 8, 15, 10, 0, 0, 0, time.UTC)
	t2 := time.Date(2026, 8, 10, 10, 0, 0, 0, time.UTC)
	return Report{
		Repos: []Item{
			{Kind: KindRepo, Repo: "astrostl/a", URL: "https://github.com/astrostl/a", UpdatedAt: t1, Description: "Alpha repo", Stars: 3, Private: true},
			{Kind: KindRepo, Repo: "astrostl/b", URL: "https://github.com/astrostl/b", UpdatedAt: t2, Description: "Beta repo", Stars: 10, Fork: true},
		},
		Issues: []Item{
			{Kind: KindIssue, Repo: "astrostl/a", Number: 1, Title: "first bug", URL: "https://github.com/astrostl/a/issues/1", UpdatedAt: t1, Author: "alice"},
		},
		PRs: []Item{
			{Kind: KindPR, Repo: "astrostl/b", Number: 2, Title: "cool feature", URL: "https://github.com/astrostl/b/pull/2", UpdatedAt: t2, Author: "bob"},
		},
	}
}

func TestBuildGroupsAndRows(t *testing.T) {
	rep := sampleReport()
	groups := buildGroups(rep)
	if len(groups) != 2 {
		t.Fatalf("len(groups) = %d, want 2", len(groups))
	}
	if groups[0].repo.Repo != "astrostl/a" || len(groups[0].issues) != 1 || len(groups[0].prs) != 0 {
		t.Fatalf("group a = %+v", groups[0])
	}
	if groups[1].repo.Repo != "astrostl/b" || len(groups[1].issues) != 0 || len(groups[1].prs) != 1 {
		t.Fatalf("group b = %+v", groups[1])
	}

	m := newModel()
	m.report = rep
	m.groups = groups
	m.setAllExpanded(true)

	// 2 repos + 1 issue + 1 pr = 4 rows when expanded
	if len(m.rows) != 4 {
		t.Fatalf("len(rows) = %d, want 4: %+v", len(m.rows), m.rows)
	}
	if m.rows[0].typ != rowRepo || m.rows[0].item.Repo != "astrostl/a" || m.rows[0].issuesCount != 1 {
		t.Fatalf("row 0 = %+v", m.rows[0])
	}
	if m.rows[1].typ != rowIssue || m.rows[1].item.Number != 1 {
		t.Fatalf("row 1 = %+v", m.rows[1])
	}
	if m.rows[2].typ != rowRepo || m.rows[2].item.Repo != "astrostl/b" || m.rows[2].prsCount != 1 {
		t.Fatalf("row 2 = %+v", m.rows[2])
	}

	// Collapse all
	m.setAllExpanded(false)
	if len(m.rows) != 2 {
		t.Fatalf("len(rows) = %d, want 2 when collapsed", len(m.rows))
	}
}

func TestNavigationAndExpandToggle(t *testing.T) {
	m := newModel()
	m.loading = false
	m.report = sampleReport()
	m.groups = buildGroups(m.report)
	m.setAllExpanded(true)
	m.cursor = 0
	m.height = 24
	m.width = 100

	// rows: [0: repo a, 1: issue a#1, 2: repo b, 3: pr b#2]
	if len(m.rows) != 4 {
		t.Fatalf("expected 4 rows, got %d", len(m.rows))
	}

	// Move down
	m.move(1)
	if m.cursor != 1 || m.selected().Number != 1 {
		t.Fatalf("expected cursor at issue #1, got cursor=%d item=%+v", m.cursor, m.selected())
	}

	// Move down to repo b
	m.move(1)
	if m.cursor != 2 || m.selected().Repo != "astrostl/b" || m.rows[m.cursor].typ != rowRepo {
		t.Fatalf("expected cursor at repo b, got cursor=%d item=%+v", m.cursor, m.selected())
	}
}

func TestLeftRightFolding(t *testing.T) {
	m := newModel()
	m.loading = false
	m.report = sampleReport()
	m.groups = buildGroups(m.report)
	m.setAllExpanded(true)
	m.cursor = 0
	m.height = 24
	m.width = 100

	// rows: [0: repo a, 1: issue a#1, 2: repo b, 3: pr b#2]
	if len(m.rows) != 4 {
		t.Fatalf("expected 4 rows, got %d", len(m.rows))
	}

	// Left folds only the active repo (a)
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyLeft})
	m = updated.(model)
	if len(m.rows) != 3 {
		t.Fatalf("expected 3 rows after folding repo a, got %d", len(m.rows))
	}
	if m.expandedRepos["astrostl/a"] {
		t.Fatal("expected repo a folded")
	}
	if !m.expandedRepos["astrostl/b"] {
		t.Fatal("expected repo b to stay expanded")
	}
	if m.cursor != 0 || m.rows[m.cursor].item.Repo != "astrostl/a" {
		t.Fatalf("expected cursor on folded repo a, got cursor=%d item=%+v", m.cursor, m.selected())
	}

	// Right unfolds the active repo
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRight})
	m = updated.(model)
	if len(m.rows) != 4 {
		t.Fatalf("expected 4 rows after unfolding repo a, got %d", len(m.rows))
	}

	// Fold from a child: cursor on issue a#1, left folds a and snaps to repo a
	m.cursor = 1
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyLeft})
	m = updated.(model)
	if m.expandedRepos["astrostl/a"] {
		t.Fatal("expected repo a folded from child row")
	}
	if m.cursor != 0 || m.rows[m.cursor].typ != rowRepo || m.rows[m.cursor].item.Repo != "astrostl/a" {
		t.Fatalf("expected cursor snapped to repo a, got cursor=%d row=%+v", m.cursor, m.rows[m.cursor])
	}

	// h folds all, l unfolds all
	m.setAllExpanded(true)
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'h'}})
	m = updated.(model)
	if len(m.rows) != 2 {
		t.Fatalf("expected 2 rows after h (fold all), got %d", len(m.rows))
	}
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}})
	m = updated.(model)
	if len(m.rows) != 4 {
		t.Fatalf("expected 4 rows after l (unfold all), got %d", len(m.rows))
	}
}

func TestSorting(t *testing.T) {
	rep := sampleReport()
	m := newModel()
	m.loading = false
	m.report = rep
	m.groups = buildGroups(rep)
	m.setAllExpanded(false)

	// Default sort is by Updated (astrostl/a is newer)
	m.sortField = SortUpdated
	m.rebuildRows()
	if m.rows[0].item.Repo != "astrostl/a" {
		t.Fatalf("expected repo a first for SortUpdated, got %s", m.rows[0].item.Repo)
	}

	// Next: Sort by Name (astrostl/a < astrostl/b)
	m.sortField = m.sortField.Next()
	if m.sortField != SortName {
		t.Fatalf("expected SortName next, got %s", m.sortField)
	}
	m.rebuildRows()
	if m.rows[0].item.Repo != "astrostl/a" {
		t.Fatalf("expected repo a first for SortName, got %s", m.rows[0].item.Repo)
	}

	// Next: Sort by Commits (neither sample repo has a count; keep updated-at order)
	m.sortField = m.sortField.Next()
	if m.sortField != SortCommits {
		t.Fatalf("expected SortCommits next, got %s", m.sortField)
	}

	// Next: Sort by Issues (astrostl/a has 1 Issue, b has 0)
	m.sortField = m.sortField.Next()
	if m.sortField != SortIssues {
		t.Fatalf("expected SortIssues next, got %s", m.sortField)
	}
	m.rebuildRows()
	if m.rows[0].item.Repo != "astrostl/a" {
		t.Fatalf("expected repo a first for SortIssues, got %s", m.rows[0].item.Repo)
	}

	// Next: Sort by PRs (astrostl/b has 1 PR, a has 0)
	m.sortField = m.sortField.Next()
	if m.sortField != SortPRs {
		t.Fatalf("expected SortPRs next, got %s", m.sortField)
	}
	m.rebuildRows()
	if m.rows[0].item.Repo != "astrostl/b" {
		t.Fatalf("expected repo b first for SortPRs, got %s", m.rows[0].item.Repo)
	}

	// Next: Sort by Stars (astrostl/b has 10 stars, a has 3)
	m.sortField = m.sortField.Next()
	if m.sortField != SortStars {
		t.Fatalf("expected SortStars next, got %s", m.sortField)
	}
	m.rebuildRows()
	if m.rows[0].item.Repo != "astrostl/b" {
		t.Fatalf("expected repo b first for SortStars, got %s", m.rows[0].item.Repo)
	}

	// Next: cycles back to SortUpdated
	m.sortField = m.sortField.Next()
	if m.sortField != SortUpdated {
		t.Fatalf("expected cycle back to SortUpdated, got %s", m.sortField)
	}

	// Prev: walks left from Update to Stars, then Name
	m.sortField = m.sortField.Prev()
	if m.sortField != SortStars {
		t.Fatalf("expected SortStars prev from Update, got %s", m.sortField)
	}
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'S'}})
	m = updated.(model)
	if m.sortField != SortPRs {
		t.Fatalf("S should move sort left to PRs, got %s", m.sortField)
	}
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
	m = updated.(model)
	if m.sortField != SortStars {
		t.Fatalf("s should move sort right to Stars, got %s", m.sortField)
	}
}

func TestFiltering(t *testing.T) {
	rep := sampleReport()
	m := newModel()
	m.loading = false
	m.report = rep
	m.groups = buildGroups(rep)

	m.filterText = "beta"
	m.rebuildRows()

	if len(m.rows) == 0 || m.rows[0].item.Repo != "astrostl/b" {
		t.Fatalf("expected repo b to match 'beta', got rows: %+v", m.rows)
	}

	m.filterText = "first bug"
	m.rebuildRows()
	if len(m.rows) == 0 || m.rows[0].item.Repo != "astrostl/a" {
		t.Fatalf("expected repo a to match 'first bug', got rows: %+v", m.rows)
	}
}

func TestPrivateAndForkToggles(t *testing.T) {
	m := newModel()
	m.loading = false
	m.report = sampleReport()
	m.groups = buildGroups(m.report)
	m.rebuildRows()
	m.height = 24
	m.width = 80

	if len(m.rows) != 2 {
		t.Fatalf("len(rows) = %d, want 2", len(m.rows))
	}

	m.cursor = 1
	m.scroll = 1
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'P'}})
	m = updated.(model)
	if !m.onlyPrivate {
		t.Fatal("expected onlyPrivate after P")
	}
	if m.cursor != 0 || m.scroll != 0 {
		t.Fatalf("expected jump to line 1 after filter, cursor=%d scroll=%d", m.cursor, m.scroll)
	}
	if len(m.rows) != 1 || m.rows[0].item.Repo != "astrostl/a" || !m.rows[0].item.Private {
		t.Fatalf("private filter rows = %+v", m.rows)
	}

	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'F'}})
	m = updated.(model)
	if !m.onlyForks {
		t.Fatal("expected onlyForks after F")
	}
	if len(m.rows) != 0 {
		t.Fatalf("expected no private forks, got %+v", m.rows)
	}

	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'P'}})
	m = updated.(model)
	if m.onlyPrivate {
		t.Fatal("expected onlyPrivate off after second P")
	}
	if len(m.rows) != 1 || m.rows[0].item.Repo != "astrostl/b" || !m.rows[0].item.Fork {
		t.Fatalf("fork filter rows = %+v", m.rows)
	}

	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'F'}})
	m = updated.(model)
	if m.onlyForks {
		t.Fatal("expected onlyForks off after second F")
	}
	if len(m.rows) != 2 {
		t.Fatalf("expected both repos after clearing filters, got %+v", m.rows)
	}

	view := stripANSI(m.View())
	if strings.Contains(view, " · private") || strings.Contains(view, " · public") || strings.Contains(view, " · forks") || strings.Contains(view, " · sources") {
		t.Fatalf("did not expect visibility labels when toggles are off:\n%s", view)
	}

	m.onlyPrivate = true
	m.onlyForks = true
	view = stripANSI(m.View())
	if !strings.Contains(view, "private") || !strings.Contains(view, "forks") {
		t.Fatalf("expected visibility labels in title:\n%s", view)
	}
}

func TestPublicToggleIncludesForks(t *testing.T) {
	t1 := time.Date(2026, 8, 15, 10, 0, 0, 0, time.UTC)
	t2 := time.Date(2026, 8, 10, 10, 0, 0, 0, time.UTC)
	t3 := time.Date(2026, 8, 5, 10, 0, 0, 0, time.UTC)
	rep := Report{
		Repos: []Item{
			{Kind: KindRepo, Repo: "org/private", UpdatedAt: t1, Private: true},
			{Kind: KindRepo, Repo: "org/public-fork", UpdatedAt: t2, Fork: true},
			{Kind: KindRepo, Repo: "org/public-src", UpdatedAt: t3},
		},
	}
	m := newModel()
	m.loading = false
	m.report = rep
	m.groups = buildGroups(rep)
	m.rebuildRows()
	m.height = 24
	m.width = 80

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}})
	m = updated.(model)
	if !m.onlyPublic || m.onlyPrivate {
		t.Fatalf("expected public-only (and not private-only), public=%v private=%v", m.onlyPublic, m.onlyPrivate)
	}
	if len(m.rows) != 2 {
		t.Fatalf("expected 2 public repos, got %+v", m.rows)
	}
	got := map[string]bool{}
	for _, r := range m.rows {
		got[r.item.Repo] = true
		if r.item.Private {
			t.Fatalf("public filter included private repo %s", r.item.Repo)
		}
	}
	if !got["org/public-fork"] || !got["org/public-src"] {
		t.Fatalf("expected public source and public fork, got %v", got)
	}

	view := stripANSI(m.View())
	if !strings.Contains(view, " · public") {
		t.Fatalf("expected public label in title:\n%s", view)
	}

	// P should replace an active public filter
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'P'}})
	m = updated.(model)
	if !m.onlyPrivate || m.onlyPublic {
		t.Fatalf("expected P to replace p, public=%v private=%v", m.onlyPublic, m.onlyPrivate)
	}

	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'f'}})
	m = updated.(model)
	if !m.onlyNonForks || m.onlyForks {
		t.Fatalf("expected f non-forks, forks=%v sources=%v", m.onlyForks, m.onlyNonForks)
	}
	if len(m.rows) != 1 || m.rows[0].item.Repo != "org/private" || m.rows[0].item.Fork {
		t.Fatalf("expected private source after P+f, got %+v", m.rows)
	}

	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	m = updated.(model)
	if m.onlyPrivate || m.onlyPublic || m.onlyForks || m.onlyNonForks {
		t.Fatalf("expected a to clear visibility filters, private=%v public=%v forks=%v sources=%v", m.onlyPrivate, m.onlyPublic, m.onlyForks, m.onlyNonForks)
	}
	if len(m.rows) != 3 {
		t.Fatalf("expected all 3 repos after a, got %+v", m.rows)
	}
}

func TestSelectedCursorIsMagenta(t *testing.T) {
	lipgloss.SetColorProfile(termenv.ANSI256)
	cw := computeColWidths(80)
	now := time.Date(2026, 8, 16, 12, 0, 0, 0, time.UTC)
	r := row{
		typ:         rowRepo,
		item:        Item{Repo: "owner/name", UpdatedAt: now},
		issuesCount: 1,
	}
	got := renderRepoRow(r, true, cw, 80, now, true, true)
	if !strings.Contains(got, "207") {
		t.Fatalf("expected main cursor color 207 in selected row, got %q", got)
	}
	if !strings.Contains(stripANSI(got), "▸ owner/name") {
		t.Fatalf("expected fold marker on selected unfoldable row: %q", stripANSI(got))
	}
}

func TestUnfoldableIndicator(t *testing.T) {
	if got := repoPrefix(row{issuesCount: 1}, false); got != "▸ " {
		t.Fatalf("collapsed unfoldable prefix = %q", got)
	}
	if got := repoPrefix(row{issuesCount: 1, expanded: true}, false); got != "▾ " {
		t.Fatalf("expanded unfoldable prefix = %q", got)
	}
	if got := repoPrefix(row{prsCount: 2}, true); got != "▸ " {
		t.Fatalf("selected collapsed unfoldable prefix = %q", got)
	}
	if got := repoPrefix(row{}, false); got != "  " {
		t.Fatalf("leaf prefix = %q", got)
	}
	if got := repoPrefix(row{}, true); got != "▸ " {
		t.Fatalf("selected leaf prefix = %q", got)
	}

	m := newModel()
	m.loading = false
	m.report = sampleReport()
	m.groups = buildGroups(m.report)
	m.setAllExpanded(false)
	m.height = 24
	m.width = 100
	plain := stripANSI(m.View())
	if !strings.Contains(plain, "▸ astrostl/a") || !strings.Contains(plain, "▸ astrostl/b") {
		t.Fatalf("expected collapsed unfold markers:\n%s", plain)
	}

	m.setAllExpanded(true)
	plain = stripANSI(m.View())
	if !strings.Contains(plain, "▾ astrostl/a") || !strings.Contains(plain, "▾ astrostl/b") {
		t.Fatalf("expected expanded unfold markers:\n%s", plain)
	}
}

func TestViewMultiColumnLayout(t *testing.T) {
	rep := sampleReport()
	m := newModel()
	m.loading = false
	m.report = rep
	m.now = time.Date(2026, 8, 16, 12, 0, 0, 0, time.UTC)
	m.groups = buildGroups(rep)
	m.setAllExpanded(true)
	m.cursor = 0
	m.height = 24
	m.width = 100

	view := m.View()
	plain := stripANSI(view)

	// Check for core frame, TYPE column, STARS column, UPDATE column & column headers (without DESCRIPTION)
	for _, want := range []string{"╭", "╰", "│", "guh", "TYPE", "REPO", "COMMITS", "ISSUES", "PRS", "STARS", "UPDATE", "⊘", "⑂", "astrostl/a", "astrostl/b", "-1", "-6"} {
		if !strings.Contains(plain, want) {
			t.Fatalf("view missing %q:\n%s", want, plain)
		}
	}
	if strings.Contains(plain, "DESCRIPTION") {
		t.Fatalf("expected DESCRIPTION column to be dropped")
	}
	for _, labeled := range []string{"2 repos", "1 issues", "1 PRs", "★ 13"} {
		if strings.Contains(plain, labeled) {
			t.Fatalf("expected unlabeled column stats, found %q:\n%s", labeled, plain)
		}
	}

	top := strings.SplitN(plain, "\n", 2)[0]
	if !strings.HasPrefix(top, "╭") || !strings.Contains(top, "guh") {
		t.Fatalf("expected title in top border, got %q", top)
	}
	if !strings.Contains(top, "13") {
		t.Fatalf("expected unlabeled stats embedded in top border, got %q", top)
	}
	if !strings.Contains(top, "─") {
		t.Fatalf("expected dash rule around stats, got %q", top)
	}
}

func TestInspectorURLOnlyWithoutDescription(t *testing.T) {
	now := time.Date(2026, 8, 16, 12, 0, 0, 0, time.UTC)
	withDesc := Item{
		Kind: KindRepo, Repo: "astrostl/a", URL: "https://github.com/astrostl/a",
		Description: "Alpha repo", UpdatedAt: now,
	}
	got := stripANSI(renderInspectorLine(withDesc, 200, now))
	if strings.Contains(got, withDesc.URL) {
		t.Fatalf("did not expect URL when description is set: %q", got)
	}
	if !strings.Contains(got, "Alpha repo") {
		t.Fatalf("expected description in inspector: %q", got)
	}

	noDesc := withDesc
	noDesc.Description = ""
	got = stripANSI(renderInspectorLine(noDesc, 200, now))
	if !strings.Contains(got, noDesc.URL) {
		t.Fatalf("expected URL when description is empty: %q", got)
	}

	iss := Item{
		Kind: KindIssue, Repo: "astrostl/a", Number: 1, Title: "first bug",
		URL: "https://github.com/astrostl/a/issues/1", UpdatedAt: now, Author: "alice",
	}
	got = stripANSI(renderInspectorLine(iss, 200, now))
	if !strings.Contains(got, iss.URL) {
		t.Fatalf("expected URL for issue without description: %q", got)
	}
}

func TestTopBorderStatsAlignWithHeader(t *testing.T) {
	width := 100
	bodyW := innerWidth(width)
	cw := computeColWidths(bodyW)
	statsAt := cw.typeCol + cw.gap + cw.repo + cw.gap
	top := displayCells(stripANSI(titledRuleWithOverlay(width, "guh", renderCountStats(cw, 88, 4, 3, 99, true, true), statsAt)))
	header := displayCells(stripANSI(sideLine(width, renderColHeader(cw, SortUpdated, bodyW, true, true))))
	data := displayCells(stripANSI(sideLine(width, renderRepoRow(row{
		typ:         rowRepo,
		item:        Item{Repo: "owner/name", Stars: 7, CommitCount: 21, UpdatedAt: time.Date(2026, 8, 15, 0, 0, 0, 0, time.UTC)},
		issuesCount: 4,
		prsCount:    3,
	}, false, cw, bodyW, time.Date(2026, 8, 16, 0, 0, 0, 0, time.UTC), true, true))))

	if len(top) != len(header) || len(header) != len(data) {
		t.Fatalf("widths top=%d header=%d data=%d\n%s\n%s\n%s", len(top), len(header), len(data), string(top), string(header), string(data))
	}

	pos := func(line []rune, ch rune) int {
		for i := len(line) - 1; i >= 0; i-- {
			if line[i] == ch {
				return i
			}
		}
		return -1
	}
	// stars 99 should end at same column as header S? no, as data stars 7
	if p, d := pos(top, '9'), pos(data, '7'); p != d {
		t.Fatalf("stars digit col top=%d data=%d\n%s\n%s\n%s", p, d, string(top), string(header), string(data))
	}
	if p, d := pos(top, '3'), pos(data, '3'); p != d {
		t.Fatalf("prs digit col top=%d data=%d\n%s\n%s\n%s", p, d, string(top), string(header), string(data))
	}
	if p, d := pos(top, '4'), pos(data, '4'); p != d {
		t.Fatalf("issues digit col top=%d data=%d\n%s\n%s\n%s", p, d, string(top), string(header), string(data))
	}
	if p, d := pos(top, '8'), pos(data, '1'); p == -1 || d == -1 {
		t.Fatalf("commits digits missing\n%s\n%s\n%s", string(top), string(header), string(data))
	}
}

func TestDropDisplayPrefixKeepsBorderColor(t *testing.T) {
	lipgloss.SetColorProfile(termenv.ANSI256)
	line := dashFill(20)
	got := dropDisplayPrefix(line, 8)
	if !strings.Contains(got, "240") {
		t.Fatalf("expected border color 240 after cutting a dash run, got %q", got)
	}
	if strings.Contains(stripANSI(titledRuleWithOverlay(100, "guh · long-org-name", renderCountStats(computeColWidths(innerWidth(100)), 88, 4, 3, 99, true, true), 40)), "…") {
		t.Fatal("top rule should not insert an ellipsis")
	}
}

func displayCells(s string) []rune {
	s = strings.ReplaceAll(s, "─", "-")
	return []rune(s)
}

func TestRenderColStatsAlignsWithColumns(t *testing.T) {
	cw := computeColWidths(80)
	plain := strings.ReplaceAll(stripANSI(renderColStats(cw, 12, 88, 4, 3, 99, 80, true, true)), "─", " ")
	for _, labeled := range []string{"repos", "issues", "PRs", "Stars", "★"} {
		if strings.Contains(plain, labeled) {
			t.Fatalf("stats should be unlabeled, found %q in %q", labeled, plain)
		}
	}

	typ, repo, commits, issues, prs, stars, update := splitColCells(plain, cw)
	if typ != "" {
		t.Fatalf("type cell = %q, want empty", typ)
	}
	if repo != "12" {
		t.Fatalf("repo cell = %q, want 12", repo)
	}
	if commits != "88" {
		t.Fatalf("commits cell = %q, want 88", commits)
	}
	if issues != "4" {
		t.Fatalf("issues cell = %q, want 4", issues)
	}
	if prs != "3" {
		t.Fatalf("prs cell = %q, want 3", prs)
	}
	if stars != "99" {
		t.Fatalf("stars cell = %q, want 99", stars)
	}
	if update != "" {
		t.Fatalf("update cell = %q, want empty", update)
	}
}

func splitColCells(line string, cw colWidths) (typ, repo, commits, issues, prs, stars, update string) {
	pos := 0
	cut := func(w int) string {
		end := pos + w
		if end > len(line) {
			end = len(line)
		}
		if pos > len(line) {
			pos = len(line)
		}
		cell := strings.TrimSpace(line[pos:end])
		pos = end + cw.gap
		return cell
	}
	typ = cut(cw.typeCol)
	repo = cut(cw.repo)
	commits = cut(cw.commits)
	issues = cut(cw.issues)
	prs = cut(cw.prs)
	stars = cut(cw.stars)
	update = cut(cw.update)
	return
}

func TestOrgPickerSelectAndEscape(t *testing.T) {
	m := newModel()
	m.loading = false
	m.report = sampleReport()
	m.groups = buildGroups(m.report)
	m.rebuildRows()
	m.height = 24
	m.width = 80
	m.login = "astrostl"
	m.orgs = []string{"acme", "widgets-inc"}
	m.orgsLoaded = true

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'o'}})
	m = updated.(model)
	if !m.showOrgs {
		t.Fatal("expected org picker to open")
	}
	if cmd != nil {
		t.Fatal("expected no fetch when orgs already loaded")
	}
	if m.orgCursor != 0 {
		t.Fatalf("expected personal account selected, cursor=%d", m.orgCursor)
	}

	view := stripANSI(m.View())
	if !strings.Contains(view, "Switch organization") || !strings.Contains(view, "acme") || !strings.Contains(view, "(you)") {
		t.Fatalf("org picker missing expected content:\n%s", view)
	}

	updated, cmd = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = updated.(model)
	if m.showOrgs {
		t.Fatal("expected esc to close org picker")
	}
	if cmd != nil {
		t.Fatal("expected no reload after esc")
	}
	if m.owner != "" {
		t.Fatalf("esc should keep current owner, got %q", m.owner)
	}

	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'o'}})
	m = updated.(model)
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = updated.(model)
	if m.ownerChoices()[m.orgCursor] != "acme" {
		t.Fatalf("expected acme highlighted, cursor=%d choices=%v", m.orgCursor, m.ownerChoices())
	}

	updated, cmd = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(model)
	if m.showOrgs {
		t.Fatal("expected picker to close on select")
	}
	if m.owner != "acme" {
		t.Fatalf("owner = %q, want acme", m.owner)
	}
	if !m.loading {
		t.Fatal("expected reload after selecting org")
	}
	if cmd == nil {
		t.Fatal("expected fetch cmd after selecting org")
	}

	// Selecting the already-active owner should not refetch.
	m.loading = false
	m.orgsLoaded = true
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'o'}})
	m = updated.(model)
	updated, cmd = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(model)
	if m.owner != "acme" {
		t.Fatalf("owner changed unexpectedly to %q", m.owner)
	}
	if cmd != nil {
		t.Fatal("expected no reload when re-selecting the same org")
	}
}

func TestURLPrompt(t *testing.T) {
	m := newModel()
	m.loading = false
	m.report = sampleReport()
	m.groups = buildGroups(m.report)
	m.rebuildRows()
	m.height = 24
	m.width = 80
	m.login = "astrostl"

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'u'}})
	m = updated.(model)
	if !m.urlMode || cmd != nil {
		t.Fatalf("expected url prompt, mode=%v cmd=%v", m.urlMode, cmd)
	}
	if m.urlText != "astrostl" {
		t.Fatalf("prefill = %q", m.urlText)
	}
	view := stripANSI(m.View())
	if !strings.Contains(view, "User/org/repo:") {
		t.Fatalf("missing URL prompt:\n%s", view)
	}

	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyCtrlU})
	m = updated.(model)
	for _, r := range "https://github.com/acme/widgets" {
		updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
		m = updated.(model)
	}
	updated, cmd = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(model)
	if m.urlMode {
		t.Fatal("expected prompt to close")
	}
	if m.owner != "acme" {
		t.Fatalf("owner = %q, want acme", m.owner)
	}
	if m.pendingRepo != "acme/widgets" {
		t.Fatalf("pendingRepo = %q", m.pendingRepo)
	}
	if !m.loading || cmd == nil {
		t.Fatal("expected fetch after URL")
	}

	m = newModel()
	m.loading = false
	m.login = "astrostl"
	m.owner = "acme"
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'u'}})
	m = updated.(model)
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = updated.(model)
	if m.urlMode {
		t.Fatal("esc should cancel")
	}
	if m.owner != "acme" {
		t.Fatalf("esc changed owner to %q", m.owner)
	}
}

func TestURLPromptOwnerRepo(t *testing.T) {
	m := newModel()
	m.loading = false
	m.login = "astrostl"
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'u'}})
	m = updated.(model)
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyCtrlU})
	m = updated.(model)
	for _, r := range "acme/widgets" {
		updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
		m = updated.(model)
	}
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(model)
	if m.owner != "acme" || m.pendingRepo != "acme/widgets" {
		t.Fatalf("owner=%q pending=%q", m.owner, m.pendingRepo)
	}
	if !m.loading || cmd == nil {
		t.Fatal("expected fetch for owner/repo")
	}
}

func TestReposMsgStartsIssueAndCommitWaves(t *testing.T) {
	m := newModel()
	m.loading = true
	updated, cmd := m.Update(reposMsg{
		id: m.fetchID,
		repos: []Item{
			{Kind: KindRepo, Repo: "acme/one"},
			{Kind: KindRepo, Repo: "acme/two"},
		},
	})
	m = updated.(model)
	if cmd == nil {
		t.Fatal("expected count-fill commands")
	}
	if m.countInflight == 0 {
		t.Fatal("expected issue/PR count wave to start")
	}
	if m.commitInflight == 0 {
		t.Fatal("expected commit count wave to start in parallel")
	}
	if m.issuesReady() {
		t.Fatal("issues should stay muted until their wave finishes")
	}
	if m.commitsReady() {
		t.Fatal("commits should stay muted until their wave finishes")
	}

	updated, _ = m.Update(countsMsg{
		id:   m.fetchID,
		kind: countIssuesPRs,
		repos: []Item{
			{Repo: "acme/one", IssueCount: 2, PRCount: 1},
			{Repo: "acme/two", IssueCount: 0, PRCount: 0},
		},
	})
	m = updated.(model)
	if !m.issuesReady() {
		t.Fatal("issues should unmute when their wave is done")
	}
	if m.commitsReady() {
		t.Fatal("commits should stay muted while still in flight")
	}
	if m.countInflight != 0 {
		t.Fatalf("issue inflight = %d, want 0", m.countInflight)
	}
	if m.commitInflight == 0 {
		t.Fatal("commit wave should keep running after issues finish")
	}

	updated, _ = m.Update(countsMsg{
		id:   m.fetchID,
		kind: countCommits,
		repos: []Item{
			{Repo: "acme/one", CommitCount: 9},
			{Repo: "acme/two", CommitCount: 4},
		},
	})
	m = updated.(model)
	if !m.commitsReady() {
		t.Fatal("commits should unmute when their wave is done")
	}
	if m.report.Repos[0].CommitCount != 9 || m.report.Repos[0].IssueCount != 2 {
		t.Fatalf("merged counts = %+v", m.report.Repos[0])
	}
}

func TestUnfoldFetchesRepoItems(t *testing.T) {
	m := newModel()
	m.loading = false
	m.report = Report{
		Repos: []Item{{Kind: KindRepo, Repo: "acme/widgets", IssueCount: 2, UpdatedAt: time.Date(2026, 8, 15, 0, 0, 0, 0, time.UTC)}},
	}
	m.groups = buildGroups(m.report)
	m.rebuildRows()
	m.height = 24
	m.width = 80

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRight})
	m = updated.(model)
	if !m.expandedRepos["acme/widgets"] {
		t.Fatal("expected repo expanded")
	}
	if !m.loadingRepos["acme/widgets"] {
		t.Fatal("expected repo items fetch to start")
	}
	if cmd == nil {
		t.Fatal("expected fetch cmd")
	}
}

func TestShiftUResetsToYou(t *testing.T) {
	m := newModel()
	m.loading = false
	m.login = "astrostl"
	m.owner = "acme"
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'U'}})
	m = updated.(model)
	if m.owner != "" {
		t.Fatalf("owner = %q, want personal", m.owner)
	}
	if !m.loading || cmd == nil {
		t.Fatal("expected fetch for personal account")
	}

	m.loading = false
	updated, cmd = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'U'}})
	m = updated.(model)
	if m.owner != "" {
		t.Fatalf("owner changed to %q", m.owner)
	}
	if cmd != nil {
		t.Fatal("expected no fetch when already you")
	}
}

func TestCommitsModal(t *testing.T) {
	m := newModel()
	m.loading = false
	m.report = sampleReport()
	m.groups = buildGroups(m.report)
	m.rebuildRows()
	m.height = 24
	m.width = 80
	m.now = time.Date(2026, 8, 16, 12, 0, 0, 0, time.UTC)

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}})
	m = updated.(model)
	if !m.showCommits || m.commitsRepo != "astrostl/a" || !m.commitsLoading {
		t.Fatalf("expected commits modal for astrostl/a, got show=%v repo=%q loading=%v", m.showCommits, m.commitsRepo, m.commitsLoading)
	}
	if cmd == nil {
		t.Fatal("expected fetch commits cmd")
	}

	updated, _ = m.Update(commitsMsg{
		repo: "astrostl/a",
		commits: []Commit{
			{SHA: "abcdef1dead", Title: "Fix the thing", Author: "Ada", Date: m.now.Add(-2 * time.Hour), URL: "https://github.com/astrostl/a/commit/abcdef1dead"},
			{SHA: "1234567beef", Title: "Start project", Author: "Bob", Date: m.now.Add(-30 * 24 * time.Hour), URL: "https://github.com/astrostl/a/commit/1234567beef"},
		},
	})
	m = updated.(model)
	if m.commitsLoading || len(m.commits) != 2 {
		t.Fatalf("expected commits loaded, loading=%v n=%d", m.commitsLoading, len(m.commits))
	}

	view := stripANSI(m.View())
	wantDate := formatLocalDateTime(m.now.Add(-2 * time.Hour))
	for _, want := range []string{"Recent commits", "astrostl/a", "abcdef1", "Fix the thing", "Ada", "Start project", "Bob", wantDate} {
		if !strings.Contains(view, want) {
			t.Fatalf("commits modal missing %q:\n%s", want, view)
		}
	}
	if strings.Contains(view, " ago") || strings.Contains(view, " · -") {
		t.Fatalf("did not expect relative date in commits modal:\n%s", view)
	}

	updated, cmd = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(model)
	if cmd == nil {
		t.Fatal("expected open cmd for selected commit")
	}
	if !strings.Contains(m.status, "https://github.com/astrostl/a/commit/abcdef1dead") {
		t.Fatalf("status = %q", m.status)
	}

	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = updated.(model)
	if m.showCommits {
		t.Fatal("expected esc to close commits modal")
	}
}

func TestHelpModalToggle(t *testing.T) {
	m := newModel()
	m.width = 80
	m.height = 24
	m.showHelp = true

	view := m.View()
	plain := stripANSI(view)
	if !strings.Contains(plain, "Keyboard Shortcuts") || !strings.Contains(plain, "Navigation") {
		t.Fatalf("help modal missing expected content:\n%s", plain)
	}
}

func TestKeyUpdateMessages(t *testing.T) {
	m := newModel()
	m.loading = false
	m.report = sampleReport()
	m.groups = buildGroups(m.report)
	m.rebuildRows()
	m.height = 24
	m.width = 80

	// Down key
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = updated.(model)
	if m.cursor != 1 {
		t.Fatalf("expected cursor=1, got %d", m.cursor)
	}

	// Quit key
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	_ = updated
	if cmd == nil {
		t.Fatal("expected quit cmd")
	}
	if msg := cmd(); msg != tea.Quit() {
		t.Fatalf("expected tea.Quit, got %#v", msg)
	}
}

func stripANSI(s string) string {
	for {
		i := strings.Index(s, "\x1b[")
		if i < 0 {
			return s
		}
		j := strings.IndexByte(s[i:], 'm')
		if j < 0 {
			return s
		}
		s = s[:i] + s[i+j+1:]
	}
}
