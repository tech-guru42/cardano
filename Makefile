# Determine root directory
ROOT_DIR=$(shell dirname $(realpath $(firstword $(MAKEFILE_LIST))))

# Gather all .go files for use in dependencies below
GO_FILES=$(shell find $(ROOT_DIR) -name '*.go')

# Gather list of expected binaries
BINARIES=$(shell cd $(ROOT_DIR)/cmd && ls -1 | grep -v ^common)

.PHONY: build mod-tidy clean test swag-update

# Alias for building program binary
build: $(BINARIES)

mod-tidy:
	# Needed to fetch new dependencies and add them to go.mod
	go mod tidy

clean:
	rm -f $(BINARIES)

test:
	go test -v ./...

swag-update:
	swag init -g internal/api/api.go

# Build our program binaries
# Depends on GO_FILES to determine when rebuild is needed
$(BINARIES): mod-tidy $(GO_FILES)
	go build -o $(@) ./cmd/$(@)
