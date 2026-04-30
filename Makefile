VERSION := $(shell git describe --tags --match 'v[0-9]*.[0-9]*.[0-9]*' --always --dirty)
LDFLAGS := -ldflags "-X github.com/mahi160/flowd/internal/fw.Version=$(VERSION)"
BINARY  := fw

.PHONY: build install clean

build:
	go build $(LDFLAGS) -o $(BINARY) ./cmd/fw

install:
	go build $(LDFLAGS) -o $(BINARY) ./cmd/fw
	sudo mv $(BINARY) /usr/local/bin/$(BINARY)

clean:
	rm -f $(BINARY)
