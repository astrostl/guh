# CLAUDE.md

Developer notes for agents working in this repo.

## Always use the Makefile

Use `make` for any target it already has (`build`, `test`, `fmt`, `lint`,
`gif`, release flow, etc.). Do not invoke the underlying `go build` /
`go test` / `gofmt` / `go vet` commands directly when a Makefile target
covers them.

## Always build after changes

After any code change, run `make build` before considering the work done. That
stamps `./guh` for the current platform so the local binary matches the tree.

```sh
make build
```

Do not skip this even if tests already passed. Tests do not produce `./guh`.

## Commands

```sh
make build    # build ./guh (required after changes)
make test     # go test ./...
make fmt      # gofmt -s -w .
make lint     # gofmt check, LICENSE, go vet
make gif      # record guh.gif from demo.tape (needs vhs)
```

See [RELEASE.md](RELEASE.md) for packaging. Local `make build` uses
`git describe --tags --always --dirty`; pass `VERSION=vX.Y.Z` only for
release stamps.
