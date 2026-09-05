# Releasing guh

## Prerequisites

- `gh` CLI authenticated (`gh auth status`)
- Go toolchain installed
- `vhs` (and its `ffmpeg` / `ttyd` deps) for the demo GIF
- Push access to `astrostl/guh`

## Steps

### 1. Decide the version

Use [semantic versioning](https://semver.org/). Bump patch for small fixes, minor for new keys or behavior, major if existing keys or output change in a breaking way.

### 2. Sync the README with changes since the last release

Diff the TUI and GitHub client against the previous release tag and update `README.md` if behavior or install instructions changed:

```sh
git diff $(git describe --tags --abbrev=0) -- tui.go github.go main.go README.md
```

If the demo GIF changed, bump the `?v=` suffix on the README image to the new version so GitHub and browsers do not keep serving the previous colors:

```md
![guh](guh.gif?v=1.2.3)
```

Commit the README updates as part of the release commit in step 5.

### 3. Smoke-test the binary

Confirm the version stamp and that the test suite still passes. Always pass `VERSION=` here — a plain `make build` uses `git describe --tags --always --dirty`, so uncommitted files stamp as `vX.Y.Z-dirty` (visible in `--version` and `?`). That is fine for local work and goes away with a clean tree or an explicit `VERSION=`.

```sh
make test
make build VERSION=v1.2.3
./guh --version
# expect: guh v1.2.3
```

Do not proceed if `--version` prints `dev`, `-dirty`, a `git describe` suffix (`v1.2.3-1-gabc1234`), or anything other than `guh v1.2.3`.

### 4. Run the release target

`VERSION=vMAJOR.MINOR.PATCH` is required. `make release` (and `package-macos` / `checksums` / `update-formula`) refuse the default `git describe --dirty` stamp and reject any VERSION with a suffix. After packaging they check that the native dist binary prints `guh v1.2.3`.

```sh
make release VERSION=v1.2.3
```

Do not run `make release` without `VERSION=`. That used to stamp whatever `git describe --dirty` said, including `-dirty`.

This will:
- Verify `gofmt -s` formatting, `LICENSE` presence, and `go vet` (`lint`)
- Build `./guh` and record a new `guh.gif` from `demo.tape` (`vhs`)
- Cross-compile binaries for all platforms
- Package the macOS binaries into tarballs (`dist/guh-v1.2.3-darwin-{arm64,amd64}.tar.gz`)
- Compute SHA256 checksums
- Patch `Formula/guh.rb` in place with the new version, URLs, and SHA256s
- Confirm the native `dist/` binary `--version` matches `guh v1.2.3`

### 5. Commit and tag

Commit every file that belongs in the release (source, tests, README, formula, GIF), not only the formula. Do not tag a dirty tree.

```sh
git add -u
git status                 # review; no leftover unstaged files
git commit -m "Release v1.2.3"
test -z "$(git status --porcelain)"   # abort if anything is still dirty
git describe --tags --dirty | grep -q -- '-dirty$' && { echo "refusing dirty tree"; exit 1; }
git tag v1.2.3
test "$(git describe --tags --dirty)" = "v1.2.3"   # abort if the tag does not match a clean tree
git push origin master v1.2.3
```

If `git describe --dirty` ends in `-dirty`, stop: the working tree still has uncommitted files and you would be tagging an incomplete tree. After the tag exists it must equal `v1.2.3` with no suffix.

### 6. Create the GitHub release and upload artifacts

Include the new `guh.gif` so the release has the same demo as the README.

```sh
gh release create v1.2.3 \
  dist/guh-v1.2.3-darwin-arm64.tar.gz \
  dist/guh-v1.2.3-darwin-amd64.tar.gz \
  dist/guh-linux-amd64 \
  dist/guh-linux-arm64 \
  dist/guh-windows-amd64.exe \
  dist/guh-windows-arm64.exe \
  guh.gif \
  --title "v1.2.3" \
  --notes "Brief description of what changed."
```

If the release already exists, replace the GIF asset:

```sh
gh release upload v1.2.3 guh.gif --clobber
```

### 7. Verify Homebrew

```sh
brew update
brew upgrade guh
guh --version
```

If testing from scratch:

```sh
brew tap astrostl/guh https://github.com/astrostl/guh
brew install guh
guh --version
```

## What the Makefile targets do

| Target | Description |
|--------|-------------|
| `make help` | Print available targets (default goal — runs when you type just `make`) |
| `make build` | Build `./guh` for the current platform with version stamping |
| `make gif` | Build `./guh` and record `guh.gif` from `demo.tape` (needs `vhs`) |
| `make fmt` | Formats all Go files with `gofmt -s -w` |
| `make lint` | Checks `gofmt -s` compliance, LICENSE presence, and `go vet` |
| `make test` | Runs `go test ./...` |
| `make all` | Cross-compiles all platform binaries into `dist/` |
| `make package-macos` | Tars the macOS binaries into versioned `.tar.gz` files |
| `make checksums` | Runs `shasum -a 256` and writes `dist/checksums.txt` |
| `make update-formula` | Patches `Formula/guh.rb` with new version and SHA256s |
| `make release VERSION=v1.2.3` | Requires an exact `vMAJOR.MINOR.PATCH` (refuses `-dirty` / git describe). Runs lint, records the demo GIF, packages, verifies the stamped `--version`, and prints next steps |
| `make clean` | Removes `./guh` and `./dist` |

## How the Homebrew tap works

The formula lives at `Formula/guh.rb` in the main repo. There is no separate tap repo. Homebrew treats the main repo as a tap when users run:

```sh
brew tap astrostl/guh https://github.com/astrostl/guh
```

Each release must have the macOS tarballs uploaded to GitHub Releases before `brew install` will work — Homebrew downloads directly from the release asset URLs in the formula.
