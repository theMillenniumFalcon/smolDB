# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
BINARY_NAME=smoldb

DIST_FOLDER=dist

all: test build
build: 
		$(GOBUILD) -o $(BINARY_NAME) -v
test: 
		$(GOTEST) -v -cover ./...
clean: 
		$(GOCLEAN)
		rm -rf $(DIST_FOLDER)
build-all:
		mkdir -p $(DIST_FOLDER)
		# [darwin/amd64] - Intel Mac
		CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 $(GOBUILD) -o $(DIST_FOLDER)/$(BINARY_NAME)_darwin_amd64 -v
		# [darwin/arm64] - Apple Silicon Mac
		CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 $(GOBUILD) -o $(DIST_FOLDER)/$(BINARY_NAME)_darwin_arm64 -v
		# [linux/amd64] - 64-bit Linux
		CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(DIST_FOLDER)/$(BINARY_NAME)_linux_amd64 -v
		# [linux/arm64] - 64-bit Linux ARM (e.g. AWS Graviton)
		CGO_ENABLED=0 GOOS=linux GOARCH=arm64 $(GOBUILD) -o $(DIST_FOLDER)/$(BINARY_NAME)_linux_arm64 -v
		# [linux/arm] - 32-bit Linux ARM (e.g. Raspberry Pi)
		CGO_ENABLED=0 GOOS=linux GOARCH=arm $(GOBUILD) -o $(DIST_FOLDER)/$(BINARY_NAME)_linux_arm -v
		# [linux/386] - 32-bit Linux
		CGO_ENABLED=0 GOOS=linux GOARCH=386 $(GOBUILD) -o $(DIST_FOLDER)/$(BINARY_NAME)_linux_386 -v
		# [windows/amd64] - 64-bit Windows
		CGO_ENABLED=0 GOOS=windows GOARCH=amd64 $(GOBUILD) -o $(DIST_FOLDER)/$(BINARY_NAME)_windows_amd64.exe -v
		# [windows/386] - 32-bit Windows
		CGO_ENABLED=0 GOOS=windows GOARCH=386 $(GOBUILD) -o $(DIST_FOLDER)/$(BINARY_NAME)_windows_386.exe -v
		# [windows/arm64] - Windows ARM64 (e.g. Surface Pro X)
		CGO_ENABLED=0 GOOS=windows GOARCH=arm64 $(GOBUILD) -o $(DIST_FOLDER)/$(BINARY_NAME)_windows_arm64.exe -v