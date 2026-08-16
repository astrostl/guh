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
	CommitCount int
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
	repos, err := fetchRepos(ctx, run, owner)
	if err != nil {
		return Report{}, err
	}
	var (
		wg                  sync.WaitGroup
		counted, commits    []Item
		countErr, commitErr error
	)
	wg.Add(2)
	go func() {
		defer wg.Done()
		counted, countErr = fetchRepoCounts(ctx, run, repos)
	}()
	go func() {
		defer wg.Done()
		commits, commitErr = fetchRepoCommitCounts(ctx, run, repos)
	}()
	wg.Wait()
	if countErr != nil {
		return Report{}, countErr
	}
	if commitErr != nil {
		return Report{}, commitErr
	}
	repos = applyIssuePRCounts(repos, counted)
	repos = applyCommitCounts(repos, commits)
	issues, prs, err := fetchReportItems(ctx, run, repos)
	if err != nil {
		return Report{}, err
	}
	return Report{Repos: repos, Issues: issues, PRs: prs}, nil
}

func fetchReportItems(ctx context.Context, run runner, repos []Item) (issues, prs []Item, err error) {
	for _, repo := range repos {
		if repo.IssueCount == 0 && repo.PRCount == 0 {
			continue
		}
		iss, pulls, err := fetchRepoItems(ctx, run, repo.Repo)
		if err != nil {
			return nil, nil, err
		}
		issues = append(issues, iss...)
		prs = append(prs, pulls...)
	}
	return issues, prs, nil
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
	// Issue/PR totalCount is cheap; commit history totalCount is slower per
	// repo, so those batches stay smaller. Both query types cost 1 GraphQL
	// point even at these sizes. The TUI runs the two waves at once.
	countBatchSize  = 50
	countWorkers    = 6
	commitBatchSize = 25
	commitWorkers   = 6

	gqlRepoIssues = `query($owner: String!, $name: String!, $after: String) {
  repository(owner: $owner, name: $name) {
    issues(states: OPEN, first: 100, after: $after, orderBy: {field: UPDATED_AT, direction: DESC}) {
      pageInfo { hasNextPage endCursor }
      nodes { number title url updatedAt author { login } }
    }
  }
}`
	gqlRepoPRs = `query($owner: String!, $name: String!, $after: String) {
  repository(owner: $owner, name: $name) {
    pullRequests(states: OPEN, first: 100, after: $after, orderBy: {field: UPDATED_AT, direction: DESC}) {
      pageInfo { hasNextPage endCursor }
      nodes { number title url updatedAt author { login } }
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
	PageInfo   gqlPageInfo    `json:"pageInfo"`
	Nodes      []gqlIssueNode `json:"nodes"`
}

type gqlPageInfo struct {
	HasNextPage bool   `json:"hasNextPage"`
	EndCursor   string `json:"endCursor"`
}

type gqlCountNode struct {
	Issues struct {
		TotalCount int `json:"totalCount"`
	} `json:"issues"`
	PullRequests struct {
		TotalCount int `json:"totalCount"`
	} `json:"pullRequests"`
	DefaultBranchRef *struct {
		Target *struct {
			History struct {
				TotalCount int `json:"totalCount"`
			} `json:"history"`
		} `json:"target"`
	} `json:"defaultBranchRef"`
}

func (n gqlCountNode) commitCount() int {
	if n.DefaultBranchRef == nil || n.DefaultBranchRef.Target == nil {
		return 0
	}
	return n.DefaultBranchRef.Target.History.TotalCount
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

func fetchRepoCounts(ctx context.Context, run runner, repos []Item) ([]Item, error) {
	return fetchAliasedBatches(ctx, run, repos, countBatchSize, countWorkers, fetchRepoCountBatch)
}

func fetchRepoCommitCounts(ctx context.Context, run runner, repos []Item) ([]Item, error) {
	return fetchAliasedBatches(ctx, run, repos, commitBatchSize, commitWorkers, fetchRepoCommitBatch)
}

func fetchAliasedBatches(ctx context.Context, run runner, repos []Item, size, workers int, fn func(context.Context, runner, []Item) ([]Item, error)) ([]Item, error) {
	if len(repos) == 0 {
		return nil, nil
	}
	type span struct{ i, j int }
	var jobs []span
	for i := 0; i < len(repos); i += size {
		j := i + size
		if j > len(repos) {
			j = len(repos)
		}
		jobs = append(jobs, span{i, j})
	}
	out := make([][]Item, len(jobs))
	errCh := make(chan error, len(jobs))
	sem := make(chan struct{}, workers)
	var wg sync.WaitGroup
	for k, job := range jobs {
		wg.Add(1)
		go func(k int, job span) {
			defer wg.Done()
			select {
			case <-ctx.Done():
				errCh <- ctx.Err()
				return
			case sem <- struct{}{}:
			}
			defer func() { <-sem }()
			batch, err := fn(ctx, run, repos[job.i:job.j])
			if err != nil {
				errCh <- err
				return
			}
			out[k] = batch
		}(k, job)
	}
	wg.Wait()
	close(errCh)
	if err, ok := <-errCh; ok && err != nil {
		return nil, err
	}
	var counted []Item
	for _, batch := range out {
		counted = append(counted, batch...)
	}
	return counted, nil
}

const gqlCountFields = "issues(states: OPEN) { totalCount } pullRequests(states: OPEN) { totalCount }"
const gqlCommitFields = "defaultBranchRef { target { ... on Commit { history { totalCount } } } }"

func fetchRepoCountBatch(ctx context.Context, run runner, repos []Item) ([]Item, error) {
	nodes, err := fetchAliasedRepoNodes(ctx, run, repos, gqlCountFields, "list issue/PR counts")
	if err != nil {
		return nil, err
	}
	var counted []Item
	for _, n := range nodes {
		counted = append(counted, Item{
			Kind:       KindRepo,
			Repo:       n.repo,
			IssueCount: n.node.Issues.TotalCount,
			PRCount:    n.node.PullRequests.TotalCount,
		})
	}
	return counted, nil
}

func fetchRepoCommitBatch(ctx context.Context, run runner, repos []Item) ([]Item, error) {
	nodes, err := fetchAliasedRepoNodes(ctx, run, repos, gqlCommitFields, "list commit counts")
	if err != nil {
		return nil, err
	}
	var counted []Item
	for _, n := range nodes {
		counted = append(counted, Item{
			Kind:        KindRepo,
			Repo:        n.repo,
			CommitCount: n.node.commitCount(),
		})
	}
	return counted, nil
}

type aliasedCount struct {
	repo string
	node *gqlCountNode
}

func fetchAliasedRepoNodes(ctx context.Context, run runner, repos []Item, fields, errLabel string) ([]aliasedCount, error) {
	if len(repos) == 0 {
		return nil, nil
	}
	var (
		b     strings.Builder
		args  = []string{"api", "graphql"}
		alias = make([]string, 0, len(repos))
	)
	b.WriteString("query(")
	firstVar := true
	for i, repo := range repos {
		owner, name, err := splitRepoName(repo.Repo)
		if err != nil {
			continue
		}
		if !firstVar {
			b.WriteString(", ")
		}
		firstVar = false
		fmt.Fprintf(&b, "$o%d: String!, $n%d: String!", i, i)
		args = append(args, "-f", fmt.Sprintf("o%d=%s", i, owner), "-f", fmt.Sprintf("n%d=%s", i, name))
		alias = append(alias, fmt.Sprintf("r%d", i)+"\t"+repo.Repo)
	}
	if len(alias) == 0 {
		return nil, nil
	}
	b.WriteString(") {")
	for i, repo := range repos {
		if _, _, err := splitRepoName(repo.Repo); err != nil {
			continue
		}
		fmt.Fprintf(&b, " r%d: repository(owner: $o%d, name: $n%d) { %s }", i, i, i, fields)
	}
	b.WriteString(" }")
	args = append(args, "-f", "query="+b.String())

	out, err := run(ctx, "gh", args...)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", errLabel, err)
	}
	var raw struct {
		Data   map[string]*gqlCountNode `json:"data"`
		Errors []struct {
			Message string `json:"message"`
		} `json:"errors"`
	}
	if err := json.Unmarshal(out, &raw); err != nil {
		return nil, fmt.Errorf("parse %s: %w", errLabel, err)
	}
	if raw.Data == nil && len(raw.Errors) > 0 {
		return nil, fmt.Errorf("%s: %s", errLabel, raw.Errors[0].Message)
	}
	var counted []aliasedCount
	for _, pair := range alias {
		key, repo, _ := strings.Cut(pair, "\t")
		node := raw.Data[key]
		if node == nil {
			continue
		}
		counted = append(counted, aliasedCount{repo: repo, node: node})
	}
	return counted, nil
}

func splitRepoName(repo string) (owner, name string, err error) {
	repo = strings.TrimSpace(repo)
	i := strings.IndexByte(repo, '/')
	if i <= 0 || i == len(repo)-1 {
		return "", "", fmt.Errorf("invalid repo %q", repo)
	}
	return repo[:i], repo[i+1:], nil
}

func fetchRepoItems(ctx context.Context, run runner, repo string) (issues, prs []Item, err error) {
	owner, name, err := splitRepoName(repo)
	if err != nil {
		return nil, nil, err
	}
	issues, err = fetchRepoConn(ctx, run, repo, owner, name, KindIssue, gqlRepoIssues)
	if err != nil {
		return nil, nil, err
	}
	prs, err = fetchRepoConn(ctx, run, repo, owner, name, KindPR, gqlRepoPRs)
	if err != nil {
		return nil, nil, err
	}
	return issues, prs, nil
}

func fetchRepoConn(ctx context.Context, run runner, repo, owner, name string, kind Kind, query string) ([]Item, error) {
	var (
		items []Item
		after string
	)
	field := "issues"
	if kind == KindPR {
		field = "pullRequests"
	}
	for {
		args := []string{"api", "graphql", "-f", "query=" + query, "-f", "owner=" + owner, "-f", "name=" + name}
		if after == "" {
			args = append(args, "-F", "after=null")
		} else {
			args = append(args, "-f", "after="+after)
		}
		out, err := run(ctx, "gh", args...)
		if err != nil {
			return nil, fmt.Errorf("list %s %s: %w", repo, field, err)
		}
		var raw struct {
			Data struct {
				Repository *struct {
					Issues       gqlIssueConn `json:"issues"`
					PullRequests gqlIssueConn `json:"pullRequests"`
				} `json:"repository"`
			} `json:"data"`
			Errors []struct {
				Message string `json:"message"`
			} `json:"errors"`
		}
		if err := json.Unmarshal(out, &raw); err != nil {
			return nil, fmt.Errorf("parse %s %s: %w", repo, field, err)
		}
		if len(raw.Errors) > 0 {
			return nil, fmt.Errorf("list %s %s: %s", repo, field, raw.Errors[0].Message)
		}
		if raw.Data.Repository == nil {
			return nil, fmt.Errorf("list %s %s: empty GraphQL response", repo, field)
		}
		conn := raw.Data.Repository.Issues
		if kind == KindPR {
			conn = raw.Data.Repository.PullRequests
		}
		for _, n := range conn.Nodes {
			items = append(items, itemFromGQL(kind, repo, n))
		}
		if !conn.PageInfo.HasNextPage || conn.PageInfo.EndCursor == "" {
			break
		}
		after = conn.PageInfo.EndCursor
	}
	return items, nil
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
	return applyIssuePRCounts(applyCommitCounts(repos, counted), counted)
}

func applyIssuePRCounts(repos, counted []Item) []Item {
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

func applyCommitCounts(repos, counted []Item) []Item {
	byName := make(map[string]Item, len(counted))
	for _, c := range counted {
		byName[c.Repo] = c
	}
	out := make([]Item, len(repos))
	copy(out, repos)
	for i, r := range out {
		if c, ok := byName[r.Repo]; ok {
			out[i].CommitCount = c.CommitCount
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
