VERSION := $(shell git describe --tags --match 'v[0-9]*.[0-9]*.[0-9]*' --always --dirty)
LDFLAGS := -ldflags "-X github.com/mahi160/flowd/internal/fw.Version=$(VERSION)"
BINARY  := fw

.PHONY: dashboard sync-plugin build build-all install install-all clean

# Build the dashboard single-file artifact embedded by Go.
dashboard:
	cd dashboard && pnpm install --frozen-lockfile && pnpm run build

# Copy plugin/flowd.lua into the embed tree so `go build` can include it.
# Run this whenever plugin/flowd.lua is edited.
sync-plugin:
	cp plugin/flowd.lua internal/fw/lua/flowd.lua

# Build Go only. This is what fw update needs after cloning a tagged release.
build: sync-plugin
	go build $(LDFLAGS) -o $(BINARY) ./cmd/fw

# Build dashboard artifact, then Go binary.
build-all: dashboard sync-plugin build

install: build
	cp $(BINARY) /usr/local/bin/$(BINARY)

install-all: build-all
	cp $(BINARY) /usr/local/bin/$(BINARY)

clean:
	rm -f $(BINARY)
