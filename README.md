# guh

A terminal UI for the GitHub repos your `gh` session can see. Yours or an org: commits, open issues and PRs, stars, last update. Fold a repo to read what's open, flip through recent commits, jump to a user, org, or repo.

![guh](guh.gif?v=0.9.0)

Needs an authenticated [GitHub CLI](https://cli.github.com/). Run it in a terminal for the UI. Pipe or redirect and you get a plain table.

Up to 1000 repos per account or org, newest first. The org picker lists up to 1000 orgs. Unfolding a repo loads every open issue and PR. Archived repos are hidden by default (`A` shows them, `a` active only, `*` all) and marked `⊟` in the type column. `c` shows the last 7 commits. `t` opens a shell in `~/src/<repo>` if that directory exists; Ctrl-D returns.

Homebrew (macOS):

```sh
brew tap astrostl/guh https://github.com/astrostl/guh
brew install guh
```

Go:

```sh
go install github.com/astrostl/guh@latest
```

`guh --demo` (or `-demo`) loads fake repos, issues, orgs, and commits. No `gh` session.

[![guh](https://img.youtube.com/vi/hj1N6vkOQDM/hqdefault.jpg)](https://www.youtube.com/shorts/hj1N6vkOQDM)
