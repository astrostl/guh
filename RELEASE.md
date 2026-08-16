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

Commit the README updates as part of the release commit in step 5.

### 3. Smoke-test the binary

Confirm the version stamp and that the test suite still passes.

```sh
make test
make build VERSION=v1.2.3
./guh --version
# expect: guh v1.2.3
```

Do not proceed if `--version` prints `dev` or a dirty git describe — the ldflags did not stamp.

### 4. Run the release target

```sh
make release VERSION=v1.2.3
```

This will:
- Verify `gofmt -s` formatting, `LICENSE` presence, and `go vet` (`lint`)
- Build `./guh` and record `guh.gif` from `demo.tape` (`vhs`)
- Cross-compile binaries for all platforms
- Package the macOS binaries into tarballs (`dist/guh-v1.2.3-darwin-{arm64,amd64}.tar.gz`)
- Compute SHA256 checksums
- Patch `Formula/guh.rb` in place with the new version, URLs, and SHA256s

### 5. Commit and tag

```sh
git add README.md Formula/guh.rb guh.gif
git commit -m "Release v1.2.3"
git tag v1.2.3
git push origin master v1.2.3
```

### 6. Create the GitHub release and upload artifacts

```sh
gh release create v1.2.3 \
  dist/guh-v1.2.3-darwin-arm64.tar.gz \
  dist/guh-v1.2.3-darwin-amd64.tar.gz \
  dist/guh-linux-amd64 \
  dist/guh-linux-arm64 \
  dist/guh-windows-amd64.exe \
  dist/guh-windows-arm64.exe \
  --title "v1.2.3" \
  --notes "Brief description of what changed."
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
| `make release` | Runs lint, records the demo GIF, then all of the above and prints next steps |
| `make clean` | Removes `./guh` and `./dist` |

## How the Homebrew tap works

The formula lives at `Formula/guh.rb` in the main repo. There is no separate tap repo. Homebrew treats the main repo as a tap when users run:

```sh
brew tap astrostl/guh https://github.com/astrostl/guh
```

Each release must have the macOS tarballs uploaded to GitHub Releases before `brew install` will work — Homebrew downloads directly from the release asset URLs in the formula.
