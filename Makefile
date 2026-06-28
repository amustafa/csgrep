BINARY := csgrep
BIN_DIR := bin
LINK_DIR ?= $(CSGREP_LINK_DIR)

.PHONY: build clean link setup

build:
	@mkdir -p $(BIN_DIR)
	go build -o $(BIN_DIR)/$(BINARY) .

clean:
	rm -rf $(BIN_DIR)

link: build
	@if [ -z "$(LINK_DIR)" ]; then \
		echo "" >&2; \
		echo "  CSGREP_LINK_DIR is not set." >&2; \
		echo "" >&2; \
		echo "  Quick fix:  cp .envrc.example .envrc && source .envrc" >&2; \
		echo "  Or inline:  CSGREP_LINK_DIR=~/.local/bin make link" >&2; \
		echo "" >&2; \
		exit 1; \
	fi
	@mkdir -p $(LINK_DIR)
	ln -sf $(CURDIR)/$(BIN_DIR)/$(BINARY) $(LINK_DIR)/$(BINARY)

setup:
	@if [ ! -f .envrc ]; then \
		cp .envrc.example .envrc; \
		echo "Created .envrc from .envrc.example"; \
	else \
		echo ".envrc already exists, skipping"; \
	fi
	@command -v direnv >/dev/null 2>&1 && direnv allow || echo "Run 'source .envrc' to load environment (or install direnv)"
	@echo ""
	@echo "Done. Next steps:"
	@echo "  make build    # compile"
	@echo "  make link     # symlink to your PATH"
