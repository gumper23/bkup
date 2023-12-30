# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
BINARY_NAME=bkup
BINARY_WINDOWS=$(BINARY_NAME).exe

all: test tidy build build-windows
build: 
		CGO_ENABLED=0 $(GOBUILD) -o $(BINARY_NAME) -v
tidy:
		$(GOMOD) tidy
test: 
		$(GOTEST) -v ./...
clean: 
		$(GOCLEAN)
		rm -f $(BINARY_NAME)
		rm -f $(BINARY_WINDOWS)
run: build
		./$(BINARY_NAME)

# Cross compilation
build-windows:
		CGO_ENABLED=0 GOOS=windows GOARCH=amd64 $(GOBUILD) -o $(BINARY_WINDOWS) -v
