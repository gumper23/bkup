# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
BINARY=bkup
BINARY_WINDOWS=$(BINARY).exe

all: test tidy build build-windows
build: 
		CGO_ENABLED=0 GOOS=linux $(GOBUILD) -o $(BINARY) -v
tidy:
		$(GOMOD) tidy
test: 
		$(GOTEST) -v ./...
clean: 
		$(GOCLEAN)
		rm -f $(BINARY)
		rm -f $(BINARY_WINDOWS)
run: build
		./$(BINARY)

# Cross compilation
build-windows:
		CGO_ENABLED=0 GOOS=windows GOARCH=amd64 $(GOBUILD) -o $(BINARY_WINDOWS) -ldflags -H=windowsgui -v
