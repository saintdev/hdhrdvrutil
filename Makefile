BIN := hdhrdvrutil
PKG := github.com/saintdev/hdhrdvrutil
VERSION := $(shell git describe --always --long --dirty)
SRCS := $(shell find . -name '*.go' | grep -v /vendor/)

all: $(BIN)

$(BIN): $(SRCS)
	go build -i -v -o $@ -ldflags="-X ${PKG}/cmd.Version=${VERSION}" ${PKG}

clean:
	-$(RM) $(BIN)

.PHONY: all clean
