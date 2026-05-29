.PHONY: all build ci prepare clean \
	go-build go-full-build \
	go-build-linux-amd64 go-build-linux-arm64 \
	go-build-windows-amd64 go-build-darwin-amd64 go-build-darwin-arm64

GO ?= go
BIN_NAME ?= wg_vanity
DIST_DIR ?= dist
GO_DIR ?= go

# =========================
# Host Information
# =========================
HOST_GOOS := $(shell $(GO) env GOOS)
HOST_GOARCH := $(shell $(GO) env GOARCH)

# =========================
# Default Target
# =========================
all: build

build: go-build

# =========================
# Prepare
# =========================
prepare:
	@mkdir -p $(DIST_DIR)

# =========================
# Local Build
# =========================
go-build: prepare
	@echo "Building for $(HOST_GOOS)/$(HOST_GOARCH) ..."
	@cd $(GO_DIR) && $(GO) build -v \
		-o ../$(DIST_DIR)/$(BIN_NAME)-$(HOST_GOOS)-$(HOST_GOARCH) .
	@echo "Build complete: $(DIST_DIR)/$(BIN_NAME)-$(HOST_GOOS)-$(HOST_GOARCH)"

# =========================
# Linux Build
# =========================
go-build-linux-amd64: prepare
	@echo "Building for linux/amd64 ..."
	@cd $(GO_DIR) && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GO) build -trimpath \
		-o ../$(DIST_DIR)/$(BIN_NAME)-linux-amd64 .
	@echo "Build complete: $(DIST_DIR)/$(BIN_NAME)-linux-amd64"

go-build-linux-arm64: prepare
	@echo "Building for linux/arm64 ..."
	@cd $(GO_DIR) && CGO_ENABLED=0 GOOS=linux GOARCH=arm64 $(GO) build -trimpath \
		-o ../$(DIST_DIR)/$(BIN_NAME)-linux-arm64 .
	@echo "Build complete: $(DIST_DIR)/$(BIN_NAME)-linux-arm64"

# =========================
# Windows Build
# =========================
go-build-windows-amd64: prepare
	@echo "Building for windows/amd64 ..."
	@cd $(GO_DIR) && CGO_ENABLED=0 GOOS=windows GOARCH=amd64 $(GO) build -trimpath \
		-o ../$(DIST_DIR)/$(BIN_NAME)-windows-amd64.exe .
	@echo "Build complete: $(DIST_DIR)/$(BIN_NAME)-windows-amd64.exe"

# =========================
# macOS Build
# =========================
go-build-darwin-amd64: prepare
	@echo "Building for darwin/amd64 ..."
	@cd $(GO_DIR) && CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 $(GO) build -trimpath \
		-o ../$(DIST_DIR)/$(BIN_NAME)-darwin-amd64 .
	@echo "Build complete: $(DIST_DIR)/$(BIN_NAME)-darwin-amd64"

go-build-darwin-arm64: prepare
	@echo "Building for darwin/arm64 ..."
	@cd $(GO_DIR) && CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 $(GO) build -trimpath \
		-o ../$(DIST_DIR)/$(BIN_NAME)-darwin-arm64 .
	@echo "Build complete: $(DIST_DIR)/$(BIN_NAME)-darwin-arm64"

# =========================
# Full Platform Build
# =========================
go-full-build: \
	go-build-linux-amd64 \
	go-build-linux-arm64 \
	go-build-windows-amd64 \
	go-build-darwin-amd64 \
	go-build-darwin-arm64
	@echo "All builds completed successfully!"

# =========================
# Clean
# =========================
clean:
	@echo "Cleaning $(DIST_DIR) ..."
	@rm -rf $(DIST_DIR)
	@echo "Clean complete"

# =========================
# CI
# =========================
ci: build