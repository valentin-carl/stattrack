GO := go
GOCMD := $(GO) build
GOBIN := ./bin
TARGET := stattrack
SRC := ./cmd/main.go

.PHONY: all build clean

all: build

build: $(GOBIN)/$(TARGET)

$(GOBIN)/$(TARGET): $(SRC)
	@echo "building $(TARGET) ..."
	@$(GOCMD) -o $(GOBIN)/$(TARGET) $(SRC)

clean:
	@echo "cleaning up ..."
	@rm -rf $(GOBIN)

