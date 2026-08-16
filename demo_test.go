package main

import (
	"strings"
	"testing"
)

func TestDemoReportPersonal(t *testing.T) {
	r := demoReport("")
	if len(r.Repos) < 8 {
		t.Fatalf("personal repos = %d, want a full desk", len(r.Repos))
	}
	var priv, fork, arch, issues, prs int
	for _, repo := range r.Repos {
		if !strings.HasPrefix(repo.Repo, "mira/") {
			t.Fatalf("personal repo %q", repo.Repo)
		}
		if repo.Private {
			priv++
		}
		if repo.Fork {
			fork++
		}
		if repo.Archived {
			arch++
		}
		if repo.IssueCount != 0 && repo.IssueCount < 0 {
			t.Fatalf("negative issue count on %s", repo.Repo)
		}
	}
	for _, it := range r.Issues {
		issues++
		if it.Kind != KindIssue || it.Author == "" || it.Title == "" {
			t.Fatalf("issue %+v", it)
		}
	}
	for _, it := range r.PRs {
		prs++
		if it.Kind != KindPR || it.Author == "" || it.Title == "" {
			t.Fatalf("pr %+v", it)
		}
	}
	if priv == 0 || fork == 0 || arch == 0 || issues == 0 || prs == 0 {
		t.Fatalf("mix priv=%d fork=%d arch=%d issues=%d prs=%d", priv, fork, arch, issues, prs)
	}
}

func TestDemoReportOrgsDiffer(t *testing.T) {
	personal := repoNames(demoReport(""))
	lantern := repoNames(demoReport("lantern"))
	if len(lantern) == 0 {
		t.Fatal("lantern empty")
	}
	for name := range lantern {
		if personal[name] {
			t.Fatalf("lantern repo %s also in personal", name)
		}
		if !strings.HasPrefix(name, "lantern/") {
			t.Fatalf("lantern repo %q", name)
		}
	}
	if len(demoReport("northwind").Repos) == 0 || len(demoReport("rivulet").Repos) == 0 {
		t.Fatal("missing org")
	}
	if len(demoReport("unknown").Repos) != 0 {
		t.Fatal("unknown owner should be empty")
	}
}

func TestDemoOwners(t *testing.T) {
	login, orgs := demoOwners()
	if login != "mira" || len(orgs) != 3 {
		t.Fatalf("login=%q orgs=%v", login, orgs)
	}
}

func TestDemoCommits(t *testing.T) {
	c := demoCommits("mira/lumen")
	if len(c) != 7 {
		t.Fatalf("len=%d", len(c))
	}
	if c[0].Author == "" || c[0].Title == "" || c[0].SHA == "" {
		t.Fatalf("thin commit %+v", c[0])
	}
	generic := demoCommits("mira/dotfiles")
	if len(generic) != 7 {
		t.Fatalf("generic len=%d", len(generic))
	}
	if demoCommits("") != nil {
		t.Fatal("empty repo")
	}
}

func TestDemoApplyReportKeepsNow(t *testing.T) {
	m := newModelWith(true)
	fixed := m.now
	r := demoReport("")
	m.applyReport(r.Repos, r.Issues, r.PRs)
	if !m.now.Equal(fixed) || !m.now.Equal(demoNow) {
		t.Fatalf("demo now moved from %v to %v", fixed, m.now)
	}
}

func TestDemoOrgPicker(t *testing.T) {
	m := newModelWith(true)
	m.width, m.height = 80, 24
	updated, cmd := m.openOrgPicker()
	m = updated.(model)
	if cmd == nil {
		t.Fatal("expected org fetch")
	}
	msg := cmd()
	orgs, ok := msg.(orgsMsg)
	if !ok {
		t.Fatalf("msg %T", msg)
	}
	updated, _ = m.Update(orgs)
	m = updated.(model)
	choices := m.ownerChoices()
	if len(choices) != 4 || choices[0] != "mira" {
		t.Fatalf("choices %v", choices)
	}
}

func repoNames(r Report) map[string]bool {
	out := make(map[string]bool, len(r.Repos))
	for _, repo := range r.Repos {
		out[repo.Repo] = true
	}
	return out
}
