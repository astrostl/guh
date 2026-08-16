# guh

A terminal UI for the GitHub repos your `gh` session can see. Yours or an org: open issues and PRs, stars, last update. Fold a repo to read what's open, flip through recent commits, jump to a user, org, or repo.

![guh](guh.gif)

Needs an authenticated [GitHub CLI](https://cli.github.com/). Run it in a terminal for the UI. Pipe or redirect and you get a plain table.

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
