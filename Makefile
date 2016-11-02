VERSION := $(shell sh -c 'git describe --always --tags')
BRANCH := $(shell sh -c 'git rev-parse --abbrev-ref HEAD')
COMMIT := $(shell sh -c 'git rev-parse HEAD')
ifdef GOBIN
PATH := $(GOBIN):$(PATH)
else
PATH := $(subst :,/bin:,$(GOPATH))/bin:$(PATH)
endif

default: prepare build

# Only run the build (no dependency grabbing)
build:
	go install -ldflags \
		"-X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.branch=$(BRANCH)" ./...

build-for-docker:
	CGO_ENABLED=0 GOOS=linux go build -installsuffix cgo -o telegraf-signalfx-output -ldflags \
		"-s -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.branch=$(BRANCH)" \
		./main.go ./signalfx.go

# Get dependencies and use gdm to checkout changesets
prepare:
	go get github.com/sparrc/gdm
	gdm restore

.PHONY: build default
