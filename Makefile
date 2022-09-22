TEST?=$$(go list ./... | grep -v 'vendor')
HOSTNAME=registry.terraform.io
NAMESPACE=redhat-developer
NAME=rhoas
BINARY=terraform-provider-rhoas
VERSION ?= 0.1


ifeq ($(OS),Windows_NT)
	OS_ARCH ?=windows_amd64
else
    UNAME_S := $(shell uname -s)
    ifeq ($(UNAME_S),Linux)
    	OS_ARCH ?=linux
    endif
    ifeq ($(UNAME_S),Darwin)
        OS_ARCH ?=darwin
    endif

    UNAME_P := $(shell uname -p)
    ifeq ($(UNAME_P),x86_64)
       	OS_ARCH := $(OS_ARCH)_amd64
    endif
    ifneq ($(filter arm%,$(UNAME_P)),)
        OS_ARCH := $(OS_ARCH)_arm64
    endif
endif

default: install

.PHONY: build
build:
	go build -o ${BINARY}

.PHONY: docs
docs:
	go install github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs@latest
	tfplugindocs

.PHONY: release
release:
	GOOS=darwin GOARCH=amd64 go build -o ./bin/${BINARY}_${VERSION}_darwin_amd64
	GOOS=linux GOARCH=386 go build -o ./bin/${BINARY}_${VERSION}_linux_386
	GOOS=linux GOARCH=amd64 go build -o ./bin/${BINARY}_${VERSION}_linux_amd64
	GOOS=linux GOARCH=arm go build -o ./bin/${BINARY}_${VERSION}_linux_arm
	GOOS=windows GOARCH=386 go build -o ./bin/${BINARY}_${VERSION}_windows_386
	GOOS=windows GOARCH=amd64 go build -o ./bin/${BINARY}_${VERSION}_windows_amd64

.PHONY: install
install: build
	mkdir -p ~/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}/${OS_ARCH}
	mv ${BINARY} ~/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}/${OS_ARCH}

.PHONY: format
format:
	@go mod tidy
	@gofmt -s -w `find . -type f -name '*.go'`


.PHONY: test
test:
	go test -i $(TEST) || exit 1
	echo $(TEST) | xargs -t -n4 go test $(TESTARGS) -timeout=30s -parallel=4

.PHONY: testacc
testacc:
	TF_ACC=1 go test $(TEST) -v $(TESTARGS) -timeout 120m

.PHONY: lint
lint:
	golangci-lint run


