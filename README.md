# guh

Repos for the logged-in `gh` user (or an org), with open issues, PRs, stars, and last update.

Needs [GitHub CLI](https://cli.github.com/) authenticated (`gh auth login`). A TTY gets the interactive view; anything else dumps a table.

```sh
brew tap astrostl/guh https://github.com/astrostl/guh
brew install guh
```

From source: `make build && ./guh`

| key | |
| --- | --- |
| `↑↓` `jk` | move |
| `←→` | fold / unfold the current repo |
| `h` `l` | fold / unfold all |
| `enter` | open in the browser |
| `c` | last 7 commits |
| `o` | switch org |
| `a` | clear visibility filters |
| `p` `P` | public / private |
| `f` `F` | sources / forks |
| `s` | cycle sort (update, name, issues, PRs, stars) |
| `/` | text filter |
| `?` | help |
| `q` | quit |
