# Flint Makefile
# ──────────────────────────────────────────────────────────────────────────────
# Targets:
#   make build       Build flint + flintd for the current platform
#   make install     Build and install to GOBIN (or /usr/local/bin)
#   make test        Run the test suite
#   make vet         Run go vet
#   make lint        Run staticcheck (install with: go install honnef.co/go/tools/cmd/staticcheck@latest)
#   make dist        Cross-compile all platforms → dist/
#   make clean       Remove dist/ and local binaries
#   make release v=  Tag a release: make release v=0.2.0

VERSION ?= dev
LDFLAGS  := -s -w -X main.version=$(VERSION)
DIST     := dist
BINARY   := flint
DAEMON   := flintd

.PHONY: build install test vet lint dist clean release check

# ── Local build ───────────────────────────────────────────────────────────────

build:
	go build -ldflags="$(LDFLAGS)" -trimpath -o $(BINARY) ./core/cmd/flint/
	go build -ldflags="$(LDFLAGS)" -trimpath -o $(DAEMON) ./core/cmd/daemon/
	@echo "Built: ./$(BINARY)  ./$(DAEMON)"

install:
	go install -ldflags="$(LDFLAGS)" -trimpath ./core/cmd/flint/
	go install -ldflags="$(LDFLAGS)" -trimpath ./core/cmd/daemon/
	@echo "Installed flint and flintd to $$(go env GOPATH)/bin"

# ── Quality ───────────────────────────────────────────────────────────────────

test:
	go test -race -count=1 ./...

vet:
	go vet ./...

lint:
	staticcheck ./...

check: vet test

# ── Cross-compile ─────────────────────────────────────────────────────────────

TARGETS := \
	darwin/amd64 \
	darwin/arm64 \
	linux/amd64 \
	linux/arm64 \
	windows/amd64

dist: clean
	@mkdir -p $(DIST)
	@for target in $(TARGETS); do \
		os=$${target%/*}; arch=$${target#*/}; \
		ext=""; [ "$$os" = "windows" ] && ext=".exe"; \
		echo "  flint  $$os/$$arch"; \
		GOOS=$$os GOARCH=$$arch go build \
			-ldflags="$(LDFLAGS)" -trimpath \
			-o $(DIST)/flint-$$os-$$arch$$ext \
			./core/cmd/flint/; \
		echo "  flintd $$os/$$arch"; \
		GOOS=$$os GOARCH=$$arch go build \
			-ldflags="$(LDFLAGS)" -trimpath \
			-o $(DIST)/flintd-$$os-$$arch$$ext \
			./core/cmd/daemon/; \
	done
	@echo ""
	@echo "Artifacts ($(shell ls $(DIST) | wc -l | tr -d ' ')) → $(DIST)/"
	@ls -lh $(DIST)

# ── Release ───────────────────────────────────────────────────────────────────

release:
ifndef v
	$(error Usage: make release v=x.y.z)
endif
	@echo "Tagging v$(v) …"
	git tag -a "v$(v)" -m "Release v$(v)"
	git push origin "v$(v)"
	@echo "Tag pushed — GitHub Actions will build and publish the release."

# ── Cleanup ───────────────────────────────────────────────────────────────────

clean:
	rm -rf $(DIST) $(BINARY) $(DAEMON) $(BINARY).exe $(DAEMON).exe
