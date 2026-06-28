BINARY := csgrep
BIN_DIR := bin
LINK_DIR ?= $(CSGREP_LINK_DIR)

.PHONY: build clean link

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
