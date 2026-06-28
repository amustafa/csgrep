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
	@if [ -z "$(LINK_DIR)" ]; then echo "Error: CSGREP_LINK_DIR is not set — copy .envrc.example to .envrc and configure it" >&2; exit 1; fi
	@mkdir -p $(LINK_DIR)
	ln -sf $(CURDIR)/$(BIN_DIR)/$(BINARY) $(LINK_DIR)/$(BINARY)
