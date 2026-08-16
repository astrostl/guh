package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os/exec"
	"sort"
	"strings"
	"sync"
	"time"
)

type Kind string

const (
	KindRepo  Kind = "repo"
	KindIssue Kind = "issue"
	KindPR    Kind = "pr"
)

type Item struct {
	Kind        Kind
	Repo        string
	Number      int
	Title       string
	URL         string
	UpdatedAt   time.Time
	PushedAt    time.Time
	Private     bool
	Fork        bool
	Archived    bool
	Description string
	Stars       int
	Language    string
	Author      string
	Comments    int
	IssueCount  int
	PRCount     int
}

func (it Item) TypeEmoji() string {
	if it.Kind != KindRepo {
		return ""
	}
	var emojis []string
	if it.Private {
		emojis = append(emojis, "🔒")
	}
	if it.Fork {
		emojis = append(emojis, "🍴")
	}
	return strings.Join(emojis, "")
}

func (it Item) Tags() []string {
	var tags []string
	if it.Private {
		tags = append(tags, "🔒")
	}
	if it.Fork {
		tags = append(tags, "🍴")
	}
	if it.Archived {
		tags = append(tags, "archived")
	}
	return tags
}

func (it Item) ChildLabel() string {
	title := strings.TrimSpace(it.Title)
	if title == "" {
		return fmt.Sprintf("#%d", it.Number)
	}
	return fmt.Sprintf("#%d  %s", it.Number, title)
}

func (it Item) Label() string {
	switch it.Kind {
	case KindIssue, KindPR:
		title := strings.TrimSpace(it.Title)
		if title == "" {
			return fmt.Sprintf("%s#%d", it.Repo, it.Number)
		}
		return fmt.Sprintf("%s#%d  %s", it.Repo, it.Number, title)
	default:
		name := it.Repo
		if tags := it.Tags(); len(tags) > 0 {
			return name + "  " + strings.Join(tags, " ")
		}
		return name
	}
}

func formatDaysOffset(t time.Time, now time.Time) string {
	if t.IsZero() {
		return "-"
	}
	if now.IsZero() {
		now = time.Now()
	}
	diff := now.Sub(t)
	if diff <= 0 {
		return "0"
	}
	days := int(diff.Hours() / 24)
	if days == 0 {
		return "0"
	}
	return fmt.Sprintf("-%d", days)
}

func formatLocalDateTime(t time.Time) string {
	if t.IsZero() {
		return "-"
	}
	return t.Local().Format("Mon Jan 2, 2006 3:04 PM MST")
}

func formatRelativeTime(t time.Time, now time.Time) string {
	if t.IsZero() {
		return "-"
	}
	if now.IsZero() {
		now = time.Now()
	}
	diff := now.Sub(t)
	if diff < 0 {
		return "just now"
	}
	switch {
	case diff < time.Minute:
		return "just now"
	case diff < time.Hour:
		mins := int(diff.Minutes())
		if mins <= 1 {
			return "1m ago"
		}
		return fmt.Sprintf("%dm ago", mins)
	case diff < 24*time.Hour:
		hrs := int(diff.Hours())
		if hrs <= 1 {
			return "1h ago"
		}
		return fmt.Sprintf("%dh ago", hrs)
	case diff < 30*24*time.Hour:
		days := int(diff.Hours() / 24)
		if days <= 1 {
			return "1d ago"
		}
		return fmt.Sprintf("%dd ago", days)
	case diff < 365*24*time.Hour:
		months := int(diff.Hours() / (24 * 30))
		if months <= 1 {
			return "1mo ago"
		}
		return fmt.Sprintf("%dmo ago", months)
	default:
		years := int(diff.Hours() / (24 * 365))
		if years <= 1 {
			return "1y ago"
		}
		return fmt.Sprintf("%dy ago", years)
	}
}

type Report struct {
	Repos  []Item
	Issues []Item
	PRs    []Item
}

func (r Report) Counts() (repos, issues, prs int) {
	return len(r.Repos), len(r.Issues), len(r.PRs)
}

type runner func(ctx context.Context, name string, args ...string) ([]byte, error)

func execRunner(ctx context.Context, name string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	out, err := cmd.Output()
	if err != nil {
		var extra string
		if ee, ok := err.(*exec.ExitError); ok {
			extra = strings.TrimSpace(string(ee.Stderr))
		}
		if extra != "" {
			return nil, fmt.Errorf("%s %s: %w: %s", name, strings.Join(args, " "), err, extra)
		}
		return nil, fmt.Errorf("%s %s: %w", name, strings.Join(args, " "), err)
	}
	return bytes.TrimSpace(out), nil
}

func loadReport(ctx context.Context, owner string) (Report, error) {
	if _, err := exec.LookPath("gh"); err != nil {
		return Report{}, fmt.Errorf("gh is not installed (https://cli.github.com)")
	}
	return loadReportWith(ctx, execRunner, owner)
}

func loadReportWith(ctx context.Context, run runner, owner string) (Report, error) {
	var (
		wg      sync.WaitGroup
		repos   []Item
		counted []Item
		issues  []Item
		prs     []Item
		errR    error
		errA    error
	)

	wg.Add(2)
	go func() {
		defer wg.Done()
		repos, errR = fetchRepos(ctx, run, owner)
	}()
	go func() {
		defer wg.Done()
		counted, issues, prs, errA = fetchAllActivity(ctx, run, owner)
	}()
	wg.Wait()

	if errR != nil {
		return Report{}, errR
	}
	if errA != nil {
		return Report{}, errA
	}

	return Report{Repos: applyRepoCounts(repos, counted), Issues: issues, PRs: prs}, nil
}

type ghRepo struct {
	NameWithOwner   string `json:"nameWithOwner"`
	URL             string `json:"url"`
	IsPrivate       bool   `json:"isPrivate"`
	IsFork          bool   `json:"isFork"`
	IsArchived      bool   `json:"isArchived"`
	Description     string `json:"description"`
	UpdatedAt       string `json:"updatedAt"`
	PushedAt        string `json:"pushedAt"`
	StargazerCount  int    `json:"stargazerCount"`
	PrimaryLanguage struct {
		Name string `json:"name"`
	} `json:"primaryLanguage"`
}

type activityPage struct {
	Repos   []Item
	Issues  []Item
	PRs     []Item
	After   string
	HasMore bool
}

const (
	gqlOwnerActivity = `query($login: String!, $after: String) {
  repositoryOwner(login: $login) {
    repositories(first: 50, after: $after, ownerAffiliations: OWNER, orderBy: {field: UPDATED_AT, direction: DESC}) {
      pageInfo { hasNextPage endCursor }
      nodes {
        nameWithOwner
        issues(states: OPEN, first: 50, orderBy: {field: UPDATED_AT, direction: DESC}) {
          totalCount
          nodes { number title url updatedAt author { login } }
        }
        pullRequests(states: OPEN, first: 50, orderBy: {field: UPDATED_AT, direction: DESC}) {
          totalCount
          nodes { number title url updatedAt author { login } }
        }
      }
    }
  }
}`
	gqlViewerActivity = `query($after: String) {
  viewer {
    repositories(first: 50, after: $after, ownerAffiliations: OWNER, orderBy: {field: UPDATED_AT, direction: DESC}) {
      pageInfo { hasNextPage endCursor }
      nodes {
        nameWithOwner
        issues(states: OPEN, first: 50, orderBy: {field: UPDATED_AT, direction: DESC}) {
          totalCount
          nodes { number title url updatedAt author { login } }
        }
        pullRequests(states: OPEN, first: 50, orderBy: {field: UPDATED_AT, direction: DESC}) {
          totalCount
          nodes { number title url updatedAt author { login } }
        }
      }
    }
  }
}`
)

type gqlActor struct {
	Login string `json:"login"`
}

type gqlIssueNode struct {
	Number    int      `json:"number"`
	Title     string   `json:"title"`
	URL       string   `json:"url"`
	UpdatedAt string   `json:"updatedAt"`
	Author    gqlActor `json:"author"`
}

type gqlIssueConn struct {
	TotalCount int            `json:"totalCount"`
	Nodes      []gqlIssueNode `json:"nodes"`
}

type gqlRepoNode struct {
	NameWithOwner string       `json:"nameWithOwner"`
	Issues        gqlIssueConn `json:"issues"`
	PullRequests  gqlIssueConn `json:"pullRequests"`
}

type gqlRepoConn struct {
	PageInfo struct {
		HasNextPage bool   `json:"hasNextPage"`
		EndCursor   string `json:"endCursor"`
	} `json:"pageInfo"`
	Nodes []gqlRepoNode `json:"nodes"`
}

type gqlActivityResp struct {
	Data struct {
		RepositoryOwner *struct {
			Repositories gqlRepoConn `json:"repositories"`
		} `json:"repositoryOwner"`
		Viewer *struct {
			Repositories gqlRepoConn `json:"repositories"`
		} `json:"viewer"`
	} `json:"data"`
	Errors []struct {
		Message string `json:"message"`
	} `json:"errors"`
}

func fetchRepos(ctx context.Context, run runner, owner string) ([]Item, error) {
	args := []string{"repo", "list"}
	if owner != "" {
		args = append(args, owner)
	}
	args = append(args, "--limit", "1000",
		"--json", "nameWithOwner,url,isPrivate,isFork,isArchived,description,updatedAt,stargazerCount,primaryLanguage")
	out, err := run(ctx, "gh", args...)
	if err != nil {
		return nil, fmt.Errorf("list repos: %w", err)
	}
	var raw []ghRepo
	if err := json.Unmarshal(out, &raw); err != nil {
		return nil, fmt.Errorf("parse repos: %w", err)
	}
	items := make([]Item, 0, len(raw))
	for _, r := range raw {
		items = append(items, Item{
			Kind:        KindRepo,
			Repo:        r.NameWithOwner,
			Title:       r.NameWithOwner,
			URL:         r.URL,
			UpdatedAt:   parseTime(r.UpdatedAt),
			Private:     r.IsPrivate,
			Fork:        r.IsFork,
			Archived:    r.IsArchived,
			Description: strings.TrimSpace(r.Description),
			Stars:       r.StargazerCount,
			Language:    r.PrimaryLanguage.Name,
		})
	}
	sort.SliceStable(items, func(i, j int) bool {
		return items[i].UpdatedAt.After(items[j].UpdatedAt)
	})
	return items, nil
}

func fetchAllActivity(ctx context.Context, run runner, owner string) (counted, issues, prs []Item, err error) {
	after := ""
	for {
		page, err := fetchActivityPage(ctx, run, owner, after)
		if err != nil {
			return nil, nil, nil, err
		}
		counted = append(counted, page.Repos...)
		issues = append(issues, page.Issues...)
		prs = append(prs, page.PRs...)
		if !page.HasMore {
			break
		}
		after = page.After
	}
	return counted, issues, prs, nil
}

func fetchActivityPage(ctx context.Context, run runner, owner, after string) (activityPage, error) {
	args := []string{"api", "graphql"}
	if owner != "" {
		args = append(args, "-f", "query="+gqlOwnerActivity, "-f", "login="+owner)
	} else {
		args = append(args, "-f", "query="+gqlViewerActivity)
	}
	if after != "" {
		args = append(args, "-f", "after="+after)
	}
	out, err := run(ctx, "gh", args...)
	if err != nil {
		return activityPage{}, fmt.Errorf("list issues and PRs: %w", err)
	}
	var raw gqlActivityResp
	if err := json.Unmarshal(out, &raw); err != nil {
		return activityPage{}, fmt.Errorf("parse issues and PRs: %w", err)
	}
	if len(raw.Errors) > 0 {
		return activityPage{}, fmt.Errorf("list issues and PRs: %s", raw.Errors[0].Message)
	}

	var conn gqlRepoConn
	switch {
	case owner != "" && raw.Data.RepositoryOwner != nil:
		conn = raw.Data.RepositoryOwner.Repositories
	case owner == "" && raw.Data.Viewer != nil:
		conn = raw.Data.Viewer.Repositories
	default:
		return activityPage{}, fmt.Errorf("list issues and PRs: empty GraphQL response")
	}

	page := activityPage{
		After:   conn.PageInfo.EndCursor,
		HasMore: conn.PageInfo.HasNextPage && conn.PageInfo.EndCursor != "",
	}
	for _, n := range conn.Nodes {
		page.Repos = append(page.Repos, Item{
			Kind:       KindRepo,
			Repo:       n.NameWithOwner,
			IssueCount: n.Issues.TotalCount,
			PRCount:    n.PullRequests.TotalCount,
		})
		for _, iss := range n.Issues.Nodes {
			page.Issues = append(page.Issues, itemFromGQL(KindIssue, n.NameWithOwner, iss))
		}
		for _, pr := range n.PullRequests.Nodes {
			page.PRs = append(page.PRs, itemFromGQL(KindPR, n.NameWithOwner, pr))
		}
	}
	return page, nil
}

func itemFromGQL(kind Kind, repo string, n gqlIssueNode) Item {
	return Item{
		Kind:      kind,
		Repo:      repo,
		Number:    n.Number,
		Title:     n.Title,
		URL:       n.URL,
		UpdatedAt: parseTime(n.UpdatedAt),
		Author:    n.Author.Login,
	}
}

func applyRepoCounts(repos, counted []Item) []Item {
	byName := make(map[string]Item, len(counted))
	for _, c := range counted {
		byName[c.Repo] = c
	}
	out := make([]Item, len(repos))
	copy(out, repos)
	for i, r := range out {
		if c, ok := byName[r.Repo]; ok {
			out[i].IssueCount = c.IssueCount
			out[i].PRCount = c.PRCount
		}
	}
	return out
}

func fetchOwnerChoices(ctx context.Context, run runner) (login string, orgs []string, err error) {
	if loginOut, loginErr := run(ctx, "gh", "api", "user", "--jq", ".login"); loginErr == nil {
		login = strings.TrimSpace(string(loginOut))
	}

	out, err := run(ctx, "gh", "org", "list", "--limit", "1000")
	if err != nil {
		return login, nil, fmt.Errorf("list orgs: %w", err)
	}
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.Contains(line, " ") {
			continue
		}
		if login != "" && strings.EqualFold(line, login) {
			continue
		}
		orgs = append(orgs, line)
	}
	return login, orgs, nil
}

type Commit struct {
	SHA    string
	Title  string
	Author string
	Date   time.Time
	URL    string
}

func (c Commit) ShortSHA() string {
	if len(c.SHA) > 7 {
		return c.SHA[:7]
	}
	return c.SHA
}

func fetchRecentCommits(ctx context.Context, run runner, repo string, limit int) ([]Commit, error) {
	repo = strings.TrimSpace(repo)
	if repo == "" {
		return nil, fmt.Errorf("no repository selected")
	}
	if limit <= 0 {
		limit = 7
	}
	out, err := run(ctx, "gh", "api", fmt.Sprintf("repos/%s/commits?per_page=%d", repo, limit))
	if err != nil {
		return nil, fmt.Errorf("list commits: %w", err)
	}
	var raw []struct {
		SHA     string `json:"sha"`
		HTMLURL string `json:"html_url"`
		Commit  struct {
			Message string `json:"message"`
			Author  struct {
				Name string `json:"name"`
				Date string `json:"date"`
			} `json:"author"`
		} `json:"commit"`
	}
	if err := json.Unmarshal(out, &raw); err != nil {
		return nil, fmt.Errorf("parse commits: %w", err)
	}
	items := make([]Commit, 0, len(raw))
	for _, c := range raw {
		title := strings.TrimSpace(c.Commit.Message)
		if i := strings.IndexByte(title, '\n'); i >= 0 {
			title = strings.TrimSpace(title[:i])
		}
		items = append(items, Commit{
			SHA:    c.SHA,
			Title:  title,
			Author: strings.TrimSpace(c.Commit.Author.Name),
			Date:   parseTime(c.Commit.Author.Date),
			URL:    c.HTMLURL,
		})
	}
	return items, nil
}

func parseTime(s string) time.Time {
	if s == "" {
		return time.Time{}
	}
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return time.Time{}
	}
	return t
}

// parseOwnerInput accepts a username, org, owner/repo, or GitHub URL.
// repo is "owner/name" when the input points at a repository.
func parseOwnerInput(raw string) (owner, repo string, err error) {
	s := strings.TrimSpace(raw)
	s = strings.Trim(s, `"'`)
	if s == "" {
		return "", "", fmt.Errorf("empty")
	}

	lower := strings.ToLower(s)
	if strings.Contains(lower, "github.com") || strings.HasPrefix(lower, "http://") || strings.HasPrefix(lower, "https://") {
		if !strings.Contains(lower, "://") {
			s = "https://" + s
			lower = strings.ToLower(s)
		}
		u, perr := url.Parse(s)
		if perr != nil || u.Host == "" {
			return "", "", fmt.Errorf("not a GitHub URL")
		}
		host := strings.ToLower(u.Hostname())
		if host != "github.com" && host != "www.github.com" {
			return "", "", fmt.Errorf("not a GitHub URL")
		}
		parts := strings.Split(strings.Trim(u.Path, "/"), "/")
		if len(parts) == 0 || parts[0] == "" {
			return "", "", fmt.Errorf("no user or org in URL")
		}
		owner = parts[0]
		if len(parts) >= 2 && parts[1] != "" && !isGitHubReservedPath(parts[1]) {
			repo = owner + "/" + parts[1]
		}
		return owner, repo, nil
	}

	if strings.ContainsAny(s, " \t") {
		return "", "", fmt.Errorf("not a user, org, or URL")
	}
	if i := strings.IndexByte(s, '/'); i > 0 {
		owner = s[:i]
		rest := strings.Trim(s[i+1:], "/")
		if rest != "" {
			if j := strings.IndexByte(rest, '/'); j >= 0 {
				rest = rest[:j]
			}
			if rest != "" {
				repo = owner + "/" + rest
			}
		}
		return owner, repo, nil
	}
	return s, "", nil
}

func isGitHubReservedPath(p string) bool {
	switch strings.ToLower(p) {
	case "orgs", "users", "settings", "login", "signup", "topics", "explore", "marketplace", "sponsors", "notifications", "issues", "pulls":
		return true
	default:
		return false
	}
}
