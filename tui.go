package main

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

const (
	cMain   = "207" // main, sort, cursor
	cCommit = "75"
	cIssue  = "197"
	cPR     = "141"
	cStar   = "227"

	cBorder   = "240"
	cText     = "254"
	cTextDim  = "245"
	cMuted    = "242"
	cURL      = cPR
	cOk       = "114"
	cErr      = "203"
	cSelBg    = "237"
	cSelFg    = "255"
	cCursor   = cMain
	cBranch   = "241"
	cHeaderFg = "252"
	cSearchBg = "238"
	cSearchFg = cMain
)

// Styles
var (
	styleBorder     = lipgloss.NewStyle().Foreground(lipgloss.Color(cBorder))
	styleTitle      = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(cMain))
	styleText       = lipgloss.NewStyle().Foreground(lipgloss.Color(cText))
	styleTextDim    = lipgloss.NewStyle().Foreground(lipgloss.Color(cTextDim))
	styleMuted      = lipgloss.NewStyle().Foreground(lipgloss.Color(cMuted))
	styleIssue      = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(cIssue))
	stylePR         = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(cPR))
	styleStar       = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(cStar))
	styleCommit     = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(cCommit))
	styleURL        = lipgloss.NewStyle().Foreground(lipgloss.Color(cURL)).Underline(true)
	styleStatus     = lipgloss.NewStyle().Foreground(lipgloss.Color(cOk))
	styleError      = lipgloss.NewStyle().Foreground(lipgloss.Color(cErr))
	styleHelp       = lipgloss.NewStyle().Foreground(lipgloss.Color(cTextDim))
	styleHelpKey    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(cMain))
	styleBranch     = lipgloss.NewStyle().Foreground(lipgloss.Color(cBranch))
	styleSelected   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(cSelFg)).Background(lipgloss.Color(cSelBg))
	styleCursor     = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(cCursor)).Background(lipgloss.Color(cSelBg))
	styleHeader     = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(cHeaderFg))
	styleHeaderSort = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(cMain))
	styleSearch     = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(cSearchFg)).Background(lipgloss.Color(cSearchBg))
)

type SortField int

const (
	SortUpdated SortField = iota
	SortName
	SortCommits
	SortIssues
	SortPRs
	SortStars
)

func (s SortField) String() string {
	switch s {
	case SortUpdated:
		return "Update"
	case SortName:
		return "Name"
	case SortCommits:
		return "Commits"
	case SortIssues:
		return "Issues"
	case SortPRs:
		return "PRs"
	case SortStars:
		return "Stars"
	default:
		return "Update"
	}
}

func (s SortField) Next() SortField {
	return (s + 1) % 6
}

type repoGroup struct {
	repo   Item
	issues []Item
	prs    []Item
}

type rowType int

const (
	rowRepo rowType = iota
	rowIssue
	rowPR
	rowSpacer
)

type row struct {
	typ         rowType
	item        Item
	repoName    string
	issuesCount int
	prsCount    int
	expanded    bool
	isLastChild bool
	spacer      bool
}

func (r row) unfoldable() bool {
	return r.issuesCount > 0 || r.prsCount > 0
}

func (r row) selectable() bool {
	return !r.spacer && r.typ != rowSpacer
}

type dataMsg struct{ report Report }
type reposMsg struct {
	id    int
	repos []Item
	err   error
}
type activityMsg struct {
	id   int
	page activityPage
	err  error
}
type countsMsg struct {
	id    int
	repos []Item
	kind  countKind
	err   error
}

type countKind int

const (
	countIssuesPRs countKind = iota
	countCommits
)

type errMsg struct{ err error }
type statusClearMsg struct{ id int }
type orgsMsg struct {
	login string
	orgs  []string
}
type orgsErrMsg struct{ err error }
type commitsMsg struct {
	repo    string
	commits []Commit
}
type commitsErrMsg struct{ err error }
type repoItemsMsg struct {
	id     int
	repo   string
	issues []Item
	prs    []Item
	err    error
}

type model struct {
	report        Report
	groups        []*repoGroup
	expandedRepos map[string]bool
	rows          []row
	cursor        int
	scroll        int
	width         int
	height        int
	loading       bool
	err           error
	status        string
	statusID      int
	sortField     SortField
	filterMode    bool
	filterText    string
	urlMode       bool
	urlText       string
	pendingRepo   string
	showHelp      bool
	now           time.Time
	onlyPrivate   bool
	onlyPublic    bool
	onlyForks     bool
	onlyNonForks  bool

	owner       string
	login       string
	orgs        []string
	orgsLoaded  bool
	showOrgs    bool
	orgsLoading bool
	orgsErr     error
	orgCursor   int

	fetchID        int
	awaitRepos     bool
	awaitActivity  bool
	counts         []Item
	loadedRepos    map[string]bool
	loadingRepos   map[string]bool
	countQueue     []int
	countInflight  int
	commitRepos    []Item
	commitQueue    []int
	commitInflight int

	showCommits    bool
	commitsRepo    string
	commits        []Commit
	commitsLoading bool
	commitsErr     error
	commitCursor   int

	demo bool
}

func forceTrueColor() {
	lipgloss.SetColorProfile(termenv.TrueColor)
}

func newModel() model {
	return newModelWith(false)
}

func newModelWith(demo bool) model {
	now := time.Now()
	if demo {
		now = demoNow
	}
	return model{
		loading:       true,
		width:         80,
		height:        24,
		sortField:     SortUpdated,
		expandedRepos: make(map[string]bool),
		now:           now,
		fetchID:       1,
		awaitRepos:    true,
		awaitActivity: true,
		loadedRepos:   make(map[string]bool),
		loadingRepos:  make(map[string]bool),
		demo:          demo,
	}
}

func (m *model) armFetch() {
	m.fetchID++
	m.awaitRepos = true
	m.awaitActivity = true
	m.counts = nil
	m.report.Issues = nil
	m.report.PRs = nil
	m.loadedRepos = make(map[string]bool)
	m.loadingRepos = make(map[string]bool)
	m.countQueue = nil
	m.countInflight = 0
	m.commitRepos = nil
	m.commitQueue = nil
	m.commitInflight = 0
}

func (m model) fetchCmd() tea.Cmd {
	id := m.fetchID
	owner := m.owner
	if m.demo {
		rep := demoReport(owner)
		return tea.Batch(
			func() tea.Msg {
				return reposMsg{id: id, repos: rep.Repos}
			},
			func() tea.Msg {
				return activityMsg{id: id, page: activityPage{Repos: rep.Repos, Issues: rep.Issues, PRs: rep.PRs}}
			},
		)
	}
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
		defer cancel()
		repos, err := fetchRepos(ctx, execRunner, owner)
		return reposMsg{id: id, repos: repos, err: err}
	}
}

func (m *model) startCountWave(kind countKind) tea.Cmd {
	queue, inflight, repos, size, max := &m.countQueue, &m.countInflight, m.report.Repos, countBatchSize, countWorkers
	if kind == countCommits {
		queue, inflight, repos, size, max = &m.commitQueue, &m.commitInflight, m.commitRepos, commitBatchSize, commitWorkers
	}
	var cmds []tea.Cmd
	for *inflight < max && len(*queue) > 0 {
		off := (*queue)[0]
		*queue = (*queue)[1:]
		*inflight++
		cmds = append(cmds, fetchCountsCmd(m.fetchID, repos, off, size, kind))
	}
	if len(cmds) == 0 {
		return nil
	}
	return tea.Batch(cmds...)
}

func (m model) commitTargets() []Item {
	seen := make(map[string]bool)
	var ordered []Item
	add := func(it Item) {
		if it.Repo == "" || seen[it.Repo] {
			return
		}
		seen[it.Repo] = true
		ordered = append(ordered, it)
	}
	for _, r := range m.rows {
		if r.typ == rowRepo {
			add(r.item)
		}
	}
	for _, it := range m.report.Repos {
		add(it)
	}
	return ordered
}

func fetchCountsCmd(id int, repos []Item, offset, size int, kind countKind) tea.Cmd {
	return func() tea.Msg {
		if offset >= len(repos) {
			return countsMsg{id: id, kind: kind}
		}
		end := offset + size
		if end > len(repos) {
			end = len(repos)
		}
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		var (
			counted []Item
			err     error
		)
		if kind == countCommits {
			counted, err = fetchRepoCommitBatch(ctx, execRunner, repos[offset:end])
		} else {
			counted, err = fetchRepoCountBatch(ctx, execRunner, repos[offset:end])
		}
		return countsMsg{id: id, repos: counted, kind: kind, err: err}
	}
}

func batchOffsets(n, size int) []int {
	if n <= 0 || size <= 0 {
		return nil
	}
	var out []int
	for i := 0; i < n; i += size {
		out = append(out, i)
	}
	return out
}

func fetchCommitsCmd(repo string, demo bool) tea.Cmd {
	return func() tea.Msg {
		if demo {
			return commitsMsg{repo: repo, commits: demoCommits(repo)}
		}
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()
		commits, err := fetchRecentCommits(ctx, execRunner, repo, 7)
		if err != nil {
			return commitsErrMsg{err}
		}
		return commitsMsg{repo: repo, commits: commits}
	}
}

func fetchOrgsCmd(demo bool) tea.Cmd {
	return func() tea.Msg {
		if demo {
			login, orgs := demoOwners()
			return orgsMsg{login: login, orgs: orgs}
		}
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()
		login, orgs, err := fetchOwnerChoices(ctx, execRunner)
		if err != nil {
			return orgsErrMsg{err}
		}
		return orgsMsg{login: login, orgs: orgs}
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(m.fetchCmd(), fetchOrgsCmd(m.demo))
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.ensureVisible()
		return m, nil

	case dataMsg:
		m.loading = false
		m.err = nil
		m.awaitRepos = false
		m.awaitActivity = false
		m.applyReport(msg.report.Repos, msg.report.Issues, msg.report.PRs)
		return m, nil

	case reposMsg:
		if msg.id != m.fetchID {
			return m, nil
		}
		m.awaitRepos = false
		if msg.err != nil {
			m.loading = false
			m.err = msg.err
			return m, nil
		}
		m.loading = false
		m.err = nil
		m.applyReport(msg.repos, m.report.Issues, m.report.PRs)
		if m.demo {
			return m, nil
		}
		m.awaitActivity = true
		m.countQueue = batchOffsets(len(m.report.Repos), countBatchSize)
		m.countInflight = 0
		return m, m.startCountWave(countIssuesPRs)

	case activityMsg:
		if msg.id != m.fetchID {
			return m, nil
		}
		if msg.err != nil {
			m.awaitActivity = false
			m.setStatus("issues/PRs: " + msg.err.Error())
			return m, nil
		}
		m.mergeActivity(msg.page)
		m.awaitActivity = false
		return m, m.nextRepoItemsCmd()

	case countsMsg:
		if msg.id != m.fetchID {
			return m, nil
		}
		if msg.kind == countCommits {
			m.commitInflight--
			if msg.err != nil {
				m.setStatus("commits: " + msg.err.Error())
			} else {
				m.applyReport(applyCommitCounts(m.report.Repos, msg.repos), m.report.Issues, m.report.PRs)
			}
			return m, m.startCountWave(countCommits)
		}
		m.countInflight--
		if msg.err != nil {
			m.awaitActivity = false
			m.setStatus("issues/PRs: " + msg.err.Error())
			return m, tea.Batch(m.nextRepoItemsCmd(), m.startCountWave(countIssuesPRs))
		}
		m.applyReport(applyIssuePRCounts(m.report.Repos, msg.repos), m.report.Issues, m.report.PRs)
		if cmd := m.startCountWave(countIssuesPRs); cmd != nil {
			return m, cmd
		}
		if m.countInflight > 0 {
			return m, nil
		}
		m.awaitActivity = false
		m.commitRepos = m.commitTargets()
		m.commitQueue = batchOffsets(len(m.commitRepos), commitBatchSize)
		m.commitInflight = 0
		return m, tea.Batch(m.nextRepoItemsCmd(), m.startCountWave(countCommits))

	case repoItemsMsg:
		if msg.id != m.fetchID {
			return m, nil
		}
		if m.loadingRepos != nil {
			delete(m.loadingRepos, msg.repo)
		}
		if msg.err != nil {
			m.setStatus(msg.err.Error())
			return m, m.nextRepoItemsCmd()
		}
		if m.loadedRepos == nil {
			m.loadedRepos = make(map[string]bool)
		}
		m.loadedRepos[msg.repo] = true
		m.mergeRepoItems(msg.repo, msg.issues, msg.prs)
		return m, m.nextRepoItemsCmd()

	case errMsg:
		m.loading = false
		m.err = msg.err
		return m, nil

	case orgsMsg:
		m.orgsLoading = false
		m.orgsErr = nil
		m.orgsLoaded = true
		m.login = msg.login
		m.orgs = msg.orgs
		m.orgCursor = m.currentOwnerIndex()
		return m, nil

	case orgsErrMsg:
		m.orgsLoading = false
		m.orgsErr = msg.err
		return m, nil

	case commitsMsg:
		if !m.showCommits || msg.repo != m.commitsRepo {
			return m, nil
		}
		m.commitsLoading = false
		m.commitsErr = nil
		m.commits = msg.commits
		m.commitCursor = 0
		return m, nil

	case commitsErrMsg:
		if !m.showCommits {
			return m, nil
		}
		m.commitsLoading = false
		m.commitsErr = msg.err
		return m, nil

	case statusClearMsg:
		if msg.id == m.statusID {
			m.status = ""
		}
		return m, nil

	case tea.KeyMsg:
		if m.urlMode {
			return m.handleURLKey(msg)
		}
		if m.filterMode {
			return m.handleFilterKey(msg)
		}
		return m.handleNormalKey(msg)
	}

	return m, nil
}

func (m model) handleFilterKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		m.filterMode = false
		m.filterText = ""
		m.rebuildRows()
		m.jumpToTop()
		return m, nil
	case tea.KeyEnter:
		m.filterMode = false
		return m, nil
	case tea.KeyBackspace:
		if len(m.filterText) > 0 {
			m.filterText = m.filterText[:len(m.filterText)-1]
			m.rebuildRows()
			m.jumpToTop()
		}
		return m, nil
	case tea.KeyRunes:
		m.filterText += string(msg.Runes)
		m.rebuildRows()
		m.jumpToTop()
		return m, nil
	case tea.KeyCtrlU:
		m.filterText = ""
		m.rebuildRows()
		m.jumpToTop()
		return m, nil
	}
	return m, nil
}

func (m model) handleURLKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		m.urlMode = false
		m.urlText = ""
		return m, nil
	case tea.KeyEnter:
		return m.submitURL()
	case tea.KeyBackspace:
		if len(m.urlText) > 0 {
			m.urlText = m.urlText[:len(m.urlText)-1]
		}
		return m, nil
	case tea.KeyRunes:
		m.urlText += string(msg.Runes)
		return m, nil
	case tea.KeyCtrlU:
		m.urlText = ""
		return m, nil
	}
	return m, nil
}

func (m model) submitURL() (tea.Model, tea.Cmd) {
	raw := strings.TrimSpace(m.urlText)
	m.urlMode = false
	m.urlText = ""
	if raw == "" {
		return m, nil
	}
	owner, repo, err := parseOwnerInput(raw)
	if err != nil {
		m.setStatus(err.Error())
		return m, nil
	}
	return m.applyOwner(owner, repo)
}

func (m model) handleNormalKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.showCommits {
		return m.handleCommitsKey(msg)
	}
	if m.showOrgs {
		return m.handleOrgPickerKey(msg)
	}

	if m.showHelp {
		if msg.String() == "?" || msg.String() == "esc" || msg.String() == "q" {
			m.showHelp = false
			return m, nil
		}
	}

	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit
	case "?":
		m.showHelp = !m.showHelp
		return m, nil
	}

	if m.loading {
		return m, nil
	}

	switch msg.String() {
	case "r":
		m.loading = true
		m.err = nil
		m.status = ""
		m.armFetch()
		return m, m.fetchCmd()

	case "/":
		m.filterMode = true
		m.filterText = ""
		return m, nil

	case "esc":
		if m.filterText != "" {
			m.filterText = ""
			m.rebuildRows()
			m.jumpToTop()
		}

	case "s":
		m.sortField = m.sortField.Next()
		m.rebuildRows()
		m.ensureVisible()
		m.setStatus(fmt.Sprintf("Sorted by %s", m.sortField))

	case "a":
		m.onlyPrivate = false
		m.onlyPublic = false
		m.onlyForks = false
		m.onlyNonForks = false
		m.rebuildRows()
		m.jumpToTop()
		m.setStatus("Showing all repos")

	case "p":
		m.onlyPublic = !m.onlyPublic
		if m.onlyPublic {
			m.onlyPrivate = false
		}
		m.rebuildRows()
		m.jumpToTop()
		if m.onlyPublic {
			m.setStatus("Public only")
		} else {
			m.setStatus("Showing public and private")
		}

	case "P", "shift+p":
		m.onlyPrivate = !m.onlyPrivate
		if m.onlyPrivate {
			m.onlyPublic = false
		}
		m.rebuildRows()
		m.jumpToTop()
		if m.onlyPrivate {
			m.setStatus("Private only")
		} else {
			m.setStatus("Showing public and private")
		}

	case "f":
		m.onlyNonForks = !m.onlyNonForks
		if m.onlyNonForks {
			m.onlyForks = false
		}
		m.rebuildRows()
		m.jumpToTop()
		if m.onlyNonForks {
			m.setStatus("Non-forks only")
		} else {
			m.setStatus("Showing forks and sources")
		}

	case "F", "shift+f":
		m.onlyForks = !m.onlyForks
		if m.onlyForks {
			m.onlyNonForks = false
		}
		m.rebuildRows()
		m.jumpToTop()
		if m.onlyForks {
			m.setStatus("Forks only")
		} else {
			m.setStatus("Showing forks and sources")
		}

	case "up", "k":
		m.move(-1)
	case "down", "j":
		m.move(1)
	case "left":
		m.setActiveExpanded(false)
	case "right":
		m.setActiveExpanded(true)
		return m, m.ensureRepoItems(m.activeRepoName())
	case "h":
		m.setAllExpanded(false)
	case "l":
		m.setAllExpanded(true)
		return m, m.nextRepoItemsCmd()
	case "pgup", "ctrl+u":
		m.move(-m.pageSize())
	case "pgdown", "ctrl+d":
		m.move(m.pageSize())
	case "home", "g":
		m.cursor = firstSelectable(m.rows, 0)
		m.ensureVisible()
	case "end", "G", "shift+g":
		m.cursor = lastSelectable(m.rows)
		m.ensureVisible()

	case "enter":
		return m, m.openSelected()

	case "o":
		return m.openOrgPicker()

	case "u":
		return m.openURLPrompt()

	case "U", "shift+u":
		if m.owner == "" {
			m.setStatus("Your repos")
			return m, nil
		}
		next, cmd := m.applyOwner("", "")
		n := next.(model)
		n.setStatus("Your repos")
		return n, cmd

	case "c":
		return m.openCommits()
	}

	return m, nil
}

func (m model) openCommits() (tea.Model, tea.Cmd) {
	repo := m.activeRepoName()
	if repo == "" {
		m.setStatus("No repo selected")
		return m, nil
	}
	m.showCommits = true
	m.commitsRepo = repo
	m.commits = nil
	m.commitsErr = nil
	m.commitCursor = 0
	m.commitsLoading = true
	return m, fetchCommitsCmd(repo, m.demo)
}

func (m model) handleCommitsKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit
	case "esc", "q", "c":
		m.showCommits = false
		m.commitsLoading = false
		return m, nil
	}

	if m.commitsLoading {
		return m, nil
	}

	switch msg.String() {
	case "up", "k":
		if m.commitCursor > 0 {
			m.commitCursor--
		}
	case "down", "j":
		if m.commitCursor < len(m.commits)-1 {
			m.commitCursor++
		}
	case "home", "g":
		m.commitCursor = 0
	case "end", "G", "shift+g":
		if n := len(m.commits); n > 0 {
			m.commitCursor = n - 1
		}
	case "enter":
		return m, m.openSelectedCommit()
	}
	return m, nil
}

func (m *model) openSelectedCommit() tea.Cmd {
	if m.commitCursor < 0 || m.commitCursor >= len(m.commits) {
		return nil
	}
	return m.openInBrowser(m.commits[m.commitCursor].URL)
}

func (m model) openOrgPicker() (tea.Model, tea.Cmd) {
	m.showOrgs = true
	m.orgsErr = nil
	if m.orgsLoaded {
		m.orgCursor = m.currentOwnerIndex()
		return m, nil
	}
	m.orgsLoading = true
	return m, fetchOrgsCmd(m.demo)
}

func (m model) handleOrgPickerKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit
	case "esc", "q":
		m.showOrgs = false
		m.orgsLoading = false
		return m, nil
	case "o":
		m.showOrgs = false
		return m, nil
	}

	if m.orgsLoading {
		return m, nil
	}

	choices := m.ownerChoices()
	switch msg.String() {
	case "up", "k":
		if m.orgCursor > 0 {
			m.orgCursor--
		}
	case "down", "j":
		if m.orgCursor < len(choices)-1 {
			m.orgCursor++
		}
	case "home", "g":
		m.orgCursor = 0
	case "end", "G", "shift+g":
		if n := len(choices); n > 0 {
			m.orgCursor = n - 1
		}
	case "enter":
		return m.selectOwner()
	}
	return m, nil
}

func (m model) ownerChoices() []string {
	choices := make([]string, 0, 1+len(m.orgs))
	if m.login != "" {
		choices = append(choices, m.login)
	}
	choices = append(choices, m.orgs...)
	return choices
}

func (m model) currentOwnerIndex() int {
	want := m.login
	if m.owner != "" {
		want = m.owner
	}
	for i, c := range m.ownerChoices() {
		if strings.EqualFold(c, want) {
			return i
		}
	}
	return 0
}

func (m model) selectOwner() (tea.Model, tea.Cmd) {
	choices := m.ownerChoices()
	if m.orgCursor < 0 || m.orgCursor >= len(choices) {
		m.showOrgs = false
		return m, nil
	}
	choice := choices[m.orgCursor]
	newOwner := choice
	if m.login != "" && strings.EqualFold(choice, m.login) {
		newOwner = ""
	}
	m.showOrgs = false
	return m.applyOwner(newOwner, "")
}

func (m model) openURLPrompt() (tea.Model, tea.Cmd) {
	m.urlMode = true
	m.urlText = m.owner
	if m.urlText == "" {
		m.urlText = m.login
	}
	return m, nil
}

func (m model) applyOwner(owner, focusRepo string) (tea.Model, tea.Cmd) {
	if m.login != "" && strings.EqualFold(owner, m.login) {
		owner = ""
	}
	if owner == m.owner {
		if focusRepo != "" {
			m.ensureExpandedMap()
			m.expandedRepos[focusRepo] = true
			m.rebuildRows()
			m.cursorToRepo(focusRepo)
			m.ensureVisible()
			return m, m.ensureRepoItems(focusRepo)
		}
		return m, nil
	}
	m.owner = owner
	m.pendingRepo = focusRepo
	m.loading = true
	m.err = nil
	m.status = ""
	m.rows = nil
	m.groups = nil
	m.report = Report{}
	m.cursor = 0
	m.scroll = 0
	m.armFetch()
	return m, m.fetchCmd()
}

func (m *model) applyReport(repos, issues, prs []Item) {
	prevURL := ""
	if sel := m.selected(); sel != nil {
		prevURL = sel.URL
	}
	m.report.Repos = repos
	m.report.Issues = issues
	m.report.PRs = prs
	if !m.demo {
		m.now = time.Now()
	}
	m.groups = buildGroups(m.report)
	m.rebuildRows()
	m.cursor = firstSelectable(m.rows, 0)
	if m.pendingRepo != "" {
		m.cursorToRepo(m.pendingRepo)
		m.ensureExpandedMap()
		m.expandedRepos[m.pendingRepo] = true
		m.rebuildRows()
		m.cursorToRepo(m.pendingRepo)
		m.pendingRepo = ""
	} else if prevURL != "" {
		for i, r := range m.rows {
			if r.selectable() && r.item.URL == prevURL {
				m.cursor = i
				break
			}
		}
	}
	m.ensureVisible()
}

func (m *model) mergeActivity(page activityPage) {
	m.counts = append(m.counts, page.Repos...)
	issues := append(append([]Item{}, m.report.Issues...), page.Issues...)
	prs := append(append([]Item{}, m.report.PRs...), page.PRs...)
	m.applyReport(applyIssuePRCounts(m.report.Repos, m.counts), issues, prs)
}

func groupIssueCount(g *repoGroup) int {
	if g.repo.IssueCount > len(g.issues) {
		return g.repo.IssueCount
	}
	return len(g.issues)
}

func groupPRCount(g *repoGroup) int {
	if g.repo.PRCount > len(g.prs) {
		return g.repo.PRCount
	}
	return len(g.prs)
}

func (m *model) mergeRepoItems(repo string, issues, prs []Item) {
	keep := func(items []Item) []Item {
		out := items[:0]
		for _, it := range items {
			if !strings.EqualFold(it.Repo, repo) {
				out = append(out, it)
			}
		}
		return out
	}
	m.report.Issues = append(keep(m.report.Issues), issues...)
	m.report.PRs = append(keep(m.report.PRs), prs...)
	m.groups = buildGroups(m.report)
	m.rebuildRows()
	m.ensureVisible()
}

func (m *model) repoNeedsItems(repo string) bool {
	if m.demo || repo == "" {
		return false
	}
	if m.loadedRepos[repo] || m.loadingRepos[repo] {
		return false
	}
	for _, g := range m.groups {
		if !strings.EqualFold(g.repo.Repo, repo) {
			continue
		}
		return g.repo.IssueCount > 0 || g.repo.PRCount > 0 || len(g.issues) > 0 || len(g.prs) > 0
	}
	return false
}

func (m model) ensureRepoItems(repo string) tea.Cmd {
	if !m.repoNeedsItems(repo) {
		return nil
	}
	if m.loadingRepos == nil {
		m.loadingRepos = make(map[string]bool)
	}
	m.loadingRepos[repo] = true
	id := m.fetchID
	demo := m.demo
	return func() tea.Msg {
		if demo {
			return repoItemsMsg{id: id, repo: repo}
		}
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()
		issues, prs, err := fetchRepoItems(ctx, execRunner, repo)
		return repoItemsMsg{id: id, repo: repo, issues: issues, prs: prs, err: err}
	}
}

func (m model) nextRepoItemsCmd() tea.Cmd {
	if m.demo {
		return nil
	}
	for repo, exp := range m.expandedRepos {
		if !exp {
			continue
		}
		if cmd := m.ensureRepoItems(repo); cmd != nil {
			return cmd
		}
	}
	return nil
}

func (m model) activeRepoName() string {
	if m.cursor < 0 || m.cursor >= len(m.rows) {
		return ""
	}
	r := m.rows[m.cursor]
	if r.repoName != "" {
		return r.repoName
	}
	return r.item.Repo
}

func (m *model) ensureExpandedMap() {
	if m.expandedRepos == nil {
		m.expandedRepos = make(map[string]bool)
	}
}

func (m *model) cursorToRepo(repo string) {
	for i, r := range m.rows {
		if r.typ == rowRepo && strings.EqualFold(r.item.Repo, repo) {
			m.cursor = i
			return
		}
	}
}

func (m *model) setActiveExpanded(expanded bool) {
	repo := m.activeRepoName()
	if repo == "" {
		return
	}
	m.ensureExpandedMap()
	m.expandedRepos[repo] = expanded
	m.rebuildRows()
	if !expanded {
		m.cursorToRepo(repo)
	}
	m.ensureVisible()
}

func (m *model) setAllExpanded(expanded bool) {
	m.ensureExpandedMap()
	if len(m.groups) == 0 && len(m.report.Repos) > 0 {
		m.groups = buildGroups(m.report)
	}
	for _, g := range m.groups {
		m.expandedRepos[g.repo.Repo] = expanded
	}
	m.rebuildRows()
	m.ensureVisible()
}

func (m *model) setStatus(msg string) {
	m.statusID++
	m.status = msg
}

func (m *model) openSelected() tea.Cmd {
	item := m.selected()
	if item == nil {
		return nil
	}
	return m.openInBrowser(item.URL)
}

func (m *model) openInBrowser(url string) tea.Cmd {
	if url == "" || m.demo {
		return nil
	}
	m.statusID++
	id := m.statusID
	m.status = "Opening " + url + " in browser…"
	return tea.Batch(
		func() tea.Msg {
			if err := openURL(url); err != nil {
				return errMsg{fmt.Errorf("open %s: %w", url, err)}
			}
			return nil
		},
		tea.Tick(3*time.Second, func(time.Time) tea.Msg {
			return statusClearMsg{id: id}
		}),
	)
}

func (m *model) move(delta int) {
	if len(m.rows) == 0 || delta == 0 {
		return
	}
	next := m.cursor
	steps := delta
	if steps < 0 {
		steps = -steps
	}
	dir := 1
	if delta < 0 {
		dir = -1
	}
	for i := 0; i < steps; i++ {
		cand := nextSelectable(m.rows, next, dir)
		if cand == next {
			break
		}
		next = cand
	}
	m.cursor = next
	m.ensureVisible()
}

func (m *model) pageSize() int {
	n := m.listHeight()
	if n < 1 {
		return 1
	}
	return n
}

func (m model) footerExtra() int {
	if m.err != nil || m.status != "" || m.filterMode || m.filterText != "" || m.urlMode {
		return 1
	}
	return 0
}

func (m model) showingTable() bool {
	return !m.loading && len(m.rows) > 0
}

func (m *model) listHeight() int {
	// Frame overhead: top, header-pinned, divider, inspector, help, bottom + optional status
	h := m.height - 6 - m.footerExtra()
	if h < 1 {
		return 1
	}
	return h
}

func (m *model) jumpToTop() {
	m.cursor = firstSelectable(m.rows, 0)
	m.scroll = 0
	m.ensureVisible()
}

func (m *model) ensureVisible() {
	if len(m.rows) == 0 {
		m.scroll = 0
		m.cursor = 0
		return
	}
	if m.cursor < 0 {
		m.cursor = 0
	}
	if m.cursor >= len(m.rows) {
		m.cursor = len(m.rows) - 1
	}
	h := m.listHeight()
	if m.cursor < m.scroll {
		m.scroll = m.cursor
	}
	if m.cursor >= m.scroll+h {
		m.scroll = m.cursor - h + 1
	}
	if m.scroll < 0 {
		m.scroll = 0
	}
}

func (m model) selected() *Item {
	if m.cursor < 0 || m.cursor >= len(m.rows) {
		return nil
	}
	if !m.rows[m.cursor].selectable() {
		return nil
	}
	item := m.rows[m.cursor].item
	return &item
}

func (m model) repoVisibilityOK(repo Item) bool {
	if m.onlyPrivate && !repo.Private {
		return false
	}
	if m.onlyPublic && repo.Private {
		return false
	}
	if m.onlyForks && !repo.Fork {
		return false
	}
	if m.onlyNonForks && repo.Fork {
		return false
	}
	return true
}

func (m model) visibleRepoStats() (repos, commits, issues, prs, stars int) {
	for _, r := range m.rows {
		if r.typ != rowRepo {
			continue
		}
		repos++
		commits += r.item.CommitCount
		issues += r.issuesCount
		prs += r.prsCount
		stars += r.item.Stars
	}
	return
}

func (m *model) rebuildRows() {
	if len(m.groups) == 0 && len(m.report.Repos) > 0 {
		m.groups = buildGroups(m.report)
	}

	sorted := make([]*repoGroup, len(m.groups))
	copy(sorted, m.groups)

	// Sort groups
	sort.SliceStable(sorted, func(i, j int) bool {
		a, b := sorted[i], sorted[j]
		switch m.sortField {
		case SortCommits:
			if a.repo.CommitCount != b.repo.CommitCount {
				return a.repo.CommitCount > b.repo.CommitCount
			}
			return a.repo.UpdatedAt.After(b.repo.UpdatedAt)
		case SortIssues:
			if ca, cb := groupIssueCount(a), groupIssueCount(b); ca != cb {
				return ca > cb
			}
			return a.repo.UpdatedAt.After(b.repo.UpdatedAt)
		case SortPRs:
			if ca, cb := groupPRCount(a), groupPRCount(b); ca != cb {
				return ca > cb
			}
			return a.repo.UpdatedAt.After(b.repo.UpdatedAt)
		case SortStars:
			if a.repo.Stars != b.repo.Stars {
				return a.repo.Stars > b.repo.Stars
			}
			return a.repo.UpdatedAt.After(b.repo.UpdatedAt)
		case SortName:
			return strings.ToLower(a.repo.Repo) < strings.ToLower(b.repo.Repo)
		default: // SortUpdated
			return a.repo.UpdatedAt.After(b.repo.UpdatedAt)
		}
	})

	query := strings.ToLower(strings.TrimSpace(m.filterText))

	var rows []row
	for _, g := range sorted {
		if !m.repoVisibilityOK(g.repo) {
			continue
		}

		repoMatches := query == "" || strings.Contains(strings.ToLower(g.repo.Repo), query) ||
			strings.Contains(strings.ToLower(g.repo.Description), query)

		matchingIssues := make([]Item, 0, len(g.issues))
		for _, iss := range g.issues {
			if query == "" || repoMatches ||
				strings.Contains(strings.ToLower(iss.Title), query) ||
				strings.Contains(fmt.Sprintf("%d", iss.Number), query) ||
				strings.Contains(strings.ToLower(iss.Author), query) {
				matchingIssues = append(matchingIssues, iss)
			}
		}

		matchingPRs := make([]Item, 0, len(g.prs))
		for _, pr := range g.prs {
			if query == "" || repoMatches ||
				strings.Contains(strings.ToLower(pr.Title), query) ||
				strings.Contains(fmt.Sprintf("%d", pr.Number), query) ||
				strings.Contains(strings.ToLower(pr.Author), query) {
				matchingPRs = append(matchingPRs, pr)
			}
		}

		if query != "" && !repoMatches && len(matchingIssues) == 0 && len(matchingPRs) == 0 {
			continue
		}

		expanded := m.expandedRepos[g.repo.Repo] || query != ""
		issCount := len(g.issues)
		if g.repo.IssueCount > issCount {
			issCount = g.repo.IssueCount
		}
		prCount := len(g.prs)
		if g.repo.PRCount > prCount {
			prCount = g.repo.PRCount
		}
		rows = append(rows, row{
			typ:         rowRepo,
			item:        g.repo,
			repoName:    g.repo.Repo,
			issuesCount: issCount,
			prsCount:    prCount,
			expanded:    expanded,
		})

		if expanded {
			children := make([]Item, 0, len(matchingIssues)+len(matchingPRs))
			children = append(children, matchingIssues...)
			children = append(children, matchingPRs...)

			for i, child := range children {
				typ := rowIssue
				if child.Kind == KindPR {
					typ = rowPR
				}
				rows = append(rows, row{
					typ:         typ,
					item:        child,
					repoName:    g.repo.Repo,
					isLastChild: i == len(children)-1,
				})
			}
		}
	}

	m.rows = rows
	if m.cursor >= len(m.rows) {
		m.cursor = len(m.rows) - 1
	}
	if m.cursor < 0 {
		m.cursor = 0
	}
}

func buildGroups(report Report) []*repoGroup {
	order := make([]string, 0, len(report.Repos))
	byRepo := make(map[string]*repoGroup, len(report.Repos))

	for _, repo := range report.Repos {
		order = append(order, repo.Repo)
		byRepo[repo.Repo] = &repoGroup{repo: repo}
	}

	for _, iss := range report.Issues {
		g, ok := byRepo[iss.Repo]
		if !ok {
			order = append(order, iss.Repo)
			g = &repoGroup{repo: Item{Kind: KindRepo, Repo: iss.Repo, Title: iss.Repo, URL: "https://github.com/" + iss.Repo}}
			byRepo[iss.Repo] = g
		}
		g.issues = append(g.issues, iss)
	}

	for _, pr := range report.PRs {
		g, ok := byRepo[pr.Repo]
		if !ok {
			order = append(order, pr.Repo)
			g = &repoGroup{repo: Item{Kind: KindRepo, Repo: pr.Repo, Title: pr.Repo, URL: "https://github.com/" + pr.Repo}}
			byRepo[pr.Repo] = g
		}
		g.prs = append(g.prs, pr)
	}

	groups := make([]*repoGroup, 0, len(order))
	for _, name := range order {
		groups = append(groups, byRepo[name])
	}
	return groups
}

func (m model) View() string {
	if m.width <= 0 {
		m.width = 80
	}
	if m.height <= 0 {
		m.height = 24
	}

	if m.showCommits {
		return m.renderCommitsModal()
	}
	if m.showOrgs {
		return m.renderOrgPicker()
	}
	if m.showHelp {
		return m.renderHelpModal()
	}

	bodyW := innerWidth(m.width)
	cw := computeColWidths(bodyW)

	_, commits, issues, prs, totalStars := m.visibleRepoStats()

	var titleRight string
	var topOverlay string
	var overlayAt int
	if m.loading {
		if m.owner != "" {
			titleRight = styleMuted.Render("fetching " + m.owner + "…")
		} else {
			titleRight = styleMuted.Render("fetching github activity…")
		}
	} else if m.awaitActivity {
		titleRight = styleMuted.Render("loading issue & PR counts…")
	} else if m.commitInflight > 0 || len(m.commitQueue) > 0 {
		titleRight = styleMuted.Render("loading commit counts…")
	}

	var pinned []string
	var body []string

	switch {
	case m.loading:
		scope := "your account"
		if m.owner != "" {
			scope = m.owner
		}
		body = []string{styleMuted.Render("  Fetching repos, issues, and pull requests for " + scope + "…")}
	case m.err != nil && len(m.rows) == 0:
		body = []string{styleError.Render("  Error: " + m.err.Error())}
	case len(m.rows) == 0:
		if m.filterText != "" {
			body = []string{styleMuted.Render(fmt.Sprintf("  No repositories matching %q (press Esc to clear)", m.filterText))}
		} else if m.onlyPrivate || m.onlyPublic || m.onlyForks || m.onlyNonForks {
			body = []string{styleMuted.Render("  No repositories match the current visibility filters")}
		} else {
			body = []string{styleMuted.Render("  No repos, issues, or pull requests found")}
		}
	default:
		topOverlay = renderCountStats(cw, commits, issues, prs, totalStars, m.commitsReady(), m.issuesReady())
		overlayAt = cw.typeCol + cw.gap + cw.repo + cw.gap
		pinned = []string{renderColHeader(cw, m.sortField, bodyW, m.commitsReady(), m.issuesReady())}
		h := m.listHeight()
		end := m.scroll + h
		if end > len(m.rows) {
			end = len(m.rows)
		}
		for i := m.scroll; i < end; i++ {
			body = append(body, m.renderRow(i, cw, bodyW))
		}
	}

	var foot []string

	// Filter or Status line
	if m.urlMode {
		foot = append(foot, styleSearch.Render(fmt.Sprintf(" User/org/repo: %s_", m.urlText))+" "+styleMuted.Render("[Enter] go · [Esc] cancel"))
	} else if m.filterMode {
		foot = append(foot, styleSearch.Render(fmt.Sprintf(" 🔍 Filter: %s_", m.filterText))+" "+styleMuted.Render("[Enter] keep · [Esc] clear"))
	} else if m.filterText != "" {
		foot = append(foot, styleSearch.Render(fmt.Sprintf(" 🔍 Filter: %s", m.filterText))+" "+styleMuted.Render(fmt.Sprintf("(%d rows) · [Esc] clear", len(m.rows))))
	} else if m.err != nil && len(m.rows) > 0 {
		foot = append(foot, styleError.Render(" ⚠ "+truncate(m.err.Error(), bodyW-4)))
	} else if m.status != "" {
		foot = append(foot, styleStatus.Render(" ✓ "+truncate(m.status, bodyW-4)))
	}

	// Inspector Line: rich details of selected item
	if sel := m.selected(); sel != nil {
		foot = append(foot, renderInspectorLine(*sel, bodyW, m.now))
	} else {
		foot = append(foot, styleMuted.Render(" —"))
	}

	// Keybind hints footer
	foldHint := "←→ un/fold · lh un/fold all"
	foot = append(foot, styleHelp.Render("↑↓/jk move · "+foldHint+" · enter open · c commits · o orgs · u url · U you · a all · p/P pub · f/F src · s sort · / filter · ? help · q quit"))

	titleLeft := styleTitle.Render("guh")
	who := m.owner
	if who == "" {
		who = m.login
	}
	if who != "" {
		titleLeft = styleTitle.Render("guh") + styleMuted.Render(" · "+who)
	}
	if m.onlyPrivate {
		titleLeft += styleMuted.Render(" · private")
	}
	if m.onlyPublic {
		titleLeft += styleMuted.Render(" · public")
	}
	if m.onlyForks {
		titleLeft += styleMuted.Render(" · forks")
	}
	if m.onlyNonForks {
		titleLeft += styleMuted.Render(" · sources")
	}
	return renderFrame(m.width, m.height, titleLeft, titleRight, topOverlay, overlayAt, pinned, body, foot)
}

func renderInspectorLine(item Item, width int, now time.Time) string {
	switch item.Kind {
	case KindIssue:
		author := ""
		if item.Author != "" {
			author = styleTextDim.Render("by @" + item.Author)
		}
		timeStr := styleMuted.Render(fmt.Sprintf("Update: %s (%s)", formatDaysOffset(item.UpdatedAt, now), formatRelativeTime(item.UpdatedAt, now)))
		prefix := styleIssue.Render(fmt.Sprintf("Issue #%d", item.Number))
		parts := []string{prefix}
		if author != "" {
			parts = append(parts, author)
		}
		parts = append(parts, timeStr)
		if item.Description == "" && item.URL != "" {
			parts = append(parts, styleURL.Render(item.URL))
		}
		return truncate(strings.Join(parts, " · "), width)

	case KindPR:
		author := ""
		if item.Author != "" {
			author = styleTextDim.Render("by @" + item.Author)
		}
		timeStr := styleMuted.Render(fmt.Sprintf("Update: %s (%s)", formatDaysOffset(item.UpdatedAt, now), formatRelativeTime(item.UpdatedAt, now)))
		prefix := stylePR.Render(fmt.Sprintf("PR #%d", item.Number))
		parts := []string{prefix}
		if author != "" {
			parts = append(parts, author)
		}
		parts = append(parts, timeStr)
		if item.Description == "" && item.URL != "" {
			parts = append(parts, styleURL.Render(item.URL))
		}
		return truncate(strings.Join(parts, " · "), width)

	default: // KindRepo
		var tags []string
		if item.Private {
			tags = append(tags, styleMuted.Render("🔒 Private"))
		}
		if item.Fork {
			tags = append(tags, styleMuted.Render("🍴 Fork"))
		}
		if item.Archived {
			tags = append(tags, styleMuted.Render("📦 Archived"))
		}
		if item.Language != "" {
			tags = append(tags, styleTextDim.Render(item.Language))
		}
		if item.Stars > 0 {
			tags = append(tags, styleMuted.Render(fmt.Sprintf("★ %d", item.Stars)))
		}
		if item.Description != "" {
			tags = append(tags, styleText.Render(item.Description))
		}
		tags = append(tags, styleMuted.Render(fmt.Sprintf("Update: %s (%s)", formatDaysOffset(item.UpdatedAt, now), formatRelativeTime(item.UpdatedAt, now))))
		if item.Description == "" && item.URL != "" {
			tags = append(tags, styleURL.Render(item.URL))
		}
		return truncate(strings.Join(tags, " · "), width)
	}
}

type colWidths struct {
	typeCol int
	repo    int
	commits int
	issues  int
	prs     int
	stars   int
	update  int
	gap     int
}

func computeColWidths(bodyWidth int) colWidths {
	gap := 2
	typeW := 4
	commitsW := 8
	issuesW := 8
	prsW := 5
	starsW := 7
	updateW := 8

	fixed := typeW + commitsW + issuesW + prsW + starsW + updateW + 6*gap
	repoW := bodyWidth - fixed
	if repoW < 16 {
		repoW = 16
	}

	return colWidths{
		typeCol: typeW,
		repo:    repoW,
		commits: commitsW,
		issues:  issuesW,
		prs:     prsW,
		stars:   starsW,
		update:  updateW,
		gap:     gap,
	}
}

func renderColStats(cw colWidths, repos, commits, issues, prs, stars, width int, commitsReady, issuesReady bool) string {
	typeCell := dashFill(cw.typeCol)
	// Indent with the same 2-cell cursor gutter used by repo rows.
	repoCell := embedTokenLeft(cw.repo, 2, styleMuted.Render(fmt.Sprintf("%d", repos)))
	line := typeCell +
		dashFill(cw.gap) + repoCell +
		dashFill(cw.gap) + renderCountStats(cw, commits, issues, prs, stars, commitsReady, issuesReady)
	return padDashes(line, width)
}

func renderCountStats(cw colWidths, commits, issues, prs, stars int, commitsReady, issuesReady bool) string {
	commitCell := embedTokenRight(cw.commits, styleCount(commits, styleCommit, fmt.Sprintf("%d", commits), commitsReady))
	issCell := embedTokenRight(cw.issues, styleCount(issues, styleIssue, fmt.Sprintf("%d", issues), issuesReady))
	prCell := embedTokenRight(cw.prs, styleCount(prs, stylePR, fmt.Sprintf("%d", prs), issuesReady))
	starsCell := embedTokenRight(cw.stars, styleCount(stars, styleStar, fmt.Sprintf("%d", stars), true))
	updCell := dashFill(cw.update)
	return commitCell +
		gapAfterToken(cw.gap) + issCell +
		gapAfterToken(cw.gap) + prCell +
		gapAfterToken(cw.gap) + starsCell +
		gapAfterToken(cw.gap) + updCell
}

func gapAfterToken(n int) string {
	if n <= 0 {
		return ""
	}
	return styleBorder.Render(" ") + dashFill(n-1)
}

func dashFill(n int) string {
	if n <= 0 {
		return ""
	}
	return styleBorder.Render(strings.Repeat("─", n))
}

func padDashes(s string, width int) string {
	if width <= 0 {
		return ""
	}
	w := lipgloss.Width(s)
	if w == width {
		return s
	}
	if w > width {
		return clipDisplay(s, width)
	}
	return s + dashFill(width-w)
}

// embedTokenLeft places styled text at pad, with a space on each side so it
// sits in the dash rule without colliding with neighboring ─ characters.
func embedTokenLeft(width, pad int, s string) string {
	sw := lipgloss.Width(s)
	if sw >= width {
		return truncate(s, width)
	}
	left := pad - 1
	if left < 0 {
		left = 0
	}
	need := 1 + sw + 1
	if left+need > width {
		rest := width - pad - sw
		if rest < 0 {
			return padDashes(dashFill(pad)+s, width)
		}
		return dashFill(pad) + s + dashFill(rest)
	}
	return dashFill(left) + styleBorder.Render(" ") + s + styleBorder.Render(" ") + dashFill(width-left-need)
}

func embedTokenRight(width int, s string) string {
	sw := lipgloss.Width(s)
	if sw >= width {
		return clipDisplay(s, width)
	}
	if sw+1 >= width {
		return dashFill(width-sw) + s
	}
	return dashFill(width-sw-1) + styleBorder.Render(" ") + s
}

func renderColHeader(cw colWidths, sortField SortField, width int, commitsReady, issuesReady bool) string {
	typeTitle := "TYPE"
	repoTitle := "REPO"
	if sortField == SortName {
		repoTitle = "REPO ▲"
	}

	commitsTitle := "COMMITS"
	if sortField == SortCommits {
		commitsTitle = "COMMITS▼"
	}

	issTitle := "ISSUES"
	if sortField == SortIssues {
		issTitle = "ISSUES ▼"
	}

	prsTitle := "PRS"
	if sortField == SortPRs {
		prsTitle = "PRS ▼"
	}

	starsTitle := "STARS"
	if sortField == SortStars {
		starsTitle = "STARS ▼"
	}

	updTitle := "UPDATE"
	if sortField == SortUpdated {
		updTitle = "UPDATE ▼"
	}

	gap := strings.Repeat(" ", cw.gap)
	line := styleHeader.Render(padTo(typeTitle, cw.typeCol)) +
		gap + headerCell(repoTitle, cw.repo, false, sortField == SortName, true) +
		gap + headerCell(commitsTitle, cw.commits, true, sortField == SortCommits, commitsReady) +
		gap + headerCell(issTitle, cw.issues, true, sortField == SortIssues, issuesReady) +
		gap + headerCell(prsTitle, cw.prs, true, sortField == SortPRs, issuesReady) +
		gap + headerCell(starsTitle, cw.stars, true, sortField == SortStars, true) +
		gap + headerCell(updTitle, cw.update, true, sortField == SortUpdated, true)

	return padTo(line, width)
}

func headerCell(title string, width int, right, sorted, ready bool) string {
	s := title
	if right {
		s = alignRight(title, width)
	} else {
		s = padTo(title, width)
	}
	if sorted {
		return styleHeaderSort.Render(s)
	}
	if !ready {
		return styleMuted.Render(s)
	}
	return styleHeader.Render(s)
}

func (m model) renderRow(i int, cw colWidths, width int) string {
	r := m.rows[i]
	selected := i == m.cursor

	if r.spacer || r.typ == rowSpacer {
		return strings.Repeat(" ", width)
	}
	if r.typ == rowIssue || r.typ == rowPR {
		return renderChildRow(r, selected, cw, width, m.now)
	}
	return renderRepoRow(r, selected, cw, width, m.now, m.commitsReady(), m.issuesReady())
}

func (m model) issuesReady() bool {
	return m.demo || (!m.loading && !m.awaitActivity)
}

func (m model) commitsReady() bool {
	return m.issuesReady() && m.commitInflight == 0 && len(m.commitQueue) == 0
}

func repoPrefix(r row, selected bool) string {
	if r.unfoldable() {
		if r.expanded {
			return "▾ "
		}
		return "▸ "
	}
	if selected {
		return "▸ "
	}
	return "  "
}

func renderRepoRow(r row, selected bool, cw colWidths, width int, now time.Time, commitsReady, issuesReady bool) string {
	cursor := repoPrefix(r, selected)

	typeStr := r.item.TypeEmoji()
	typeCell := padTo(typeStr, cw.typeCol)

	name := r.item.Repo
	if name == "" {
		name = r.item.Title
	}

	prefix := cursor
	nameBudget := cw.repo - lipgloss.Width(prefix)
	if nameBudget < 4 {
		nameBudget = 4
	}
	name = truncate(name, nameBudget)

	commitStr := fmt.Sprintf("%d", r.item.CommitCount)
	if r.item.CommitCount == 0 {
		commitStr = "0"
	}

	issStr := fmt.Sprintf("%d", r.issuesCount)
	if r.issuesCount == 0 {
		issStr = "0"
	}
	prStr := fmt.Sprintf("%d", r.prsCount)
	if r.prsCount == 0 {
		prStr = "0"
	}

	starsStr := fmt.Sprintf("%d", r.item.Stars)
	if r.item.Stars == 0 {
		starsStr = "0"
	}

	updStr := formatDaysOffset(r.item.UpdatedAt, now)

	if selected {
		prefixW := lipgloss.Width(prefix)
		nameCell := styleSelected.Render(padTo(name, cw.repo-prefixW))
		repoCell := styleCursor.Render(prefix) + nameCell
		rest := strings.Repeat(" ", cw.gap) + alignRight(commitStr, cw.commits) +
			strings.Repeat(" ", cw.gap) + alignRight(issStr, cw.issues) +
			strings.Repeat(" ", cw.gap) + alignRight(prStr, cw.prs) +
			strings.Repeat(" ", cw.gap) + alignRight(starsStr, cw.stars) +
			strings.Repeat(" ", cw.gap) + alignRight(updStr, cw.update)
		line := styleSelected.Render(typeCell+strings.Repeat(" ", cw.gap)) + repoCell + styleSelected.Render(rest)
		if w := lipgloss.Width(line); w < width {
			line += styleSelected.Render(strings.Repeat(" ", width-w))
		}
		return line
	}

	// Colored rendering for unselected row
	repoFormatted := prefix + styleText.Render(name)
	repoCell := padTo(repoFormatted, cw.repo)

	commitCell := alignRight(styleCount(r.item.CommitCount, styleCommit, commitStr, commitsReady), cw.commits)
	issCell := alignRight(styleCount(r.issuesCount, styleIssue, issStr, issuesReady), cw.issues)
	prCell := alignRight(styleCount(r.prsCount, stylePR, prStr, issuesReady), cw.prs)
	starsCell := alignRight(styleCount(r.item.Stars, styleStar, starsStr, true), cw.stars)
	updCell := alignRight(styleMuted.Render(updStr), cw.update)

	line := typeCell + strings.Repeat(" ", cw.gap) + repoCell + strings.Repeat(" ", cw.gap) + commitCell + strings.Repeat(" ", cw.gap) + issCell + strings.Repeat(" ", cw.gap) + prCell + strings.Repeat(" ", cw.gap) + starsCell + strings.Repeat(" ", cw.gap) + updCell
	return padTo(line, width)
}

func styleCount(n int, hi lipgloss.Style, text string, ready bool) string {
	if !ready {
		return styleMuted.Render(text)
	}
	return hi.Render(text)
}

func renderChildRow(r row, selected bool, cw colWidths, width int, now time.Time) string {
	branch := "├─ "
	if r.isLastChild {
		branch = "└─ "
	}
	prefix := "  " + branch
	if selected {
		prefix = styleCursor.Render("▸ ") + styleSelected.Render(branch)
	}

	// Child rows have empty TYPE column to align cleanly under repo name
	typeCell := strings.Repeat(" ", cw.typeCol)

	num := fmt.Sprintf("#%-4d", r.item.Number)
	title := strings.TrimSpace(r.item.Title)
	author := ""
	if r.item.Author != "" {
		author = "@" + r.item.Author
	}
	updStr := formatDaysOffset(r.item.UpdatedAt, now)

	if selected {
		left := styleSelected.Render(typeCell+strings.Repeat(" ", cw.gap)) + prefix + styleSelected.Render(num+" "+title)
		rightParts := []string{}
		if author != "" {
			rightParts = append(rightParts, author)
		}
		rightParts = append(rightParts, updStr)
		right := styleSelected.Render(strings.Join(rightParts, "  "))

		var line string
		if lipgloss.Width(left)+lipgloss.Width(right)+2 <= width {
			gap := width - lipgloss.Width(left) - lipgloss.Width(right)
			line = left + styleSelected.Render(strings.Repeat(" ", gap)) + right
		} else {
			line = truncate(left, width)
		}
		if w := lipgloss.Width(line); w < width {
			line += styleSelected.Render(strings.Repeat(" ", width-w))
		}
		return line
	}

	numStyle := styleIssue
	if r.typ == rowPR {
		numStyle = stylePR
	}

	left := typeCell + strings.Repeat(" ", cw.gap) + styleBranch.Render(prefix) + numStyle.Render(num) + " " + styleText.Render(title)

	var rightParts []string
	if author != "" {
		rightParts = append(rightParts, styleTextDim.Render(author))
	}
	rightParts = append(rightParts, styleMuted.Render(updStr))
	right := strings.Join(rightParts, "  ")

	avail := width - lipgloss.Width(right) - 2
	prefixLen := cw.typeCol + cw.gap + lipgloss.Width(prefix) + lipgloss.Width(num)
	if avail > prefixLen+4 {
		titleBudget := avail - prefixLen - 1
		left = typeCell + strings.Repeat(" ", cw.gap) + styleBranch.Render(prefix) + numStyle.Render(num) + " " + styleText.Render(truncate(title, titleBudget))
		gap := width - lipgloss.Width(left) - lipgloss.Width(right)
		if gap < 1 {
			gap = 1
		}
		return padTo(left+strings.Repeat(" ", gap)+right, width)
	}

	return padTo(truncate(left, width), width)
}

func (m model) renderHelpModal() string {
	boxW := 54
	if boxW > m.width-4 {
		boxW = m.width - 4
	}
	if boxW < 30 {
		boxW = 30
	}

	lines := []string{
		styleTitle.Render("guh — Keyboard Shortcuts"),
		"",
		styleHeaderSort.Render("Navigation:"),
		fmt.Sprintf("  %-16s %s", styleHelpKey.Render("↑ / k, ↓ / j"), "Move cursor up / down"),
		fmt.Sprintf("  %-16s %s", styleHelpKey.Render("g / G, Home/End"), "Jump to top / bottom"),
		fmt.Sprintf("  %-16s %s", styleHelpKey.Render("PgUp / PgDn"), "Scroll by page"),
		"",
		styleHeaderSort.Render("Actions & Folding:"),
		fmt.Sprintf("  %-16s %s", styleHelpKey.Render("←"), "Fold the current repo"),
		fmt.Sprintf("  %-16s %s", styleHelpKey.Render("→"), "Unfold the current repo"),
		fmt.Sprintf("  %-16s %s", styleHelpKey.Render("h"), "Fold all repos"),
		fmt.Sprintf("  %-16s %s", styleHelpKey.Render("l"), "Unfold all repos"),
		fmt.Sprintf("  %-16s %s", styleHelpKey.Render("Enter"), "Open repo, issue, PR, or commit in browser"),
		fmt.Sprintf("  %-16s %s", styleHelpKey.Render("c"), "Last 7 commits for the current repo"),
		"",
		styleHeaderSort.Render("View & Tools:"),
		fmt.Sprintf("  %-16s %s", styleHelpKey.Render("o"), "Switch organization"),
		fmt.Sprintf("  %-16s %s", styleHelpKey.Render("u"), "Open a user, org, repo, or GitHub URL"),
		fmt.Sprintf("  %-16s %s", styleHelpKey.Render("U"), "Reset to your account"),
		fmt.Sprintf("  %-16s %s", styleHelpKey.Render("a"), "Show all repos"),
		fmt.Sprintf("  %-16s %s", styleHelpKey.Render("p"), "Toggle public repos only"),
		fmt.Sprintf("  %-16s %s", styleHelpKey.Render("P"), "Toggle private repos only"),
		fmt.Sprintf("  %-16s %s", styleHelpKey.Render("f"), "Toggle non-forks only"),
		fmt.Sprintf("  %-16s %s", styleHelpKey.Render("F"), "Toggle forks only"),
		fmt.Sprintf("  %-16s %s", styleHelpKey.Render("s"), "Cycle sort (Update, Name, Commits, Issues, PRs, Stars)"),
		fmt.Sprintf("  %-16s %s", styleHelpKey.Render("/"), "Search & filter repos / issues / PRs"),
		fmt.Sprintf("  %-16s %s", styleHelpKey.Render("Esc"), "Clear search filter / close org picker"),
		fmt.Sprintf("  %-16s %s", styleHelpKey.Render("r"), "Refresh data from GitHub"),
		fmt.Sprintf("  %-16s %s", styleHelpKey.Render("?"), "Toggle this help modal"),
		fmt.Sprintf("  %-16s %s", styleHelpKey.Render("q / Ctrl+C"), "Quit"),
	}

	var framed []string
	framed = append(framed, styleBorder.Render("╭"+strings.Repeat("─", boxW-2)+"╮"))
	for _, l := range lines {
		framed = append(framed, styleBorder.Render("│")+" "+padTo(l, boxW-4)+" "+styleBorder.Render("│"))
	}
	framed = append(framed, styleBorder.Render("╰"+strings.Repeat("─", boxW-2)+"╯"))

	modalContent := strings.Join(framed, "\n")
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, modalContent)
}

func (m model) renderCommitsModal() string {
	boxW := 72
	if boxW > m.width-4 {
		boxW = m.width - 4
	}
	if boxW < 36 {
		boxW = 36
	}
	inner := boxW - 4

	title := "Recent commits"
	if m.commitsRepo != "" {
		title += " — " + m.commitsRepo
	}
	lines := []string{
		styleTitle.Render(truncate(title, inner)),
		"",
	}

	switch {
	case m.commitsLoading:
		lines = append(lines, styleMuted.Render("  Loading commits…"))
	case m.commitsErr != nil:
		lines = append(lines, styleError.Render("  "+truncate(m.commitsErr.Error(), inner-2)))
	case len(m.commits) == 0:
		lines = append(lines, styleMuted.Render("  No commits found"))
	default:
		for i, c := range m.commits {
			sha := styleMuted.Render(c.ShortSHA())
			desc := styleText.Render(truncate(c.Title, inner-12))
			head := sha + "  " + desc
			meta := c.Author
			if !c.Date.IsZero() {
				when := formatLocalDateTime(c.Date)
				if meta != "" {
					meta += " · " + when
				} else {
					meta = when
				}
			}
			if i == m.commitCursor {
				lines = append(lines, styleCursor.Render("▸ ")+styleSelected.Render(truncate(c.ShortSHA()+"  "+c.Title, inner-2)))
				if meta != "" {
					lines = append(lines, styleSelected.Render("  "+truncate(meta, inner-2)))
				}
			} else {
				lines = append(lines, "  "+head)
				if meta != "" {
					lines = append(lines, "  "+styleMuted.Render(truncate(meta, inner-2)))
				}
			}
			if i != len(m.commits)-1 {
				lines = append(lines, "")
			}
		}
	}

	lines = append(lines, "", styleHelp.Render("↑↓/jk move · enter open · esc close"))

	var framed []string
	framed = append(framed, styleBorder.Render("╭"+strings.Repeat("─", boxW-2)+"╮"))
	for _, l := range lines {
		framed = append(framed, styleBorder.Render("│")+" "+padTo(l, inner)+" "+styleBorder.Render("│"))
	}
	framed = append(framed, styleBorder.Render("╰"+strings.Repeat("─", boxW-2)+"╯"))
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, strings.Join(framed, "\n"))
}

func (m model) renderOrgPicker() string {
	boxW := 44
	if boxW > m.width-4 {
		boxW = m.width - 4
	}
	if boxW < 28 {
		boxW = 28
	}

	lines := []string{
		styleTitle.Render("Switch organization"),
		"",
	}

	switch {
	case m.orgsLoading:
		lines = append(lines, styleMuted.Render("  Loading organizations…"))
	case m.orgsErr != nil:
		lines = append(lines, styleError.Render("  "+truncate(m.orgsErr.Error(), boxW-6)))
	default:
		choices := m.ownerChoices()
		if len(choices) == 0 {
			lines = append(lines, styleMuted.Render("  No organizations found"))
		}
		for i, name := range choices {
			label := name
			if m.login != "" && strings.EqualFold(name, m.login) {
				label = name + "  (you)"
			}
			cursor := "  "
			row := label
			if i == m.orgCursor {
				cursor = "▸ "
				row = styleSelected.Render(label)
			} else if m.owner == "" && m.login != "" && strings.EqualFold(name, m.login) {
				row = styleText.Render(label)
			} else if m.owner != "" && strings.EqualFold(name, m.owner) {
				row = styleText.Render(label)
			}
			lines = append(lines, cursor+row)
		}
	}

	lines = append(lines, "", styleHelp.Render("↑↓/jk move · enter select · esc cancel"))

	var framed []string
	framed = append(framed, styleBorder.Render("╭"+strings.Repeat("─", boxW-2)+"╮"))
	for _, l := range lines {
		framed = append(framed, styleBorder.Render("│")+" "+padTo(l, boxW-4)+" "+styleBorder.Render("│"))
	}
	framed = append(framed, styleBorder.Render("╰"+strings.Repeat("─", boxW-2)+"╯"))
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, strings.Join(framed, "\n"))
}

func firstSelectable(rows []row, start int) int {
	if start < 0 {
		start = 0
	}
	for i := start; i < len(rows); i++ {
		if rows[i].selectable() {
			return i
		}
	}
	for i := start - 1; i >= 0; i-- {
		if rows[i].selectable() {
			return i
		}
	}
	return 0
}

func lastSelectable(rows []row) int {
	for i := len(rows) - 1; i >= 0; i-- {
		if rows[i].selectable() {
			return i
		}
	}
	return 0
}

func nextSelectable(rows []row, from, dir int) int {
	if dir == 0 || len(rows) == 0 {
		return from
	}
	i := from + dir
	for i >= 0 && i < len(rows) {
		if rows[i].selectable() {
			return i
		}
		i += dir
	}
	return from
}

func innerWidth(width int) int {
	w := width - 4
	if w < 8 {
		return 8
	}
	return w
}

func renderFrame(width, height int, titleLeft, titleRight, topOverlay string, overlayAt int, pinned, body, foot []string) string {
	if width < 24 {
		width = 24
	}
	if height < 8 {
		height = 8
	}

	lines := make([]string, 0, height)
	if topOverlay != "" {
		lines = append(lines, titledRuleWithOverlay(width, titleLeft, topOverlay, overlayAt))
	} else {
		lines = append(lines, titledRule(width, "╭", "╮", titleLeft, titleRight))
	}
	for _, line := range pinned {
		lines = append(lines, sideLine(width, line))
	}

	bodyH := height - 2 - len(foot) - 1 - len(pinned)
	if bodyH < 1 {
		bodyH = 1
	}
	for i := 0; i < bodyH; i++ {
		content := ""
		if i < len(body) {
			content = body[i]
		}
		lines = append(lines, sideLine(width, content))
	}
	lines = append(lines, hRule(width, "├", "┤"))
	for _, line := range foot {
		lines = append(lines, sideLine(width, line))
	}
	lines = append(lines, hRule(width, "╰", "╯"))
	return strings.Join(lines, "\n")
}

func titledRuleWithOverlay(width int, left, stats string, statsAt int) string {
	if width < 24 {
		width = 24
	}
	innerW := width - 2
	gap := styleBorder.Render(" ")
	leftBit := dashFill(1) + gap + left + gap
	leftW := lipgloss.Width(leftBit)
	// Body content is inset one cell more than the title inner (│ + space vs ╭).
	start := statsAt + 1
	if start < leftW {
		start = leftW
	}
	fill := start - leftW
	if fill < 0 {
		fill = 0
	}
	inner := leftBit + dashFill(fill) + stats
	if w := lipgloss.Width(inner); w < innerW {
		inner += dashFill(innerW - w)
	}
	return styleBorder.Render("╭") + inner + styleBorder.Render("╮")
}

func dropDisplayPrefix(s string, n int) string {
	if n <= 0 {
		return s
	}
	i := 0
	skipped := 0
	lastSGR := ""
	for i < len(s) && skipped < n {
		if s[i] == '\x1b' {
			j := strings.IndexByte(s[i:], 'm')
			if j < 0 {
				break
			}
			lastSGR = s[i : i+j+1]
			i += j + 1
			continue
		}
		_, size := utf8.DecodeRuneInString(s[i:])
		i += size
		skipped++
	}
	rest := s[i:]
	if rest == "" || lastSGR == "" || lastSGR == "\x1b[0m" {
		return rest
	}
	return lastSGR + rest
}

func clipDisplay(s string, width int) string {
	if width <= 0 {
		return ""
	}
	if lipgloss.Width(s) <= width {
		return s
	}
	i := 0
	kept := 0
	for i < len(s) && kept < width {
		if s[i] == '\x1b' {
			j := strings.IndexByte(s[i:], 'm')
			if j < 0 {
				break
			}
			i += j + 1
			continue
		}
		_, size := utf8.DecodeRuneInString(s[i:])
		i += size
		kept++
	}
	return s[:i]
}

func titledRule(width int, lc, rc, left, right string) string {
	inner := width - 2
	leftBit := dashFill(1) + styleBorder.Render(" ") + left + styleBorder.Render(" ")
	rightBit := dashFill(1)
	if right != "" {
		rightBit = styleBorder.Render(" ") + right + styleBorder.Render(" ") + dashFill(1)
	}
	fill := inner - lipgloss.Width(leftBit) - lipgloss.Width(rightBit)
	if fill < 1 {
		fill = 1
		if lipgloss.Width(leftBit)+1+lipgloss.Width(rightBit) > inner && right != "" {
			rightBit = styleBorder.Render("─")
			fill = inner - lipgloss.Width(leftBit) - lipgloss.Width(rightBit)
			if fill < 1 {
				fill = 1
			}
		}
	}
	return styleBorder.Render(lc) + leftBit + styleBorder.Render(strings.Repeat("─", fill)) + rightBit + styleBorder.Render(rc)
}

func hRule(width int, lc, rc string) string {
	inner := width - 2
	if inner < 1 {
		inner = 1
	}
	return styleBorder.Render(lc + strings.Repeat("─", inner) + rc)
}

func sideLine(width int, content string) string {
	bodyW := innerWidth(width)
	return styleBorder.Render("│") + " " + padTo(content, bodyW) + " " + styleBorder.Render("│")
}

func padTo(s string, width int) string {
	if width <= 0 {
		return ""
	}
	w := lipgloss.Width(s)
	if w == width {
		return s
	}
	if w > width {
		return truncate(s, width)
	}
	return s + strings.Repeat(" ", width-w)
}

func alignRight(s string, width int) string {
	w := lipgloss.Width(s)
	if w == width {
		return s
	}
	if w > width {
		return truncate(s, width)
	}
	return strings.Repeat(" ", width-w) + s
}

func truncate(s string, width int) string {
	if width <= 0 {
		return ""
	}
	if lipgloss.Width(s) <= width {
		return s
	}
	if width <= 1 {
		return "…"
	}
	var b strings.Builder
	used := 0
	for _, r := range s {
		rw := lipgloss.Width(string(r))
		if used+rw >= width {
			break
		}
		b.WriteRune(r)
		used += rw
	}
	return b.String() + "…"
}
