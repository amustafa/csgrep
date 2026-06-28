BINARY := csgrep
BIN_DIR := bin

.PHONY: build clean install link

build:
	@mkdir -p $(BIN_DIR)
	go build -o $(BIN_DIR)/$(BINARY) .

clean:
	rm -rf $(BIN_DIR)

install: build
	cp $(BIN_DIR)/$(BINARY) $(GOPATH)/bin/$(BINARY) 2>/dev/null || \
	cp $(BIN_DIR)/$(BINARY) $(HOME)/go/bin/$(BINARY)

link: build
	@mkdir -p $(HOME)/.local/bin
	ln -sf $(CURDIR)/$(BIN_DIR)/$(BINARY) $(HOME)/.local/bin/$(BINARY)
