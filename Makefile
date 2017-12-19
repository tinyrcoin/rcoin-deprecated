OLDGOPATH := $(GOPATH)/src/
GOPATH := $(GOPATH):$(PWD)
-include config.mk

all: bin bin/rcoind$(EXE)
bin:
	mkdir -p bin
bin/rcoind$(EXE): bin $(wildcard src/rcoin/*.go)
	go build -o bin/rcoind rcoin

deps: $(OLDGOPATH)golang.org/x/crypto/ed25519 $(OLDGOPATH)golang.org/x/crypto/scrypt

$(OLDGOPATH)golang.org/x/crypto/ed25519:
	go get -d golang.org/x/crypto/ed25519
$(OLDGOPATH)golang.org/x/crypto/scrypt:
	go get -d golang.org/x/crypto/scrypt
