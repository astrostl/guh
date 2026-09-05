# Local builds may use git describe --dirty (so uncommitted work shows as -dirty).
# Packaging/release must pass VERSION=vMAJOR.MINOR.PATCH on the command line.
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS := -s -w -X main.version=$(VERSION)
BIN := guh
DIST := dist

RELEASE_GOALS := release update-formula package-macos checksums
ifneq ($(filter $(RELEASE_GOALS),$(MAKECMDGOALS)),)
ifeq ($(filter command line environment,$(origin VERSION)),)
$(error pass VERSION=vX.Y.Z on the command line; refusing to package git describe --dirty (see RELEASE.md))
endif
endif

.PHONY: help build clean all test lint fmt release package-macos checksums update-formula gif check-release-version verify-release-version

.DEFAULT_GOAL := help

help:
	@echo "Usage: make <target>"
	@echo ""
	@echo "Build:"
	@echo "  build                    Build ./guh for the current platform with version stamping"
	@echo "  gif                      Record guh.gif from demo.tape (needs vhs)"
	@echo "  all                      Cross-compile binaries for all platforms into dist/"
	@echo "  clean                    Remove ./guh and ./dist"
	@echo ""
	@echo "Quality:"
	@echo "  fmt                      Auto-format Go files (gofmt -s)"
	@echo "  lint                     Verify gofmt -s, LICENSE, and go vet"
	@echo "  test                     Run go test ./..."
	@echo ""
	@echo "Release (see RELEASE.md):"
	@echo "  release VERSION=v1.2.3   Lint, record GIF, cross-build, package, checksum, patch formula"
	@echo "                            VERSION is required and must be vMAJOR.MINOR.PATCH (no -dirty)"
	@echo "  package-macos            Tar dist/ macOS binaries into versioned .tar.gz files"
	@echo "  checksums                shasum -a 256 the macOS tarballs into dist/checksums.txt"
	@echo "  update-formula           Patch Formula/guh.rb with new version + SHA256s"

build:
	go build -ldflags "$(LDFLAGS)" -o $(BIN) .

gif: build
	@command -v vhs >/dev/null || { echo "vhs is required to record guh.gif (brew install vhs)"; exit 1; }
	vhs demo.tape

fmt:
	gofmt -s -w .

lint:
	@out=$$(gofmt -s -l .); [ -z "$$out" ] || { echo "gofmt -s issues in: $$out"; exit 1; }
	@test -f LICENSE || { echo "LICENSE file missing"; exit 1; }
	go vet ./...

test:
	go test ./...

all: clean
	mkdir -p $(DIST)
	GOOS=darwin  GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o $(DIST)/$(BIN)-darwin-amd64 .
	GOOS=darwin  GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o $(DIST)/$(BIN)-darwin-arm64 .
	GOOS=linux   GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o $(DIST)/$(BIN)-linux-amd64 .
	GOOS=linux   GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o $(DIST)/$(BIN)-linux-arm64 .
	GOOS=windows GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o $(DIST)/$(BIN)-windows-amd64.exe .
	GOOS=windows GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o $(DIST)/$(BIN)-windows-arm64.exe .

check-release-version:
	@echo "$(VERSION)" | grep -Eq '^v[0-9][0-9]*\.[0-9][0-9]*\.[0-9][0-9]*$$' || { \
		echo "error: VERSION must be vMAJOR.MINOR.PATCH with no suffix (got '$(VERSION)')"; \
		exit 1; \
	}

verify-release-version: check-release-version
	@native="$(DIST)/$(BIN)-darwin-arm64"; \
	if [ "$$(uname -s)" = "Darwin" ] && [ "$$(uname -m)" != "arm64" ]; then \
		native="$(DIST)/$(BIN)-darwin-amd64"; \
	fi; \
	if [ ! -x "$$native" ]; then \
		echo "error: missing native dist binary $$native"; \
		exit 1; \
	fi; \
	got=$$($$native --version); \
	want="guh $(VERSION)"; \
	if [ "$$got" != "$$want" ]; then \
		echo "error: $$native --version is '$$got', want '$$want'"; \
		exit 1; \
	fi; \
	echo "stamped $$native -> $$got"

package-macos: check-release-version all
	cd $(DIST) && tar czf $(BIN)-$(VERSION)-darwin-arm64.tar.gz $(BIN)-darwin-arm64
	cd $(DIST) && tar czf $(BIN)-$(VERSION)-darwin-amd64.tar.gz $(BIN)-darwin-amd64

checksums: check-release-version package-macos
	cd $(DIST) && shasum -a 256 $(BIN)-$(VERSION)-darwin-arm64.tar.gz $(BIN)-$(VERSION)-darwin-amd64.tar.gz | tee checksums.txt

update-formula: check-release-version checksums
	$(eval ARM64_SHA := $(shell grep darwin-arm64 $(DIST)/checksums.txt | awk '{print $$1}'))
	$(eval AMD64_SHA := $(shell grep darwin-amd64 $(DIST)/checksums.txt | awk '{print $$1}'))
	sed -i '' 's|version ".*"|version "$(VERSION)"|g' Formula/$(BIN).rb
	sed -i '' 's|releases/download/v[^/]*/$(BIN)-v[^-]*-darwin-arm64|releases/download/$(VERSION)/$(BIN)-$(VERSION)-darwin-arm64|g' Formula/$(BIN).rb
	sed -i '' 's|releases/download/v[^/]*/$(BIN)-v[^-]*-darwin-amd64|releases/download/$(VERSION)/$(BIN)-$(VERSION)-darwin-amd64|g' Formula/$(BIN).rb
	awk '/darwin-arm64/{found_arm64=1} found_arm64 && /sha256/ && !done_arm64{sub(/sha256 "[^"]*"/, "sha256 \"$(ARM64_SHA)\""); done_arm64=1} /darwin-amd64/{found_amd64=1} found_amd64 && /sha256/ && !done_amd64{sub(/sha256 "[^"]*"/, "sha256 \"$(AMD64_SHA)\""); done_amd64=1} {print}' Formula/$(BIN).rb > Formula/$(BIN).rb.tmp && mv Formula/$(BIN).rb.tmp Formula/$(BIN).rb

# Full release flow: make release VERSION=v1.2.3
# Then: git tag v1.2.3 && git push origin v1.2.3
# Then upload dist/ tarballs to the GitHub release
release: check-release-version lint gif update-formula verify-release-version
	@echo "Formula updated for $(VERSION). Next steps:"
	@echo "  1. git add -u && git commit -m 'Release $(VERSION)'"
	@echo "  2. Confirm git status is clean and git describe --tags --dirty has no -dirty suffix"
	@echo "  3. git tag $(VERSION) && git push origin master $(VERSION)"
	@echo "  4. Upload $(DIST)/$(BIN)-$(VERSION)-darwin-*.tar.gz and guh.gif to the GitHub release"

clean:
	rm -rf $(BIN) $(DIST)
