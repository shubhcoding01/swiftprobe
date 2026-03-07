# ============================================================
#  SwiftProbe Makefile
# ============================================================

APP     = swiftprobe
CMD     = ./cmd/swiftprobe
MODULE  = github.com/shubhcoding01/swiftprobe

# Detect OS for binary extension
ifeq ($(OS),Windows_NT)
    BINARY = $(APP).exe
    RM     = del /f
else
    BINARY = $(APP)
    RM     = rm -f
endif

# ── Build ────────────────────────────────────────────────────
.PHONY: build
build:
	@echo "[*] Building $(BINARY)..."
	go build -o $(BINARY) $(CMD)
	@echo "[+] Done → $(BINARY)"

# ── Build for all platforms ──────────────────────────────────
.PHONY: build-all
build-all:
	@echo "[*] Building for all platforms..."
	GOOS=linux   GOARCH=amd64 go build -o dist/$(APP)-linux-amd64   $(CMD)
	GOOS=windows GOARCH=amd64 go build -o dist/$(APP)-windows-amd64.exe $(CMD)
	GOOS=darwin  GOARCH=amd64 go build -o dist/$(APP)-macos-amd64   $(CMD)
	GOOS=darwin  GOARCH=arm64 go build -o dist/$(APP)-macos-arm64   $(CMD)
	@echo "[+] Binaries in dist/"

# ── Run ──────────────────────────────────────────────────────
.PHONY: run
run: build
	./$(BINARY) -u http://testphp.vulnweb.com -w wordlists/common.txt

# ── Test ─────────────────────────────────────────────────────
.PHONY: test
test:
	@echo "[*] Running tests..."
	go test ./... -v

# ── Test with coverage ───────────────────────────────────────
.PHONY: coverage
coverage:
	@echo "[*] Running tests with coverage..."
	go test ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html
	@echo "[+] Coverage report → coverage.html"

# ── Lint & Vet ───────────────────────────────────────────────
.PHONY: vet
vet:
	@echo "[*] Running go vet..."
	go vet ./...
	@echo "[+] No issues found"

# ── Tidy dependencies ────────────────────────────────────────
.PHONY: tidy
tidy:
	@echo "[*] Tidying dependencies..."
	go mod tidy
	@echo "[+] Done"

# ── Clean ────────────────────────────────────────────────────
.PHONY: clean
clean:
	@echo "[*] Cleaning..."
	$(RM) $(BINARY)
	$(RM) coverage.out coverage.html
	@echo "[+] Clean"

# ── Install globally ─────────────────────────────────────────
.PHONY: install
install:
	@echo "[*] Installing $(APP) globally..."
	go install $(CMD)
	@echo "[+] Installed — run: $(APP) --help"

# ── Quick local test (spins up Python server) ────────────────
.PHONY: localtest
localtest: build
	@echo "[*] Setting up local test server..."
	mkdir -p testserver/admin testserver/api/v1
	echo "admin page"      > testserver/admin/index.html
	echo '{"status":"ok"}' > testserver/api/v1/health
	@echo "[*] Starting server on :8080 — press Ctrl+C to stop"
	cd testserver && python3 -m http.server 8080 &
	sleep 1
	./$(BINARY) -u http://localhost:8080 -w wordlists/common.txt -mc 200,301,404 -v

# ── Help ─────────────────────────────────────────────────────
.PHONY: help
help:
	@echo ""
	@echo "  SwiftProbe — Makefile Commands"
	@echo "  ──────────────────────────────"
	@echo "  make build       Build binary for current OS"
	@echo "  make build-all   Build for Linux, Windows, macOS"
	@echo "  make run         Build and run against test target"
	@echo "  make test        Run all unit tests"
	@echo "  make coverage    Run tests and generate HTML coverage report"
	@echo "  make vet         Run go vet for static analysis"
	@echo "  make tidy        Run go mod tidy"
	@echo "  make clean       Remove build artifacts"
	@echo "  make install     Install binary globally via go install"
	@echo "  make localtest   Spin up local server and fuzz it"
	@echo ""